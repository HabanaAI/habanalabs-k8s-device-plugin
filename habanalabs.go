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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	hlml "github.com/HabanaAI/gohlml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

const (
	// Base for sysfs accels
	sysClassAccel    = "/sys/class/accel/accel"
	gaudinetFilePath = "/etc/habanalabs/gaudinet.json"

	gaudiCdiNamespacePrefix = "intel.com/gaudi"
	cdiDynamicDir           = "/var/run/cdi"
	cdiVersion              = "0.5.0"
)

// ResourceManager interface
type ResourceManager interface {
	Devices() ([]*pluginapi.Device, error)
}

// DeviceManager string devType: GOYA / GAUDI
type DeviceManager struct {
	log     *slog.Logger
	devType string
}

type deviceEntity struct {
	devicePaths []string
	uuid        string
	moduleId    uint
	deviceId    uint
}

// NewDeviceManager Init Manager
func NewDeviceManager(log *slog.Logger, devType string) *DeviceManager {
	return &DeviceManager{log: log, devType: devType}
}

// Devices Get Habana Device
func (dm *DeviceManager) Devices() ([]*pluginapi.Device, error) {
	NumOfDevices, err := hlml.DeviceCount()
	if err != nil {
		return nil, err
	}

	var devs []*pluginapi.Device

	dm.log.Info("Discovering devices...")
	for i := uint(0); i < NumOfDevices; i++ {
		newDevice, err := hlml.DeviceHandleByIndex(i)
		if err != nil {
			return nil, err
		}

		pciID, err := newDevice.PCIID()
		if err != nil {
			return nil, err
		}

		serial, err := newDevice.SerialNumber()
		if err != nil {
			return nil, err
		}

		uuid, err := newDevice.UUID()
		if err != nil {
			return nil, err
		}

		pciBusID, _ := newDevice.PCIBusID()
		dID := fmt.Sprintf("%x", pciID)
		dm.log.Info(
			"Device found",
			"device", strings.ToUpper(dm.devType),
			"serial", serial,
			"uuid", uuid,
			"id", dID,
			"pci_bus_id", pciBusID,
		)

		dev := pluginapi.Device{
			ID:     serial,
			Health: pluginapi.Healthy,
		}

		cpuAffinity, err := newDevice.NumaNode()
		if err != nil {
			return nil, err
		}

		if cpuAffinity != nil {
			dm.log.Info("Device cpu affinity", "id", dID, "cpu_affinity", *cpuAffinity)
			dev.Topology = &pluginapi.TopologyInfo{
				Nodes: []*pluginapi.NUMANode{{ID: int64(*cpuAffinity)}},
			}
		}
		devs = append(devs, &dev)
	}

	return devs, nil
}

func getDevice(devs []*pluginapi.Device, id string) *pluginapi.Device {
	for _, d := range devs {
		if d.ID == id {
			return d
		}
	}
	return nil
}

func watchXIDs(ctx context.Context, devs []*pluginapi.Device, xids chan<- *pluginapi.Device) {
	eventSet := hlml.NewEventSet()
	defer hlml.DeleteEventSet(eventSet)

	for _, d := range devs {
		err := hlml.RegisterEventForDevice(eventSet, hlml.HlmlCriticalError, d.ID)
		if err != nil {
			slog.Error("Failed registering critial event for device. Marking it unhealthy", "device_id", d.ID, "error", err)
			xids <- d
			continue
		}
	}

	// TODO: provide as flag
	healthCheckInterval := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-healthCheckInterval.C:
			e, err := hlml.WaitForEvent(eventSet, 1000)
			if err != nil {
				slog.Error("hlml WaitForEvent failed", "errror", err.Error())
				time.Sleep(2 * time.Second)
				continue
			}

			if e.Etype != hlml.HlmlCriticalError {
				continue
			}

			dev, err := hlml.DeviceHandleBySerial(e.Serial)
			if err != nil {
				slog.Error("XidCriticalError: All devices will go unhealthy", "xid", e.Etype)
				// All devices are unhealthy
				for _, d := range devs {
					xids <- d
				}
				continue
			}

			uuid, err := dev.UUID()
			if err != nil || len(uuid) == 0 {
				slog.Error("XidCriticalError: All devices will go unhealthy", "xid", e.Etype)
				// All devices are unhealthy
				for _, d := range devs {
					xids <- d
				}
				continue
			}

			for _, d := range devs {
				if d.ID == uuid {
					slog.Error("XidCriticalError: the device will go unhealthy", "xid", e.Etype, "aip", d.ID)
					xids <- d
				}
			}
		}
	}
}

// UverbsForAccelerators returns list of infiniband char devices attached to each
// accelerator provided.
// Copied from habana-container-runtime's discover.go
func (m *HabanalabsDevicePlugin) uverbsForAccelerator(deviceID uint) string {
	content, err := os.ReadFile(fmt.Sprintf("%s%d/device/pci_addr", sysClassAccel, deviceID))
	if err != nil {
		return ""
	}

	pciAddr := strings.TrimSpace(string(content))

	// Check at the device folder for the uverb dir
	deviceUverbPath := fmt.Sprintf("/sys/bus/pci/devices/%s/infiniband_verbs", pciAddr)

	uverbs, err := os.ReadDir(deviceUverbPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return ""
	}
	if len(uverbs) == 0 {
		return ""
	}

	return fmt.Sprintf("/dev/infiniband/%s", uverbs[0].Name())
}

