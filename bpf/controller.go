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
	"context"
	"log"
	"time"

	bpf "github.com/iovisor/gobpf/elf"
)

/*
#include <linux/bpf.h>
*/
import "C"

const (
	lookupInterval        = 1 * time.Second
	packetsKey     uint32 = 0
	bytesKey       uint32 = 1
)

type Controller struct {
	cgroup string
	module *bpf.Module

	packetsHandler func(uint64) error
	bytesHandler   func(uint64) error
}

func (c *Controller) SetPacketsHandler(h func(uint64) error) {
	c.packetsHandler = h
}
func (c *Controller) SetBytesHandler(h func(uint64) error) {
	c.bytesHandler = h
}

func (c *Controller) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(lookupInterval):
			// TODO this needs to be solved differently
			// Maybe sending new values on individual channels?
			packets, err := lookup(c.module, packetsKey)
			if err != nil {
				log.Printf("lookup failed: %s", err)
				continue
			}
			if err := c.packetsHandler(packets); err != nil {
				log.Printf("packetsHandler failed: %s", err)
			}

			bytes, err := lookup(c.module, bytesKey)
			if err != nil {
				log.Printf("lookup failed: %s", err)
				continue
			}
			if err := c.bytesHandler(bytes); err != nil {
				log.Printf("bytesHandler failed: %s", err)
			}
		}
	}
}
