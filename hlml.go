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

// #cgo LDFLAGS: "./hlml/libhlml.a" -ldl -Wl,--unresolved-symbols=ignore-in-object-files
// #include "hlml/hlml.h"
import "C"
import "fmt"

const (
	szUUID = 256

	// HlmlCriticalError indicates a critical error in the device
	HlmlCriticalError = C.HLML_EVENT_CRITICAL_ERR
)

type handle struct{ dev C.hlml_device_t }

// PCIInfo contains the PCI properties of the device
type PCIInfo struct {
	BusID string
}

// Device contains the information about the device
type Device struct {
	handle

	Path string
	UUID string
	PCI  PCIInfo
}

func uintPtr(c C.uint) *uint {
	i := uint(c)
	return &i
}

func uint64Ptr(c C.ulonglong) *uint64 {
	i := uint64(c)
	return &i
}

func stringPtr(c *C.char) *string {
	s := C.GoString(c)
	return &s
}

func errorString(ret C.hlml_return_t) error {
	switch ret {
	case C.HLML_SUCCESS:
		return nil
	case C.HLML_ERROR_UNINITIALIZED:
		return fmt.Errorf("HLML not initialized")
	case C.HLML_ERROR_INVALID_ARGUMENT:
		return fmt.Errorf("Invalid argument")
	case C.HLML_ERROR_NOT_SUPPORTED:
		return fmt.Errorf("Not supported")
	case C.HLML_ERROR_ALREADY_INITIALIZED:
		return fmt.Errorf("HLML already initialized")
	case C.HLML_ERROR_NOT_FOUND:
		return fmt.Errorf("Not found")
	case C.HLML_ERROR_INSUFFICIENT_SIZE:
		return fmt.Errorf("Insufficient size")
	case C.HLML_ERROR_DRIVER_NOT_LOADED:
		return fmt.Errorf("Driver not loaded")
	case C.HLML_ERROR_AIP_IS_LOST:
		return fmt.Errorf("AIP is lost")
	case C.HLML_ERROR_MEMORY:
		return fmt.Errorf("Memory error")
	case C.HLML_ERROR_NO_DATA:
		return fmt.Errorf("No data")
	case C.HLML_ERROR_UNKNOWN:
		return fmt.Errorf("Unknown error")
	}

	return fmt.Errorf("Invalid error return code")
}

func hlmlInit() error {
	return errorString(C.hlml_init())
}

func hlmlShutdown() error {
	return errorString(C.hlml_shutdown())
}

func hlmlGetDeviceCount() (uint, error) {
	var NumOfDevices C.uint

	rc := C.hlml_device_get_count(&NumOfDevices)
	return uint(NumOfDevices), errorString(rc)
}

func hlmlDeviceGetHandleByIndex(idx uint) (handle, error) {
	var dev C.hlml_device_t

	rc := C.hlml_device_get_handle_by_index(C.uint(idx), &dev)
	return handle{dev}, errorString(rc)
}

func hlmlDeviceGetMinorNumber(h handle) (*uint, error) {
	var minor C.uint

	rc := C.hlml_device_get_minor_number(h.dev, &minor)
	return uintPtr(minor), errorString(rc)
}

func hlmlDeviceGetUUID(h handle) (*string, error) {
	var uuid [szUUID]C.char

	rc := C.hlml_device_get_uuid(h.dev, &uuid[0], szUUID)
	return stringPtr(&uuid[0]), errorString(rc)
}

func hlmlDeviceGetPciInfo(h handle) (*string, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(h.dev, &pci)
	return stringPtr(&pci.busId[0]), errorString(rc)
}

func hlmlNewDevice(idx uint) (device *Device, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	deviceHandle, err := hlmlDeviceGetHandleByIndex(idx)
	checkErr(err)
	busid, err := hlmlDeviceGetPciInfo(deviceHandle)
	checkErr(err)
	minor, err := hlmlDeviceGetMinorNumber(deviceHandle)
	checkErr(err)
	uuid, err := hlmlDeviceGetUUID(deviceHandle)
	checkErr(err)

	path := fmt.Sprintf("/dev/hl%d", *minor)

	device = &Device{
		handle: deviceHandle,
		UUID:   *uuid,
		Path:   path,
		PCI: PCIInfo{
			BusID: *busid,
		},
	}
	return
}
