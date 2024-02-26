/*
 * Copyright (c) 2022, HabanaLabs Ltd.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"google.golang.org/grpc"

	hlml "github.com/HabanaAI/gohlml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// HabanalabsDevicePlugin implements the Kubernetes device plugin API
type HabanalabsDevicePlugin struct {
	ResourceManager
	log          *slog.Logger
	stop         chan interface{}
	health       chan *pluginapi.Device
	server       *grpc.Server
	resourceName string
	socket       string
	devs         []*pluginapi.Device
}

// GetPreferredAllocation returns a preferred set of devices to allocate
// from a list of available ones. The resulting preferred allocation is not
// guaranteed to be the allocation ultimately performed by the
// devicemanager. It is only designed to help the devicemanager make a more
// informed allocation decision when possible.
// NOT Implemented
func (m *HabanalabsDevicePlugin) GetPreferredAllocation(ctx context.Context, request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, errors.New("GetPreferredAllocation should not be called as this device plugin doesn't implement it")
}

// NewHabanalabsDevicePlugin returns an initialized HabanalabsDevicePlugin.
func NewHabanalabsDevicePlugin(log *slog.Logger, resourceManager ResourceManager, resourceName string, socket string) *HabanalabsDevicePlugin {
	return &HabanalabsDevicePlugin{
		log:             log,
		ResourceManager: resourceManager,
		resourceName:    resourceName,
		socket:          socket,

		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),

		// will be initialized on every server restart.
		devs: nil,
	}
}

// GetDevicePluginOptions returns the device plugin options.
func (m *HabanalabsDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		GetPreferredAllocationAvailable: false, // Indicate to kubelet we don't have an implementation.
	}, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c, err := grpc.DialContext(ctx, unixSocketPath,
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return net.DialTimeout("unix", s, timeout)
		}),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start starts the gRPC server of the device plugin
func (m *HabanalabsDevicePlugin) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	if m.stop == nil {
		m.stop = make(chan interface{})
	}

	//  initialize Devices
	m.devs, err = m.Devices()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	// First start serving the gRPC connection before registering.
	// It is required since kubernetes 1.26. Change is backward compatible.
	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(m.server, m)

	// Ignore error returns since the next block will fail if Serve fails.
	go func() { _ = m.server.Serve(sock) }()

	// Wait for server to start by launching a blocking connection
	conn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	go m.healthcheck()

	return nil
}

// Stop gRPC server
func (m *HabanalabsDevicePlugin) Stop() error {
	if m.server == nil {
		return nil
	}

	m.log.Info("Stoppping device plugin", "resource_name", m.resourceName, "socket", m.socket)
	m.server.Stop()
	m.server = nil
	close(m.stop)
	m.stop = nil

	return m.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (m *HabanalabsDevicePlugin) Register() error {
	conn, err := dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: m.resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (m *HabanalabsDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	err := s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
	if err != nil {
		return err
	}

	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			d.Health = pluginapi.Unhealthy
			m.log.Info("Device is unhealthy", "resource", m.resourceName, "id", d.ID)
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs}); err != nil {
				m.log.Error("Failed sending ListAndWatch to kubelet", "error", err)
			}
		}
	}
}

func (m *HabanalabsDevicePlugin) unhealthy(dev *pluginapi.Device) {
	m.health <- dev
}

// Allocate which return list of devices.
func (m *HabanalabsDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	devs := m.devs
	response := pluginapi.AllocateResponse{ContainerResponses: []*pluginapi.ContainerAllocateResponse{}}
	for _, req := range reqs.ContainerRequests {
		var devicesList []*pluginapi.DeviceSpec
		netConfig := make([]string, 0, len(req.DevicesIDs))
		paths := make([]string, 0, len(req.DevicesIDs))
		uuids := make([]string, 0, len(req.DevicesIDs))
		visibleModule := make([]string, 0, len(req.DevicesIDs))

		for _, id := range req.DevicesIDs {
			device := getDevice(devs, id)
			if device == nil {
				return nil, fmt.Errorf("invalid request for %q: device unknown: %s", m.resourceName, id)
			}
			m.log.Info("Preparing device for registration", "device", device)

			m.log.Info("Getting device handle from hlml")
			deviceHandle, err := hlml.DeviceHandleBySerial(id)
			if err != nil {
				m.log.Error(err.Error())
				return nil, err
			}

			m.log.Info("Getting device minor number")
			minor, err := deviceHandle.MinorNumber()
			if err != nil {
				m.log.Error(err.Error())
				return nil, err
			}

			m.log.Info("Getting device module id")
			moduleID, err := deviceHandle.ModuleID()
			if err != nil {
				m.log.Error(err.Error())
				return nil, err
			}

			path := fmt.Sprintf("/dev/accel/accel%d", minor)
			paths = append(paths, path)
			uuids = append(uuids, id)
			netConfig = append(netConfig, fmt.Sprintf("%d", minor))
			visibleModule = append(visibleModule, fmt.Sprintf("%d", moduleID))

			ds := &pluginapi.DeviceSpec{
				ContainerPath: path,
				HostPath:      path,
				Permissions:   "rw",
			}
			devicesList = append(devicesList, ds)
			path = fmt.Sprintf("/dev/accel/accel_controlD%d", minor)

			ds = &pluginapi.DeviceSpec{
				ContainerPath: path,
				HostPath:      path,
				Permissions:   "rw",
			}
			devicesList = append(devicesList, ds)
		}

		envMap := map[string]string{
			"HABANA_VISIBLE_DEVICES":  strings.Join(netConfig, ","),
			"HL_VISIBLE_DEVICES":      strings.Join(paths, ","),
			"HL_VISIBLE_DEVICES_UUID": strings.Join(uuids, ","),
		}

		if len(req.DevicesIDs) < len(m.devs) {
			envMap["HABANA_VISIBLE_MODULES"] = strings.Join(visibleModule, ",")
		}

		response.ContainerResponses = append(response.ContainerResponses, &pluginapi.ContainerAllocateResponse{
			Devices: devicesList,
			Envs:    envMap,
		})
	}

	return &response, nil
}

// PreStartContainer performs actions before the container start
func (m *HabanalabsDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (m *HabanalabsDevicePlugin) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (m *HabanalabsDevicePlugin) healthcheck() {
	ctx, cancel := context.WithCancel(context.Background())

	xids := make(chan *pluginapi.Device)
	go watchXIDs(ctx, m.devs, xids)

	for {
		select {
		case <-m.stop:
			cancel()
			return
		case dev := <-xids:
			m.unhealthy(dev)
		}
	}
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *HabanalabsDevicePlugin) Serve() error {
	err := m.Start()
	if err != nil {
		return fmt.Errorf("could not start device plugln: %w", err)
	}
	m.log.Info("Starting to serve", "socket", m.socket)

	err = m.Register()
	if err != nil {
		_ = m.Stop()
		return fmt.Errorf("could not register device plugin: %w", err)
	}
	m.log.Info("Registered device plugin with Kubelet")

	return nil
}
