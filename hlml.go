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
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	szUUID = 256
	// HlmlCriticalError indicates a critical error in the device
	HlmlCriticalError = C.HLML_EVENT_CRITICAL_ERR
)

type handle struct{ dev C.hlml_device_t }

// EventSet is a cast of the C type of the hlml event set
type EventSet struct{ set C.hlml_event_set_t }

// Event contains uuid and event type
type Event struct {
	UUID  *string
	Etype uint64
}

// PCIInfo contains the PCI properties of the device
type PCIInfo struct {
	BusID    string
	DeviceID uint
}

// Device contains the information about the device
type Device struct {
	handle
	pluginapi.Device

	Path 	string
	UUID 	string
	Serial	string
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
		fallthrough
	case C.HLML_ERROR_TIMEOUT:
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

	return fmt.Errorf("Invalid HLML error return code %d", ret)
}

func hlmlInit() error {
	return errorString(C.hlml_init())
}

func hlmlInitWithLogs() error {
	return errorString(C.hlml_init_with_flags(0x6))
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

func hlmlDeviceGetHandleByUUID(uuid string) (handle, error) {
	var dev C.hlml_device_t

	cstr := C.CString(uuid)
	defer C.free(unsafe.Pointer(cstr))

	rc := C.hlml_device_get_handle_by_UUID(cstr, &dev)
	return handle{dev}, errorString(rc)
}

// consider replacing with a direct call to the C API if it is added later
func hlmlDeviceGetHandleBySerial(serial string) (*handle, error) {
	numDevices, err := hlmlGetDeviceCount()
	checkErr(err)

	for i := uint(0); i < numDevices; i++ {
		handle, err := hlmlDeviceGetHandleByIndex(i)
		checkErr(err)

		currentSerial, err := hlmlDeviceGetSerial(handle)
		checkErr(err)

		if *currentSerial == serial { return &handle, nil }
	}

	return nil, errors.New("could not find device with serial number")
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

func hlmlDeviceGetSerial(h handle) (*string, error) {
	var serial [szUUID]C.char

	rc := C.hlml_device_get_serial(h.dev, &serial[0], szUUID)
	return stringPtr(&serial[0]), errorString(rc)
}

func hlmlDeviceGetPciInfo(h handle) (*string, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(h.dev, &pci)
	return stringPtr(&pci.bus_id[0]), errorString(rc)
}

func hlmlDeviceGetDeviceID(h handle) (*uint, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(h.dev, &pci)
	return uintPtr(pci.pci_device_id), errorString(rc)
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
	deviceID, err := hlmlDeviceGetDeviceID(deviceHandle)
	checkErr(err)
	minor, err := hlmlDeviceGetMinorNumber(deviceHandle)
	checkErr(err)
	uuid, err := hlmlDeviceGetUUID(deviceHandle)
	checkErr(err)
	serial, err := hlmlDeviceGetSerial(deviceHandle)
	checkErr(err)

	path := fmt.Sprintf("/dev/hl%d", *minor)

	device = &Device{
		handle: deviceHandle,
		UUID:   *uuid,
		Serial:	*serial,
		Path:   path,
		PCI: PCIInfo{
			BusID:    *busid,
			DeviceID: *deviceID,
		},
	}
	return
}

func hlmlNewEventSet() EventSet {
	var set C.hlml_event_set_t
	C.hlml_event_set_create(&set)

	return EventSet{set}
}

func hlmlRegisterEventForDevice(es EventSet, event int, serial string) error {

	deviceHandle, err := hlmlDeviceGetHandleBySerial(serial)

	if err != nil {
		return fmt.Errorf("hlml: device not found")
	}

	r := C.hlml_device_register_events(deviceHandle.dev, C.ulonglong(event), es.set)
	if r != C.HLML_SUCCESS {
		return errorString(r)
	}

	return nil
}

func hlmlDeleteEventSet(es EventSet) {
	C.hlml_event_set_free(es.set)
}

func hlmlWaitForEvent(es EventSet, timeout uint) (Event, error) {
	var data C.hlml_event_data_t

	r := C.hlml_event_set_wait(es.set, &data, C.uint(timeout))
	uuid, _ := hlmlDeviceGetUUID(handle{data.device})

	return Event{
			UUID:  uuid,
			Etype: uint64(data.event_type),
		},
		errorString(r)
}