// Writes a new CDI spec into the dynamic CDI dir and
// returns the names of the CDI devices that were written
// to the spec. The names can then be given to DP's Allocate response.
func (m *HabanalabsDevicePlugin) writeCdiSpecAndReturnDevices(devEntities []deviceEntity, envMap map[string]string) []*pluginapi.CDIDevice {
	devices := []*pluginapi.CDIDevice{}

	// create a random name for the device and spec file
	randPrefix := strconv.FormatInt(time.Now().UnixMilli(), 10)

	baseSpec := cdispec.Spec{
		Version: cdiVersion,
		Kind:    gaudiCdiNamespacePrefix,
		ContainerEdits: cdispec.ContainerEdits{
			Env: []string{},
			Hooks: []*cdispec.Hook{
				{
					HookName: "createRuntime",
					Path:     "/usr/bin/habana-container-hook",
					Args:     []string{"habana-container-hook", "createRuntime"},
				},
			},
		},
		Devices: make([]cdispec.Device, len(devEntities)),
	}

	// Add gaudinet mount
	if _, err := os.Stat(gaudinetFilePath); err == nil {
		baseSpec.ContainerEdits.Mounts = []*cdispec.Mount{
			{
				HostPath:      gaudinetFilePath,
				ContainerPath: gaudinetFilePath,
				Options:       []string{"bind"},
			},
		}
	}

	for k, v := range envMap {
		baseSpec.ContainerEdits.Env = append(baseSpec.ContainerEdits.Env, k+"="+v)
	}

	for i, dev := range devEntities {
		name := fmt.Sprintf("%s-gaudi%d", randPrefix, dev.deviceId)

		devices = append(devices, &pluginapi.CDIDevice{
			Name: fmt.Sprintf("%s=%s", gaudiCdiNamespacePrefix, name),
		})

		baseSpec.Devices[i].Name = name
		baseSpec.Devices[i].ContainerEdits = cdispec.ContainerEdits{
			DeviceNodes: make([]*cdispec.DeviceNode, len(dev.devicePaths)),
		}

		for j, path := range dev.devicePaths {
			baseSpec.Devices[i].ContainerEdits.DeviceNodes[j] = &cdispec.DeviceNode{
				Path:     path,
				HostPath: path,
				Type:     "c",
			}
		}
	}

	cdiJson, err := json.MarshalIndent(baseSpec, "", "  ")
	if err != nil {
		m.log.Error(err.Error())

		return []*pluginapi.CDIDevice{}
	}

	cdiSpecFilePath := filepath.Join(cdiDynamicDir, fmt.Sprintf("dp-gaudi-%s.json", randPrefix))

	err = os.WriteFile(cdiSpecFilePath, cdiJson, 0666)
	if err != nil {
		m.log.Error(err.Error())

		return []*pluginapi.CDIDevice{}
	}

	return devices
}

// Creates env variables map from the given Gaudi devices.
func envsFromDevices(devEntities []deviceEntity, subsetOfDevices bool) map[string]string {
	visibleIds := []string{}
	visiblePaths := []string{}
	visibleUuids := []string{}
	visibleModules := []string{}

	for _, dev := range devEntities {
		visibleIds = append(visibleIds, strconv.FormatUint(uint64(dev.deviceId), 10))
		visiblePaths = append(visiblePaths, dev.devicePaths[0])
		visibleUuids = append(visibleUuids, dev.uuid)
		visibleModules = append(visibleModules, strconv.FormatUint(uint64(dev.moduleId), 10))
	}

	envMap := map[string]string{
		"HABANA_VISIBLE_DEVICES":  strings.Join(visibleIds, ","),
		"HL_VISIBLE_DEVICES":      strings.Join(visiblePaths, ","),
		"HL_VISIBLE_DEVICES_UUID": strings.Join(visibleUuids, ","),
	}

	if subsetOfDevices {
		envMap["HABANA_VISIBLE_MODULES"] = strings.Join(visibleModules, ",")
	}

	return envMap
}

// Creates DeviceSpecs to be returned with the Allocate response.
func (m *HabanalabsDevicePlugin) createDeviceSpecs(devEntities []deviceEntity) []*pluginapi.DeviceSpec {
	specs := []*pluginapi.DeviceSpec{}

	for _, dev := range devEntities {
		for _, path := range dev.devicePaths {
			specs = append(specs, &pluginapi.DeviceSpec{
				HostPath:      path,
				ContainerPath: path,
				Permissions:   "rw",
			})
		}
	}

	return specs
}

// Remove DP Gaudi CDI specs
func (m *HabanalabsDevicePlugin) cdiSpecCleanup() {
	files, err := filepath.Glob(cdiDynamicDir + "/dp-gaudi-*.json")
	if err != nil {
		m.log.Warn("failed to list CDI spec files", "err", err)

		return
	}

	cleanupCount := 0
	for _, spec := range files {
		if err = os.Remove(spec); err != nil {
			m.log.Warn("failed to remove cdi spec", "err", err)
		} else {
			cleanupCount++
		}
	}

	m.log.Info("CDI spec cleanup", "count", cleanupCount)
}
