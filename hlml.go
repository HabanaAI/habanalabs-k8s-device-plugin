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
	if ret == C.HLML_SUCCESS {
		return nil
	}
	return fmt.Errorf("Unknown error")
}

func hlmlInit() error {
	return errorString(C.hlml_init())
}

func hlmlShutdown() error {
	return errorString(C.hlml_shutdown())
}
