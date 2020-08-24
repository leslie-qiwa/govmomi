/*
Copyright (c) 2016-2017 VMware, Inc. All Rights Reserved.

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

package pci

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type toggle struct {
	*flags.ClientFlag
	*flags.HostSystemFlag
}

func init() {
	cli.Register("host.hardware.pci.toggle.passthrough", &toggle{})
}

func (cmd *toggle) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.ClientFlag, ctx = flags.NewClientFlag(ctx)
	cmd.ClientFlag.Register(ctx, f)

	cmd.HostSystemFlag, ctx = flags.NewHostSystemFlag(ctx)
	cmd.HostSystemFlag.Register(ctx, f)

}

func (cmd *toggle) Process(ctx context.Context) error {
	if err := cmd.HostSystemFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *toggle) Description() string {
	return `
Examples:
  govc host.hardware.pci.toggle.passthrough deviceAddress`
}

func (cmd *toggle) Run(ctx context.Context, f *flag.FlagSet) error {
	if len(f.Args()) == 0 {
		return errors.New("one device address is required at least")
	}

	addrs := map[string]bool {}
	for _, addr := range f.Args() {
		addrs[addr] = false
	}
	c, err := cmd.Client()
	if err != nil {
		return err
	}

	reply := []mo.HostSystem{}
	props := []string{"hardware.pciDevice", "config.pciPassthruInfo"}

	// We could do without the -host flag, leaving it for compat
	host, err := cmd.HostSystemIfSpecified()
	if err != nil {
		return err
	}

	// Default only if there is a single host
	if host == nil {
		host, err = cmd.HostSystem()
		if err != nil {
			return err
		} else if host == nil {
			return errors.New("Host is not specified")
		}
	}

	refs := []types.ManagedObjectReference{host.Reference()}

	pc := property.DefaultCollector(c)
	err = pc.Retrieve(ctx, refs, props, &reply)
	if err != nil {
		return err
	}

	if len(reply) == 0 {
		return errors.New("System not exist")
	}

	config := []types.BaseHostPciPassthruConfig{}
	for _, system := range reply {
		for _, pi := range system.Config.PciPassthruInfo {
			info := pi.GetHostPciPassthruInfo()
			_, ok := addrs[info.Id]
			if !ok {
				continue
			}
			if !info.PassthruCapable {
				return fmt.Errorf("%s is not capable to toggle pci passthrough", info.Id)
			}
			addrs[info.Id] = true
			config = append(config, &types.HostPciPassthruConfig{
				Id: info.Id, PassthruEnabled: !info.PassthruEnabled,
			})
		}
	}

	for id, found := range addrs {
		if !found {
			return fmt.Errorf("%s is not found in device list", id)
		}
	}
	s, err := host.ConfigManager().PciPassthruSystem(ctx)
	if err != nil {
		return err
	}
	return s.Update(ctx, config)
}