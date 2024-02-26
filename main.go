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
	"fmt"
	"log/slog"
	"os"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"

	hlml "github.com/HabanaAI/gohlml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// build is overridden with an actual version in the build process.
var build = "develop"

func main() {
	log := initLogger()
	if err := run(log); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func initLogger() *slog.Logger {
	lvl := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == slog.LevelDebug.String() {
		lvl = slog.LevelDebug
	}
	attrs := []slog.Attr{
		slog.String("service", "habana-device-plugin"),
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}).WithAttrs(attrs)
	return slog.New(h)
}

func run(log *slog.Logger) error {
	restart := true
	log.Info("Started Habana device plugin manager", "version", build)

	log.Info("Initializing HLML...")
	if err := hlml.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize HLML: %w", err)
	}
	defer func() {
		log.Info("Shutting down hlml")
		err := hlml.Shutdown()
		if err != nil {
			log.Error(err.Error())
		}
	}()

	log.Info("Starting FS watcher...")
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		return fmt.Errorf("failed to create FS watcher: %w", err)
	}
	defer watcher.Close()

	log.Info("Starting OS watcher...")
	sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	dev, err := hlml.GetDeviceTypeName()
	if err != nil {
		return fmt.Errorf("failed detecting Habana's devices on the system: %w", err)
	}

	devicePlugin := NewHabanalabsDevicePlugin(
		log,
		NewDeviceManager(log, strings.ToUpper(dev)),
		"habana.ai/"+dev,
		pluginapi.DevicePluginPath+dev+"_habanalabs.sock",
	)

L:
	for {
		if restart {
			err = devicePlugin.Stop()
			if err != nil {
				log.Warn("Failed stopping device plugin gracefully", "error", err)
			}

			numDevices, err := hlml.DeviceCount()
			if err != nil {
				return fmt.Errorf("failed getting number of devices: %w", err)
			}

			if numDevices == 0 {
				continue
			}

			if err := devicePlugin.Serve(); err != nil {
				log.Error(err.Error())
				return fmt.Errorf("could not contact Kubelet, retrying. Did you enable the device plugin feature gate?")
			}
			restart = false
		}

		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Warn("Kubelet restart detected, restarting device plugin.")
				restart = true
			}
		case err := <-watcher.Errors:
			log.Error("Watcher error received", "error", err)
		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Info("Received SIGHUP, restarting.")
				restart = true
			default:
				log.Info("Received OS signal. Shutting down", "signal", s)
				if err := devicePlugin.Stop(); err != nil {
					log.Error("Failed stopping device plugin gracefully", "error", err)
				}
				break L
			}
		}
	}
	return nil
}
