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
	"log"
	"time"

	hlml "github.com/HabanaAI/gohlml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type DevID string

const (
	GOYA 	DevID = "HL-102"
	GAUDI 	DevID = "HL-205"
)

// ResourceManager interface
type ResourceManager interface {
	Devices() []*pluginapi.Device
}

// DeviceManager string devType: GOYA / GAUDI
type DeviceManager struct {
	devType DevID
}

func NewDeviceManager(devType DevID) *DeviceManager {
	return &DeviceManager{devType: devType}
}

func (dm *DeviceManager) Devices() []*pluginapi.Device {
	NumOfDevices, err := hlml.DeviceCount()
	checkErr(err)

	var devs []*pluginapi.Device

	for i := uint(0); i < NumOfDevices; i++ {

		newDevice, err := hlml.DeviceHandleByIndex(i)
		checkErr(err)

		deviceType, err := newDevice.Name()
		checkErr(err)

		if DevID(deviceType) != dm.devType {
			continue
		}

		serial, err := newDevice.SerialNumber()
		checkErr(err)

		uuid, err := newDevice.UUID()
		checkErr(err)

		log.Printf("%s \t serial: %s \t UUID: %s", deviceType, serial, uuid)
		
		dev := pluginapi.Device{
			ID:     serial,
			Health: pluginapi.Healthy,
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
	eventSet := hlml.HlmlNewEventSet()
	defer hlml.HlmlDeleteEventSet(eventSet)

	for _, d := range devs {
		err := hlml.HlmlRegisterEventForDevice(eventSet, hlml.HlmlCriticalError, d.ID)
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

		e, err := hlml.HlmlWaitForEvent(eventSet, 5000)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}

		if e.Etype != hlml.HlmlCriticalError {
			continue
		}

	}
}
