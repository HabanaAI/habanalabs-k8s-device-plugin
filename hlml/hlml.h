/* SPDX-License-Identifier: MIT
 *
 * Copyright 2016-2019 HabanaLabs, Ltd.
 * All Rights Reserved.
 *
 */

#ifndef __HLML_H__
#define __HLML_H__

#define PCI_DOMAIN_LEN		5
#define PCI_ADDR_LEN		((PCI_DOMAIN_LEN) + 10)

/* Event about single/double bit ECC errors. */
#define HLML_EVENT_ECC_ERR		(1 << 0)
/* Event about critical errors that occurred on the device */
#define HLML_EVENT_CRITICAL_ERR		(1 << 1)
/* Event about changes in clock rate */
#define HLML_EVENT_CLOCK_RATE		(1 << 2)

/* Enum for returned values of the different APIs */
enum hlml_return_t {
	HLML_SUCCESS = 0,
	HLML_ERROR_UNINITIALIZED = 1,
	HLML_ERROR_INVALID_ARGUMENT = 2,
	HLML_ERROR_NOT_SUPPORTED = 3,
	HLML_ERROR_ALREADY_INITIALIZED = 5,
	HLML_ERROR_NOT_FOUND = 6,
	HLML_ERROR_INSUFFICIENT_SIZE = 7,
	HLML_ERROR_DRIVER_NOT_LOADED = 9,
	HLML_ERROR_AIP_IS_LOST = 15,
	HLML_ERROR_MEMORY = 20,
	HLML_ERROR_NO_DATA = 21,
	HLML_ERROR_UNKNOWN = 49,
};

/*
 * bus - The bus on which the device resides, 0 to 0xf
 * busId - The tuple domain:bus:device.function
 * device - The device's id on the bus, 0 to 31
 * domain - The PCI domain on which the device's bus resides
 * pciDeviceId - The combined 16b deviceId and 16b vendor id
 */
struct hlml_pci_info_t {
	unsigned int bus;
	char busId[PCI_ADDR_LEN];
	unsigned int device;
	unsigned int domain;
	unsigned int pciDeviceId;
};

enum hlml_clock_type_t {
	HLML_CLOCK_SOC = 0,
	HLML_CLOCK_IC = 1,
	HLML_CLOCK_MME = 2,
	HLML_CLOCK_TPC = 3,
	HLML_CLOCK_COUNT
};

struct hlml_utilization_t {
	unsigned int aip;
};

struct hlml_memory_t {
	unsigned long long free;
	unsigned long long total; /* Total installed memory (in bytes) */
	unsigned long long used;
};

enum hlml_temperature_sensors_t {
	HLML_TEMPERATURE_ON_AIP = 0,
	HLML_TEMPERATURE_ON_BOARD = 1
};

enum hlml_temperature_thresholds_t {
	HLML_TEMPERATURE_THRESHOLD_SHUTDOWN = 0,
	HLML_TEMPERATURE_THRESHOLD_SLOWDOWN = 1,
	HLML_TEMPERATURE_THRESHOLD_MEM_MAX = 2,
	HLML_TEMPERATURE_THRESHOLD_GPU_MAX = 3,
	HLML_TEMPERATURE_THRESHOLD_COUNT
};

enum hlml_enable_state_t {
	HLML_FEATURE_DISABLED = 0,
	HLML_FEATURE_ENABLED = 1
};

enum hlml_p_states_t {
	HLML_PSTATE_0 = 0,
	HLML_PSTATE_UNKNOWN = 32
};

enum hlml_memory_error_type_t {
	HLML_MEMORY_ERROR_TYPE_CORRECTED = 0,
	HLML_MEMORY_ERROR_TYPE_UNCORRECTED = 1,
	HLML_MEMORY_ERROR_TYPE_COUNT
};

enum hlml_ecc_counter_type_t {
	HLML_VOLATILE_ECC = 0,
	HLML_AGGREGATE_ECC = 1,
	HLML_ECC_COUNTER_TYPE_COUNT
};

typedef void* hlmlDevice_t;

typedef struct hlml_event_data {
	hlmlDevice_t device; /* Specific device where the event occurred. */
	unsigned long long event_type; /* Specific event that occurred */
} hlmlEventData_t;

typedef void* hlmlEventSet_t;

/* supported APIs */
hlml_return_t hlml_init(void);

hlml_return_t hlml_init_with_flags(unsigned int flags);

hlml_return_t hlml_shutdown(void);

hlml_return_t hlml_device_get_count(unsigned int *device_count);

hlml_return_t hlml_device_get_handle_by_pci_bus_id(const char *pci_addr, hlmlDevice_t *device);

hlml_return_t hlml_device_get_handle_by_index(unsigned int index, hlmlDevice_t *device);

hlml_return_t hlml_device_get_name(hlmlDevice_t device, char *name,
				   unsigned int  length);

hlml_return_t hlml_device_get_pci_info(hlmlDevice_t device,
				       hlml_pci_info_t *pci);

hlml_return_t hlml_device_get_clock_info(hlmlDevice_t device,
					 hlml_clock_type_t type,
					 unsigned int *clock);

hlml_return_t hlml_device_get_max_clock_info(hlmlDevice_t device,
					     hlml_clock_type_t type,
					     unsigned int *clock);

hlml_return_t hlml_device_get_utilization_rates(hlmlDevice_t device,
					hlml_utilization_t *utilization);

hlml_return_t hlml_device_get_memory_info(hlmlDevice_t device,
					  hlml_memory_t *memory);

hlml_return_t hlml_device_get_temperature(hlmlDevice_t device,
					  hlml_temperature_sensors_t sensor_type,
					  unsigned int *temp);

hlml_return_t hlml_device_get_temperature_threshold(hlmlDevice_t device,
				hlml_temperature_thresholds_t threshold_type,
				unsigned int *temp);

hlml_return_t hlml_device_get_persistence_mode(hlmlDevice_t device,
						hlml_enable_state_t *mode);

hlml_return_t hlml_device_get_performance_state(hlmlDevice_t device,
						hlml_p_states_t *p_state);

hlml_return_t hlml_device_get_power_usage(hlmlDevice_t device,
					  unsigned int *power);

hlml_return_t hlml_device_get_power_management_default_limit(hlmlDevice_t device,
						unsigned int *default_limit);

hlml_return_t hlml_device_get_ecc_mode(hlmlDevice_t device,
				       hlml_enable_state_t *current,
				       hlml_enable_state_t *pending);

hlml_return_t hlml_device_get_total_ecc_errors(hlmlDevice_t device,
					hlml_memory_error_type_t error_type,
					hlml_ecc_counter_type_t counter_type,
					unsigned long long *ecc_counts);

hlml_return_t hlml_device_get_uuid(hlmlDevice_t device,
				   char *uuid,
				   unsigned int length);

hlml_return_t hlml_device_get_minor_number(hlmlDevice_t device,
					   unsigned int *minor_number);

hlml_return_t hlml_device_register_events(hlmlDevice_t device,
					  unsigned long long event_types,
					  hlmlEventSet_t set);

hlml_return_t hlml_event_set_create(hlmlEventSet_t *set);

hlml_return_t hlml_event_set_free(hlmlEventSet_t set);

hlml_return_t hlml_event_set_wait(hlmlEventSet_t set,
				  hlmlEventData_t *data,
				  unsigned int timeoutms);

#endif /* __HLML_H__ */
