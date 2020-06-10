/* SPDX-License-Identifier: MIT
 *
 * Copyright 2016-2019 HabanaLabs, Ltd.
 * All Rights Reserved.
 *
 */

#ifndef __HLML_H__
#define __HLML_H__

#include <net/ethernet.h>

#ifdef __cplusplus
extern "C" {
#endif

#define PCI_DOMAIN_LEN		5
#define PCI_ADDR_LEN		((PCI_DOMAIN_LEN) + 10)
#define PCI_LINK_INFO_LEN	10

#define HLML_DEVICE_MAC_MAX_ADDRESSES	20

/* Event about single/double bit ECC errors. */
#define HLML_EVENT_ECC_ERR		(1 << 0)
/* Event about critical errors that occurred on the device */
#define HLML_EVENT_CRITICAL_ERR		(1 << 1)
/* Event about changes in clock rate */
#define HLML_EVENT_CLOCK_RATE		(1 << 2)

/* Enum for returned values of the different APIs */
typedef enum hlml_return {
	HLML_SUCCESS = 0,
	HLML_ERROR_UNINITIALIZED = 1,
	HLML_ERROR_INVALID_ARGUMENT = 2,
	HLML_ERROR_NOT_SUPPORTED = 3,
	HLML_ERROR_ALREADY_INITIALIZED = 5,
	HLML_ERROR_NOT_FOUND = 6,
	HLML_ERROR_INSUFFICIENT_SIZE = 7,
	HLML_ERROR_DRIVER_NOT_LOADED = 9,
	HLML_ERROR_TIMEOUT = 10,
	HLML_ERROR_AIP_IS_LOST = 15,
	HLML_ERROR_MEMORY = 20,
	HLML_ERROR_NO_DATA = 21,
	HLML_ERROR_UNKNOWN = 49,
} hlml_return_t;

/*
 * link_speed - current pci link speed
 * link_width - current pci link width
 */
typedef struct hlml_pci_cap {
	char link_speed[PCI_LINK_INFO_LEN];
	char link_width[PCI_LINK_INFO_LEN];
} hlml_pci_cap_t;

/*
 * bus - The bus on which the device resides, 0 to 0xf
 * bus_id - The tuple domain:bus:device.function
 * device - The device's id on the bus, 0 to 31
 * domain - The PCI domain on which the device's bus resides
 * pci_device_id - The combined 16b deviceId and 16b vendor id
 */
typedef struct hlml_pci_info {
	unsigned int bus;
	char bus_id[PCI_ADDR_LEN];
	unsigned int device;
	unsigned int domain;
	unsigned int pci_device_id;
	hlml_pci_cap_t caps;
} hlml_pci_info_t;

typedef enum hlml_clock_type {
	HLML_CLOCK_SOC = 0,
	HLML_CLOCK_IC = 1,
	HLML_CLOCK_MME = 2,
	HLML_CLOCK_TPC = 3,
	HLML_CLOCK_COUNT
} hlml_clock_type_t;

typedef struct hlml_utilization {
	unsigned int aip;
} hlml_utilization_t;

typedef struct hlml_memory {
	unsigned long long free;
	unsigned long long total; /* Total installed memory (in bytes) */
	unsigned long long used;
} hlml_memory_t;

typedef enum hlml_temperature_sensors {
	HLML_TEMPERATURE_ON_AIP = 0,
	HLML_TEMPERATURE_ON_BOARD = 1
} hlml_temperature_sensors_t;

typedef enum hlml_temperature_thresholds {
	HLML_TEMPERATURE_THRESHOLD_SHUTDOWN = 0,
	HLML_TEMPERATURE_THRESHOLD_SLOWDOWN = 1,
	HLML_TEMPERATURE_THRESHOLD_MEM_MAX = 2,
	HLML_TEMPERATURE_THRESHOLD_GPU_MAX = 3,
	HLML_TEMPERATURE_THRESHOLD_COUNT
} hlml_temperature_thresholds_t;

typedef enum hlml_enable_state {
	HLML_FEATURE_DISABLED = 0,
	HLML_FEATURE_ENABLED = 1
} hlml_enable_state_t;

typedef enum hlml_p_states {
	HLML_PSTATE_0 = 0,
	HLML_PSTATE_UNKNOWN = 32
} hlml_p_states_t;

typedef enum hlml_memory_error_type {
	HLML_MEMORY_ERROR_TYPE_CORRECTED = 0, /* Not supported*/
	HLML_MEMORY_ERROR_TYPE_UNCORRECTED = 1,
	HLML_MEMORY_ERROR_TYPE_COUNT
} hlml_memory_error_type_t;

typedef enum hlml_ecc_counter_type {
	HLML_VOLATILE_ECC = 0,
	HLML_AGGREGATE_ECC = 1,
	HLML_ECC_COUNTER_TYPE_COUNT
} hlml_ecc_counter_type_t;

typedef enum hlml_err_inject {
	HLML_ERR_INJECT_ENDLESS_COMMAND = 0,
	HLML_ERR_INJECT_NON_FATAL_EVENT = 1,
	HLML_ERR_INJECT_FATAL_EVENT = 2,
	HLML_ERR_INJECT_LOSS_OF_HEARTBEAT = 3,
	HLML_ERR_INJECT_THERMAL_EVENT = 4,
	HLML_ERR_INJECT_COUNT
} hlml_err_inject_t;

typedef void* hlml_device_t;

typedef struct hlml_event_data {
	hlml_device_t device; /* Specific device where the event occurred. */
	unsigned long long event_type; /* Specific event that occurred */
} hlml_event_data_t;

typedef void* hlml_event_set_t;

typedef struct hlml_mac_info {
	unsigned char addr[ETHER_ADDR_LEN];
	int id;
} hlml_mac_info_t;

/* supported APIs */
hlml_return_t hlml_init(void);

hlml_return_t hlml_init_with_flags(unsigned int flags);

hlml_return_t hlml_shutdown(void);

hlml_return_t hlml_device_get_count(unsigned int *device_count);

hlml_return_t hlml_device_get_handle_by_pci_bus_id(const char *pci_addr, hlml_device_t *device);

hlml_return_t hlml_device_get_handle_by_index(unsigned int index, hlml_device_t *device);

hlml_return_t hlml_device_get_handle_by_UUID (const char* uuid, hlml_device_t *device);

hlml_return_t hlml_device_get_name(hlml_device_t device, char *name,
				   unsigned int  length);

hlml_return_t hlml_device_get_pci_info(hlml_device_t device,
				       hlml_pci_info_t *pci);

hlml_return_t hlml_device_get_clock_info(hlml_device_t device,
					 hlml_clock_type_t type,
					 unsigned int *clock);

hlml_return_t hlml_device_get_max_clock_info(hlml_device_t device,
					     hlml_clock_type_t type,
					     unsigned int *clock);

hlml_return_t hlml_device_get_utilization_rates(hlml_device_t device,
					hlml_utilization_t *utilization);

hlml_return_t hlml_device_get_memory_info(hlml_device_t device,
					  hlml_memory_t *memory);

hlml_return_t hlml_device_get_temperature(hlml_device_t device,
					  hlml_temperature_sensors_t sensor_type,
					  unsigned int *temp);

hlml_return_t hlml_device_get_temperature_threshold(hlml_device_t device,
				hlml_temperature_thresholds_t threshold_type,
				unsigned int *temp);

hlml_return_t hlml_device_get_persistence_mode(hlml_device_t device,
						hlml_enable_state_t *mode);

hlml_return_t hlml_device_get_performance_state(hlml_device_t device,
						hlml_p_states_t *p_state);

hlml_return_t hlml_device_get_power_usage(hlml_device_t device,
					  unsigned int *power);

hlml_return_t hlml_device_get_power_management_default_limit(hlml_device_t device,
						unsigned int *default_limit);

hlml_return_t hlml_device_get_ecc_mode(hlml_device_t device,
				       hlml_enable_state_t *current,
				       hlml_enable_state_t *pending);

hlml_return_t hlml_device_get_total_ecc_errors(hlml_device_t device,
					hlml_memory_error_type_t error_type,
					hlml_ecc_counter_type_t counter_type,
					unsigned long long *ecc_counts);

hlml_return_t hlml_device_get_uuid(hlml_device_t device,
				   char *uuid,
				   unsigned int length);

hlml_return_t hlml_device_get_minor_number(hlml_device_t device,
					   unsigned int *minor_number);

hlml_return_t hlml_device_register_events(hlml_device_t device,
					  unsigned long long event_types,
					  hlml_event_set_t set);

hlml_return_t hlml_event_set_create(hlml_event_set_t *set);

hlml_return_t hlml_event_set_free(hlml_event_set_t set);

hlml_return_t hlml_event_set_wait(hlml_event_set_t set,
				  hlml_event_data_t *data,
				  unsigned int timeoutms);

hlml_return_t hlml_device_get_mac_info(hlml_device_t device,
				       hlml_mac_info_t *mac_info,
				       unsigned int mac_info_size,
				       unsigned int start_mac_id,
				       unsigned int *actual_mac_count);

hlml_return_t hlml_device_err_inject(hlml_device_t device, hlml_err_inject_t err_type);

#ifdef __cplusplus
}   //extern "C"
#endif

#endif /* __HLML_H__ */
