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

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
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

func NewDeviceManager(devType string) *DeviceManager {
	return &DeviceManager{devType: devType}
}

func (dm *DeviceManager) Devices() []*pluginapi.Device {
	NumOfDevices, err := hlmlGetDeviceCount()
	checkErr(err)

	var devs []*pluginapi.Device

	for i := uint(0); i < NumOfDevices; i++ {
		newDevice, err := hlmlNewDevice(i)
		checkErr(err)

		dID := fmt.Sprintf("%x", newDevice.PCI.DeviceID)
		if !strings.HasSuffix(dID, DevID(dm.devType).String()) {
			continue
		}
		log.Printf(
			"device %s,\tserial %s,\tuuid %s",
			strings.ToUpper(dm.devType),
			newDevice.Serial,
			newDevice.UUID,
		)

		dev := pluginapi.Device{
			ID:     newDevice.Serial,
			Health: pluginapi.Healthy,
		}
		devs = append(devs, &dev)
	}

	return devs
}

func check(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
}

func getAllDevices() []*pluginapi.Device {
	NumOfDevices, err := hlmlGetDeviceCount()
	checkErr(err)

	var devs []*pluginapi.Device

	for i := uint(0); i < NumOfDevices; i++ {
		newDevice, err := hlmlNewDevice(i)
		checkErr(err)

		dev := pluginapi.Device{
			ID:     newDevice.Serial,
			Health: pluginapi.Healthy,
		}
		devs = append(devs, &dev)
	}

	return devs
}

func deviceExists(devs []*pluginapi.Device, id string) bool {
	for _, d := range devs {
		if d.ID == id {
			return true
		}
	}
	return false
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
	eventSet := hlmlNewEventSet()
	defer hlmlDeleteEventSet(eventSet)

	for _, d := range devs {
		err := hlmlRegisterEventForDevice(eventSet, HlmlCriticalError, d.ID)
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

		e, err := hlmlWaitForEvent(eventSet, 5000)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}

		if e.Etype != HlmlCriticalError {
			continue
		}

		if e.UUID == nil || len(*e.UUID) == 0 {
			log.Printf("XidCriticalError: Xid=%d, All devices will go unhealthy", e.Etype)
			// All devices are unhealthy
			for _, d := range devs {
				xids <- d
			}
			continue
		}

		for _, d := range devs {
			if d.ID == *e.UUID {
				log.Printf("XidCriticalError: Xid=%d on AIP=%s, the device will go unhealthy", e.Etype, d.ID)
				xids <- d
			}
		}
	}
}
