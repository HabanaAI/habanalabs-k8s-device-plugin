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
