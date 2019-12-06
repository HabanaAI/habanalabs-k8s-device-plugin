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
	"log"
	"os"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

func getDevices() []*pluginapi.Device {
	/*	n, err := hlml.GetDeviceCount()
		check(err)
	*/
	var devs []*pluginapi.Device
	/*
		for i := uint(0); i < n; i++ {
			d, err := hlml.NewDevice(i)
			check(err)

			dev := pluginapi.Device{
				ID:     d.UUID,
				Health: pluginapi.Healthy,
			}
			devs = append(devs, &dev)
		}
	*/
	return devs
}

func main() {
	/*	log.Println("Loading HLML")
		if err := hlmlInit(); err != nil {
			log.Printf("Failed to initialize HLML: %s.", err)
		}
		defer func() { log.Println("Shutdown of HLML returned:", hlmlShutdown()) }()*/

	log.Println("Fetching devices")

	devList := getDevices()
	if len(devList) == 0 {
		log.Println("No devices found")
		return
	}

	log.Printf("HabanaLabs device list: %v", devList)

	log.Println("Starting FS watcher")
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		log.Println("Failed to created FS watcher")
		os.Exit(1)
	}
	defer watcher.Close()

	log.Println("Starting OS watcher")
	/*sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	restart := true
	var devicePlugin *HabanalabsDevicePlugin*/
}
