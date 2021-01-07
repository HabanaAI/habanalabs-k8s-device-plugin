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
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"

	hlml "github.com/HabanaAI/gohlml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func main() {
	var devicePlugin *HabanalabsDevicePlugin
	var err error
	restart := true

	log.Println("Starting Habana device plugin manager")
	log.Println("Loading HLML")
	if err := hlml.Initialize(); err != nil {
		log.Printf("Failed to initialize HLML: %s", err)
		return
	}
	defer func() { log.Println("Shutdown of HLML returned:", hlml.Shutdown()) }()

	log.Println("Starting FS watcher notifications of filesystem changes")
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		log.Println("Failed to created FS watcher")
		os.Exit(1)
	}
	defer watcher.Close()

	log.Println("Starting OS watcher for system signal notifications")
	sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	devType := flag.String("dev_type", "goya", "Device type which can be either goya (default) or gaudi")
	flag.Parse()

	dev := strings.TrimSpace(*devType)
	switch dev {
	case "goya", "gaudi":
		devicePlugin = NewHabanalabsDevicePlugin(NewDeviceManager(strings.ToUpper(dev)), "habana.ai/"+dev, pluginapi.DevicePluginPath+dev+"_habanalabs.sock")
	default:
		err = fmt.Errorf("Unknown device type: %s", dev)
	}
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

L:
	for {
		if restart {
			devicePlugin.Stop()

			if len(devicePlugin.Devices()) == 0 {
				continue
			}

			if err := devicePlugin.Serve(); err != nil {
				log.Println("Could not contact Kubelet, retrying. Did you enable the device plugin feature gate?")
			} else {
				restart = false
			}
		}

		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				restart = true
			}
		case err := <-watcher.Errors:
			log.Printf("inotify: %s", err)
		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, restarting.")
				restart = true
			default:
				log.Printf("Received signal \"%v\", shutting down", s)
				devicePlugin.Stop()
				break L
			}
		}
	}
}
