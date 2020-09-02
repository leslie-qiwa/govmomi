/*
Copyright (c) 2016 VMware, Inc. All Rights Reserved.

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

package object

import (
	"context"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

type PciPassthruSystem struct {
	Common
}

func NewPciPassthruSystem(c *vim25.Client, ref types.ManagedObjectReference) *PciPassthruSystem {
	return &PciPassthruSystem{
		Common: NewCommon(c, ref),
	}
}

func (s PciPassthruSystem) Update(ctx context.Context, config []types.BaseHostPciPassthruConfig) error {
	req := types.UpdatePassthruConfig{
		This:   s.Reference(),
		Config: config,
	}

	_, err := methods.UpdatePassthruConfig(ctx, s.c, &req)
	return err
}