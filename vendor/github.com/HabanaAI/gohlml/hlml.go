/*
 * Copyright (c) 2020, HabanaLabs Ltd.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the Lic
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gohlml

// #cgo CFLAGS: -I${SRCDIR}/hlml/
// #cgo LDFLAGS: "${SRCDIR}/hlml/libhlml.a" -ldl -Wl,--unresolved-symbols=ignore-in-object-files
// #include "hlml.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"io/ioutil"
	"unsafe"
)

const (
	szUUID = 256
	// HlmlCriticalError indicates a critical error in the device
	HlmlCriticalError = C.HLML_EVENT_CRITICAL_ERR
	// HLDriverPath indicates on habana device dir
	HLDriverPath = "/sys/class/habanalabs"
	// HLModulePath indicates on habana module dir
	HLModulePath = "/sys/module/habanalabs"
)

// Device struct maps to C HLML structure
type Device struct{ dev C.hlml_device_t }

// EventSet is a cast of the C type of the hlml event set
type EventSet struct{ set C.hlml_event_set_t }

// Event contains uuid and event type
type Event struct {
	UUID  string
	Etype uint64
}

// PCIInfo contains the PCI properties of the device
type PCIInfo struct {
	BusID    string
	DeviceID uint
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

// Initialize initializes the HLML library
func Initialize() error {
	return errorString(C.hlml_init())
}

// InitWithLogs initializes the HLML library with logging on
func InitWithLogs() error {
	return errorString(C.hlml_init_with_flags(0x6))
}

// Shutdown shutdowns the HLML library
func Shutdown() error {
	return errorString(C.hlml_shutdown())
}

// DeviceCount gets number of Habana devices in the system
func DeviceCount() (uint, error) {
	var NumOfDevices C.uint

	rc := C.hlml_device_get_count(&NumOfDevices)
	return uint(NumOfDevices), errorString(rc)
}

// DeviceHandleByIndex gets a handle to a particular device by index
func DeviceHandleByIndex(idx uint) (Device, error) {
	var dev C.hlml_device_t

	rc := C.hlml_device_get_handle_by_index(C.uint(idx), &dev)
	return Device{dev}, errorString(rc)
}

// DeviceHandleByUUID gets a handle to a particular device by UUIC
func DeviceHandleByUUID(uuid string) (Device, error) {
	var dev C.hlml_device_t

	cstr := C.CString(uuid)
	defer C.free(unsafe.Pointer(cstr))

	rc := C.hlml_device_get_handle_by_UUID(cstr, &dev)
	return Device{dev}, errorString(rc)
}

// MinorNumber returns Minor number:
// minor
func (d Device) MinorNumber() (uint, error) {
	var minor C.uint

	rc := C.hlml_device_get_minor_number(d.dev, &minor)
	return uint(minor), errorString(rc)
}

// Name returns Device Name
// name
func (d Device) Name() (string, error) {
	var name [szUUID]C.char

	rc := C.hlml_device_get_name(d.dev, &name[0], szUUID)
	return C.GoString(&name[0]), errorString(rc)
}

// UUID returns the unique id for a given device
func (d Device) UUID() (string, error) {
	var uuid [szUUID]C.char

	rc := C.hlml_device_get_uuid(d.dev, &uuid[0], szUUID)
	return C.GoString(&uuid[0]), errorString(rc)
}

// PCIDomain returns the PCI domain for a given device
func (d Device) PCIDomain() (uint, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(d.dev, &pci)
	return uint(pci.domain), errorString(rc)
}

// PCIBus returns the PCI bus info for a given device
func (d Device) PCIBus() (uint, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(d.dev, &pci)
	return uint(pci.bus), errorString(rc)
}

// PCIBusID returns the PCI bus id for a given device
func (d Device) PCIBusID() (string, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(d.dev, &pci)
	return C.GoString(&pci.bus_id[0]), errorString(rc)
}

// PCIID returns the PCI id for a given device
func (d Device) PCIID() (uint, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(d.dev, &pci)
	return uint(pci.pci_device_id), errorString(rc)
}

// PCILinkSpeed returns the current PCI link speed for a given device
func (d Device) PCILinkSpeed() (string, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(d.dev, &pci)
	return C.GoString(&pci.caps.link_speed[0]), errorString(rc)
}

// PCILinkWidth returns the current PCI link width for a given device
func (d Device) PCILinkWidth() (string, error) {
	var pci C.hlml_pci_info_t

	rc := C.hlml_device_get_pci_info(d.dev, &pci)
	return C.GoString(&pci.caps.link_width[0]), errorString(rc)
}

// MemoryInfo returns the current memory usage in bytes for total, used, free
func (d Device) MemoryInfo() (uint64, uint64, uint64, error) {
	var mem C.hlml_memory_t

	rc := C.hlml_device_get_memory_info(d.dev, &mem)
	return uint64(mem.total), uint64(mem.used), uint64(mem.total - mem.used), errorString(rc)
}

// UtilizationInfo returns the utilization aip rate for a given device
func (d Device) UtilizationInfo() (uint, error) {
	var util C.hlml_utilization_t

	rc := C.hlml_device_get_utilization_rates(d.dev, &util)
	return uint(util.aip), errorString(rc)
}

// SOCClockInfo returns the SoC clock frequency for a given device
func (d Device) SOCClockInfo() (uint, error) {
	var freq C.uint

	rc := C.hlml_device_get_clock_info(d.dev, C.HLML_CLOCK_SOC, &freq)
	return uint(freq), errorString(rc)
}

// ICClockInfo returns the IC clock frequency for a given device
func (d Device) ICClockInfo() (uint, error) {
	var freq C.uint

	rc := C.hlml_device_get_clock_info(d.dev, C.HLML_CLOCK_IC, &freq)
	return uint(freq), errorString(rc)
}

// MMEClockInfo returns the MME clock frequency for a given device
func (d Device) MMEClockInfo() (uint, error) {
	var freq C.uint

	rc := C.hlml_device_get_clock_info(d.dev, C.HLML_CLOCK_MME, &freq)
	return uint(freq), errorString(rc)
}

// TPCClockInfo returns the TPC clock frequency for a given device
func (d Device) TPCClockInfo() (uint, error) {
	var freq C.uint

	rc := C.hlml_device_get_clock_info(d.dev, C.HLML_CLOCK_TPC, &freq)
	return uint(freq), errorString(rc)
}

// PowerUsage returns the power usage in milliwatts for a given device
func (d Device) PowerUsage() (uint, error) {
	var power C.uint

	rc := C.hlml_device_get_power_usage(d.dev, &power)
	return uint(power), errorString(rc)
}

// Temperature returns the temperature in celsius for a given device
func (d Device) Temperature() (uint, uint, error) {
	var onBoard C.uint
	var onChip C.uint

	rc := C.hlml_device_get_temperature(d.dev, C.HLML_TEMPERATURE_ON_BOARD, &onBoard)
	rc = C.hlml_device_get_temperature(d.dev, C.HLML_TEMPERATURE_ON_AIP, &onChip)
	return uint(onBoard), uint(onChip), errorString(rc)
}

// ECCVolatileErrors returns the running out of ECC errors in volatile memory
func (d Device) ECCVolatileErrors() (uint64, error) {
	var eccErr C.ulonglong

	rc := C.hlml_device_get_total_ecc_errors(d.dev, C.HLML_MEMORY_ERROR_TYPE_UNCORRECTED, C.HLML_VOLATILE_ECC, &eccErr)
	return uint64(eccErr), errorString(rc)
}

// ECCAggregateErrors returns the running out of ECC errors in aggregate memory
func (d Device) ECCAggregateErrors() (uint64, error) {
	var eccErr C.ulonglong

	rc := C.hlml_device_get_total_ecc_errors(d.dev, C.HLML_MEMORY_ERROR_TYPE_UNCORRECTED, C.HLML_AGGREGATE_ECC, &eccErr)
	return uint64(eccErr), errorString(rc)
}

// HLRevision returns the revision of the HL library
func (d Device) HLRevision() (int, error) {
	var rev C.int

	rc := C.hlml_device_get_hl_revision(d.dev, &rev)
	return int(rev), errorString(rc)
}

// PCBVersion returns the PCB version
func (d Device) PCBVersion() (string, error) {
	var pcb C.hlml_pcb_info_t

	rc := C.hlml_device_get_pcb_info(d.dev, &pcb)
	return C.GoString(&pcb.pcb_ver[0]), errorString(rc)
}

// PCBAssemblyVersion returns the PCB Assembly info
func (d Device) PCBAssemblyVersion() (string, error) {
	var pcb C.hlml_pcb_info_t

	rc := C.hlml_device_get_pcb_info(d.dev, &pcb)
	return C.GoString(&pcb.pcb_assembly_ver[0]), errorString(rc)
}

// SerialNumber returns the device serial number
func (d Device) SerialNumber() (string, error) {
	var serial [szUUID]C.char

	rc := C.hlml_device_get_serial(d.dev, &serial[0], szUUID)
	return C.GoString(&serial[0]), errorString(rc)
}

// BoardID returns an ID for the PCB board
func (d Device) BoardID() (uint, error) {
	var id C.uint

	rc := C.hlml_device_get_board_id(d.dev, &id)
	return uint(id), errorString(rc)
}

// PCIeTX returns PCIe transmit throughput
func (d Device) PCIeTX() (uint, error) {
	var val C.uint

	rc := C.hlml_device_get_pcie_throughput(d.dev, C.HLML_PCIE_UTIL_TX_BYTES, &val)
	return uint(val), errorString(rc)
}

// PCIeRX returns PCIe receive throughput
func (d Device) PCIeRX() (uint, error) {
	var val C.uint

	rc := C.hlml_device_get_pcie_throughput(d.dev, C.HLML_PCIE_UTIL_RX_BYTES, &val)
	return uint(val), errorString(rc)
}

// PCIReplayCounter returns PCIe replay count
func (d Device) PCIReplayCounter() (uint, error) {
	var val C.uint

	rc := C.hlml_device_get_pcie_replay_counter(d.dev, &val)
	return uint(val), errorString(rc)
}

// PCIeLinkGeneration returns PCIe replay count
func (d Device) PCIeLinkGeneration() (uint, error) {
	var gen C.uint

	rc := C.hlml_device_get_curr_pcie_link_generation(d.dev, &gen)
	return uint(gen), errorString(rc)
}

// PCIeLinkWidth returns PCIe replay count
func (d Device) PCIeLinkWidth() (uint, error) {
	var width C.uint

	rc := C.hlml_device_get_curr_pcie_link_width(d.dev, &width)
	return uint(width), errorString(rc)
}

// ClockThrottleReasons returns current clock throttle reasons
func (d Device) ClockThrottleReasons() (uint64, error) {
	var reasons C.ulonglong

	rc := C.hlml_device_get_current_clocks_throttle_reasons(d.dev, &reasons)
	return uint64(reasons), errorString(rc)
}

// EnergyConsumptionCounter returns current clock throttle reasons
func (d Device) EnergyConsumptionCounter() (uint64, error) {
	var energy C.ulonglong

	rc := C.hlml_device_get_total_energy_consumption(d.dev, &energy)
	return uint64(energy), errorString(rc)
}

// FWVersion returns the firmware version for a given device
func FWVersion(idx uint) (string, string, error) {
	kernel, err := ioutil.ReadFile(HLDriverPath + "/hl" + fmt.Sprint(idx) + "/armcp_kernel_ver")
	if err != nil {
		return "", "", fmt.Errorf("File reading error %s", err)
	}

	uboot, err := ioutil.ReadFile(HLDriverPath + "/hl" + fmt.Sprint(idx) + "/uboot_ver")
	if err != nil {
		return "", "", fmt.Errorf("File reading error %s", err)
	}
	return string(kernel), string(uboot), nil
}

// SystemDriverVersion returns the driver version on the system
func SystemDriverVersion() (string, error) {
	driver, err := ioutil.ReadFile(HLModulePath + "/version")

	if err != nil {
		return "", fmt.Errorf("File reading error %s", err)
	}
	return string(driver), nil
}

func hlmlNewEventSet() EventSet {
	var set C.hlml_event_set_t
	C.hlml_event_set_create(&set)

	return EventSet{set}
}

func hlmlRegisterEventForDevice(es EventSet, event int, uuid string) error {

	deviceHandle, err := DeviceHandleByUUID(uuid)

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
	uuid, _ := Device{data.device}.UUID()

	return Event{
			UUID:  uuid,
			Etype: uint64(data.event_type),
		},
		errorString(r)
}
