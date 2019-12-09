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

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

func checkErr(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
}

func getDevices() []*pluginapi.Device {
	NumOfDevices, err := hlmlGetDeviceCount()
	checkErr(err)

	var devs []*pluginapi.Device

	for i := uint(0); i < NumOfDevices; i++ {
		newDevice, err := hlmlNewDevice(i)
		checkErr(err)

		dev := pluginapi.Device{
			ID:     newDevice.UUID,
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

func watchXIDs(ctx context.Context, devs []*pluginapi.Device, xids chan<- *pluginapi.Device) {
	for {
		select {
		case <-ctx.Done():
			return
		}

		// TODO: check Habanalabs device healthy status
	}
}
