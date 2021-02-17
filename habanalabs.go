/*
 * Copyright (c) 2019, HabanaLabs Ltd.  All rights reserved.
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
	"log"
	"strings"
	"time"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	hlml "github.com/HabanaAI/gohlml"
)

type DevID string

const (
	GOYA  DevID = "GOYA"
	GAUDI DevID = "GAUDI"
)

// Device ID 16bit LSB
func (e DevID) String() string {
	switch e {
	case GOYA:
		return "0001"
	case GAUDI:
		return "1000"
	}
	return "N/A"
}

// ResourceManager interface
type ResourceManager interface {
	Devices() []*pluginapi.Device
}

// DeviceManager string devType: GOYA / GAUDI
type DeviceManager struct {
	devType string
}

// NewDeviceManager Init Manager
func NewDeviceManager(devType string) *DeviceManager {
	return &DeviceManager{devType: devType}
}

// Devices Get Habana Device
func (dm *DeviceManager) Devices() []*pluginapi.Device {
	NumOfDevices, err := hlml.DeviceCount()
	checkErr(err)

	var devs []*pluginapi.Device

	log.Println("Finding devices...")
	for i := uint(0); i < NumOfDevices; i++ {
		newDevice, err := hlml.DeviceHandleByIndex(i)
		checkErr(err)

		pciID, err := newDevice.PCIID()
		checkErr(err)

		serial, err := newDevice.SerialNumber()
		checkErr(err)

		uuid, err := newDevice.UUID()
		checkErr(err)

		pciBusID, _ := newDevice.PCIBusID()

		dID := fmt.Sprintf("%x", pciID)

		if !strings.HasSuffix(dID, DevID(dm.devType).String()) {
			log.Printf("Not correct device type")
			continue
		}

		log.Printf(
			"device: %s,\tserial: %s,\tuuid: %s",
			strings.ToUpper(dm.devType),
			serial,
			uuid,
		)

		log.Printf("pci id: %s\t pci bus id: %s",
			dID,
			pciBusID,
		)

		dev := pluginapi.Device{
			ID:     serial,
			Health: pluginapi.Healthy,
		}

		cpuAffinity, err := newDevice.NumaNode()
		checkErr(err)

		if cpuAffinity != nil {
			log.Printf("cpu affinity: %d", *cpuAffinity)
			dev.Topology = &pluginapi.TopologyInfo{
				Nodes: []*pluginapi.NUMANode{{ID: int64(*cpuAffinity)},
				},
			}
		}
		devs = append(devs, &dev)
	}

	return devs
}

func checkErr(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
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
			log.Printf("Failed to register critical events for %s, error %s. Marking it unhealthy", d.ID, err)

			xids <- d
			continue
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		e, err := hlml.WaitForEvent(eventSet, 5000)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}

		if e.Etype != hlml.HlmlCriticalError {
			continue
		}

		dev, err := hlml.DeviceHandleBySerial(e.Serial)
		uuid, err := dev.UUID()

		if err != nil || len(uuid) == 0 {
			log.Printf("XidCriticalError: Xid=%d, All devices will go unhealthy", e.Etype)
			// All devices are unhealthy
			for _, d := range devs {
				xids <- d
			}
			continue
		}

		for _, d := range devs {
			if d.ID == uuid {
				log.Printf("XidCriticalError: Xid=%d on AIP=%s, the device will go unhealthy", e.Etype, d.ID)
				xids <- d
			}
		}

	}
}
