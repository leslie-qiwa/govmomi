/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

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

package guest

import (
	"flag"
	"strings"

	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/govc/flags"
)

type GuestFlag struct {
	*flags.ClientFlag
	*flags.VirtualMachineFlag

	*AuthFlag
}

func (flag *GuestFlag) Register(f *flag.FlagSet) {}

func (flag *GuestFlag) Process() error { return nil }

func (flag *GuestFlag) FileManager() (*govmomi.GuestFileManager, error) {
	c, err := flag.Client()
	if err != nil {
		return nil, err
	}
	return c.GuestOperationsManager().FileManager()
}

func (flag *GuestFlag) ProcessManager() (*govmomi.GuestProcessManager, error) {
	c, err := flag.Client()
	if err != nil {
		return nil, err
	}
	return c.GuestOperationsManager().ProcessManager()
}

func (flag *GuestFlag) ParseURL(urlStr string) (*url.URL, error) {
	c, err := flag.Client()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	host := strings.Split(u.Host, ":")
	if host[0] == "*" {
		host[0] = strings.Split(c.Client.URL().Host, ":")[0]
		u.Host = strings.Join(host, ":")
	}

	return u, nil
}