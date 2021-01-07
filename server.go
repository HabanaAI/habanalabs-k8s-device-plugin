/*
 * Copyright (c) 2021, HabanaLabs Ltd.  All rights reserved.
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
	"fmt"
	hlml "github.com/HabanaAI/gohlml"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"google.golang.org/grpc"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// HabanalabsDevicePlugin implements the Kubernetes device plugin API
type HabanalabsDevicePlugin struct {
	ResourceManager
	resourceName string
	socket       string

	devs   []*pluginapi.Device
	stop   chan interface{}
	health chan *pluginapi.Device
	server *grpc.Server
}

// NewHabanalabsDevicePlugin returns an initialized HabanalabsDevicePlugin
func NewHabanalabsDevicePlugin(resourceManager ResourceManager, resourceName string, socket string) *HabanalabsDevicePlugin {
	return &HabanalabsDevicePlugin{
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
		PreStartRequired:	true,
		GetPreferredAllocationAvailable: false,
	}, nil
}

func (m *HabanalabsDevicePlugin) GetPreferredAllocation(ctx context.Context, rqt *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	response := &pluginapi.PreferredAllocationResponse{}
	return response, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start starts the gRPC server of the device plugin
func (m *HabanalabsDevicePlugin) Start() error {
	if err := m.cleanup(); err != nil {
		return err
	}

	//  initialize Devices
	m.devs = m.Devices()

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(m.server, m)

	go m.server.Serve(sock)

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

	log.Printf("Stopping to serve '%s' on %s", m.resourceName, m.socket)
	m.server.Stop()
	m.server = nil
	close(m.stop)

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
	s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})

	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			d.Health = pluginapi.Unhealthy
			log.Printf("'%s' device %s is unhealthy", m.resourceName, d.ID)
			s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
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
		paths := make([]string, 0, len(req.DevicesIDs))
		uuids := make([]string, 0, len(req.DevicesIDs))

		for _, id := range req.DevicesIDs {
			device := getDevice(devs, id)
			if device == nil {
				return nil, fmt.Errorf("Invalid request for '%s': device unknown: %s", m.resourceName, id)
			}
			log.Printf("device == %s", device)

			deviceHandle, err := hlml.DeviceHandleByUUID(id)
			checkErr(err)

			minor, err := deviceHandle.MinorNumber()
			checkErr(err)

			path := fmt.Sprintf("/dev/hl%d", minor)
			paths = append(paths, path)
			uuids = append(uuids, id)

			log.Printf("path == %s", path)

			ds := &pluginapi.DeviceSpec{
				ContainerPath: path,
				HostPath:      path,
				Permissions:   "rw",
			}
			devicesList = append(devicesList, ds)

			path = fmt.Sprintf("/dev/hl_controlD%d", minor)

			log.Printf("path == %s", path)

			ds = &pluginapi.DeviceSpec{
				ContainerPath: path,
				HostPath:      path,
				Permissions:   "rw",
			}
			devicesList = append(devicesList, ds)
		}

		response.ContainerResponses = append(response.ContainerResponses, &pluginapi.ContainerAllocateResponse{
			Devices: devicesList,
			Envs: map[string]string{
				"HL_VISIBLE_DEVICES":      strings.Join(paths[:], ","),
				"HL_VISIBLE_DEVICES_UUID": strings.Join(uuids[:], ","),
			},
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
		log.Printf("Could not start device plugin: %s", err)
		return err
	}
	log.Println("Starting to serve on", m.socket)

	err = m.Register()
	if err != nil {
		log.Printf("Could not register device plugin: %s", err)
		m.Stop()
		return err
	}
	log.Println("Registered device plugin with Kubelet")

	return nil
}
