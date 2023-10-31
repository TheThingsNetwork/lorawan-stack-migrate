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

type SourceOptions struct {
	opts, appOpts, devOpts []Option
}

// Extend merges respectable fields from src into s.
func (s *SourceOptions) Extend(src SourceOptions) {
	s.opts = append(s.opts, src.opts...)
	s.appOpts = append(s.appOpts, src.appOpts...)
	s.devOpts = append(s.devOpts, src.devOpts...)
}

// WithSourceOptions returns SourceOptions with opts field set to opts.
func WithSourceOptions(opts ...Option) SourceOptions {
	return SourceOptions{
		opts: opts,
	}
}

// WithApplicationOptions returns SourceOptions with appOpts field set to opts.
func WithApplicationOptions(opts ...Option) SourceOptions {
	return SourceOptions{
		appOpts: opts,
	}
}

// WithDevicesOptions returns SourceOptions with devOpts field set to opts.
func WithDevicesOptions(opts ...Option) SourceOptions {
	return SourceOptions{
		devOpts: opts,
	}
}

// Source returns a new source command.
func Source(sourceName, short string, opts ...SourceOptions) *cobra.Command {
	fs, err := source.FlagSet(sourceName)
	if err != nil {
		panic(err)
	}

	sourceOpts := new(SourceOptions)
	for _, opt := range opts {
		sourceOpts.Extend(opt)
	}

	defaults := []Option{
		WithFlagSet(fs),
		WithPersistentPreRunE(ExecuteParentPersistentPreRun),
	}
	appCmd := Application(
		append(defaults, sourceOpts.appOpts...)...,
	)
	devCmd := Devices(
		append(defaults, sourceOpts.devOpts...)...,
	)

	defaults = []Option{
		WithUse(sourceName + " ..."),
		WithShort(short),
		WithPersistentPreRunE(SourcePersistentPreRunE()),
		WithSubcommands(appCmd, devCmd),
		WithGroupID("sources"),
	}
	cmd := New(
		append(defaults, sourceOpts.opts...)...,
	)

	return cmd
}

// Application returns a new application command.
func Application(opts ...Option) *cobra.Command {
	defaultOpts := []Option{
		WithUse("application ..."),
		WithShort("Export all devices of an application"),
		WithAliases([]string{"applications", "apps", "app", "a"}),
		WithRunE(ExportApplication()),
	}
	return New(append(defaultOpts, opts...)...)
}

// Devices returns a new devices command.
func Devices(opts ...Option) *cobra.Command {
	defaultOpts := []Option{
		WithUse("device ..."),
		WithShort("Export devices by Device ID"),
		WithAliases([]string{"end-devices", "end-device", "devices", "devs", "dev", "d"}),
		WithRunE(ExportDevices()),
	}
	return New(append(defaultOpts, opts...)...)
}

// SourcePersistentPreRunE returns a new function that sets the active source.
func SourcePersistentPreRunE() CobraRunE {
	return func(cmd *cobra.Command, args []string) error {
		s := cmd.Name()
		if ok := source.RootConfig.SetSource(s); !ok {
			return source.ErrNotRegistered.WithAttributes("source", s).New()
		}
		return ExecuteParentPersistentPreRun(cmd, args)
	}
}

// ExecuteParentPersistentPreRun executes cmd's parent's PersistentPreRunE or PersistentPreRun.
func ExecuteParentPersistentPreRun(cmd *cobra.Command, args []string) error {
	if !cmd.HasParent() {
		return nil
	}
	p := cmd.Parent()

	if f := p.PersistentPreRunE; f != nil {
		if err := f(p, args); err != nil {
			return err
		}
	} else if f := p.PersistentPreRun; f != nil {
		f(p, args)
	}
	return nil
}

// ExecuteParentPersistentPostRun executes cmd's parent's PersistentPostRunE or PersistentPostRun.
func ExecuteParentPersistentPostRun(cmd *cobra.Command, args []string) error {
	if !cmd.HasParent() {
		return nil
	}
	p := cmd.Parent()

	if f := p.PersistentPostRunE; f != nil {
		if err := f(p, args); err != nil {
			return err
		}
	} else if f := p.PersistentPostRun; f != nil {
		f(p, args)
	}
	return nil
}
