// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

func Source(sourceName, short string, opts ...Option) *cobra.Command {
	fs, _ := source.FlagSet(sourceName)

	appCmd := Application(
		WithFlagSet(fs),
		WithPersistentPreRunE(ExecuteParentPersistentPreRun),
	)
	devCmd := Devices(
		WithFlagSet(fs),
		WithPersistentPreRunE(ExecuteParentPersistentPreRun),
	)

	cmd := New(append(opts,
		WithUse(sourceName+" ..."),
		WithShort(short),
		WithPersistentPreRunE(SourcePersistentPreRunE()),
		WithSubcommands(appCmd, devCmd),
	)...)

	return cmd
}

func Application(opts ...Option) *cobra.Command {
	defaultOpts := []Option{
		WithUse("application ..."),
		WithShort("Export all devices of an application"),
		WithAliases([]string{"applications", "apps", "app", "a"}),
		WithRun(ExportApplication()),
	}
	return New(append(defaultOpts, opts...)...)
}

func Devices(opts ...Option) *cobra.Command {
	defaultOpts := []Option{
		WithUse("device ..."),
		WithShort("Export devices by DevEUI"),
		WithAliases([]string{"end-devices", "end-device", "devices", "devs", "dev", "d"}),
		WithRunE(ExportDevices()),
	}
	return New(append(defaultOpts, opts...)...)
}

func SourcePersistentPreRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		s := cmd.Name()
		if ok := source.RootConfig.SetSource(s); !ok {
			return source.ErrNotRegistered.WithAttributes("source", s).New()
		}
		return ExecuteParentPersistentPreRun(cmd, args)
	}
}
