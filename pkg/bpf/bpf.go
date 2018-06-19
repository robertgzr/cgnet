/*
Copyright 2017 Kinvolk GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bpf

import (
	"errors"
	"fmt"
	"os"
	"unsafe"

	bpf "github.com/iovisor/gobpf/elf"
)

/*
#include <linux/bpf.h>
*/
import "C"

var (
	// BPFProgramPath should be set outside this package or via the BPF_PROG_PATH environment variable
	BPFProgramPath string
	// BPFMapName should be set outside this package or via the BPF_MAP_NAME environment variable
	BPFMapName string
)

func init() {
	if p := os.Getenv("BPF_PROG_PATH"); p != "" {
		BPFProgramPath = p
	}
	if n := os.Getenv("BPF_MAP_NAME"); n != "" {
		BPFMapName = n
	}
}

// Attach attaches the BPF program to the cgroup at cgroupPath
//
// TODO:
// Right now this limits us to monitoring EgressType connections
//
func Attach(cgroupPath string) (*bpf.Module, error) {
	if BPFProgramPath == "" {
		return nil, errors.New("BPF program path unset. Please use BPF_PROG_PATH environment variable")
	}

	b, err := initModule(BPFProgramPath)
	if err != nil {
		return nil, err
	}

	for prog := range b.IterCgroupProgram() {
		if err := bpf.AttachCgroupProgram(prog, cgroupPath, bpf.EgressType); err != nil {
			return nil, fmt.Errorf("error attaching to cgroup %s: %s", cgroupPath, err)
		}
	}

	return b, nil
}

// initModule takes the path to a bpf program object file on disk and loads it
func initModule(path string) (*bpf.Module, error) {
	var b = bpf.NewModule(path)

	if b == nil {
		return nil, fmt.Errorf("system doesn't seem to support BPF")
	}

	if err := b.Load(nil); err != nil {
		return nil, fmt.Errorf("loading module failed: %s", err)
	}
	return b, nil
}

// UpdateKey takes a reference to a bpf.Module and a key/value pair that will be updated in the internal bpf.Map
//
// Mainly used to initialize the map with appropriate zero values.
//
func UpdateKey(b *bpf.Module, key uint32, value uint64) error {
	if BPFMapName == "" {
		return errors.New("BPF map name unset. Please use BPF_MAP_NAME environment variable.")
	}

	mp := b.Map(BPFMapName)
	if err := b.UpdateElement(mp, unsafe.Pointer(&key), unsafe.Pointer(&value), C.BPF_ANY); err != nil {
		return err
	}
	return nil
}

// LookupKey takes a reference to a bpf.Module and a key to a value in the internal map.
func LookupKey(b *bpf.Module, key uint32) (uint64, error) {
	if BPFMapName == "" {
		return 0, errors.New("BPF map name unset. Please use BPF_MAP_NAME environment variable.")
	}

	mp := b.Map(BPFMapName)
	var value uint64
	if err := b.LookupElement(mp, unsafe.Pointer(&key), unsafe.Pointer(&value)); err != nil {
		return 0, err
	}
	return value, nil
}
