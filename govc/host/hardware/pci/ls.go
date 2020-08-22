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
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
)

type ls struct {
	*flags.ClientFlag
	*flags.HostSystemFlag
}

func init() {
	cli.Register("host.hardware.pci.ls", &ls{})
}

func (cmd *ls) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.ClientFlag, ctx = flags.NewClientFlag(ctx)
	cmd.ClientFlag.Register(ctx, f)

	cmd.HostSystemFlag, ctx = flags.NewHostSystemFlag(ctx)
	cmd.HostSystemFlag.Register(ctx, f)
}

func (cmd *ls) Process(ctx context.Context) error {
	if err := cmd.HostSystemFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *ls) Description() string {
	return `
Examples:
  govc host.hardware.pci.ls`
}

func (cmd *ls) Run(ctx context.Context, f *flag.FlagSet) error {
	c, err := cmd.Client()
	if err != nil {
		return err
	}

	var (
		res     infoResult
		objects []*object.HostSystem
	)

	props := []string{"summary", "hardware.pciDevice", "config.pciPassthruInfo"}

	// We could do without the -host flag, leaving it for compat
	host, err := cmd.HostSystemIfSpecified()
	if err != nil {
		return err
	}

	// Default only if there is a single host
	if host == nil && f.NArg() == 0 {
		host, err = cmd.HostSystem()
		if err != nil {
			return err
		}
	}

	if host != nil {
		objects = append(objects, host)
	} else {
		objects, err = cmd.HostSystems(f.Args())
		if err != nil {
			return err
		}
	}

	if len(objects) != 0 {
		refs := make([]types.ManagedObjectReference, 0, len(objects))
		for _, o := range objects {
			refs = append(refs, o.Reference())
		}

		pc := property.DefaultCollector(c)
		err = pc.Retrieve(ctx, refs, props, &res.HostSystems)
		if err != nil {
			return err
		}
	}

	return cmd.WriteResult(&res)
}

type infoResult struct {
	HostSystems []mo.HostSystem
}

func (r *infoResult) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	for _, obj := range r.HostSystems {
		infos := map[string]*types.HostPciPassthruInfo{}
		for _, o := range obj.Config.PciPassthruInfo {
			info := o.GetHostPciPassthruInfo()
			infos[info.Id] = info
		}
		fmt.Fprintf(tw, "Name:\t%s\n", obj.Summary.Config.Name)
		fmt.Fprintf(tw, "  Address\tDescription\tParent\tPassthrough\n")
		for _, o := range obj.Hardware.PciDevice {
			passthrough := "Not Capable"
			info, ok := infos[o.Id]
			if ok {
				if info.PassthruActive {
					passthrough = "Active"
				} else if info.PassthruEnabled {
					passthrough = "Enabled"
				} else if info.PassthruCapable {
					passthrough = "Disabled"
				}
			}
			fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n", o.Id, o.VendorName, o.ParentBridge, passthrough)
		}
	}
	tw.Flush()
	return nil
}
