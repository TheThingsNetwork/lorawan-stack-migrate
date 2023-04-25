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
	"github.com/spf13/pflag"
)

type CobraRun func(cmd *cobra.Command, args []string)

type CobraRunE func(cmd *cobra.Command, args []string) error

// Option allows extending the command when it is instantiated with New.
type Option func(*cobra.Command)

// New returns a new command.
func New(opts ...Option) *cobra.Command {
	cmd := new(cobra.Command)
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

// WithUse returns an option that sets the command's Use field.
func WithUse(s string) Option {
	return func(c *cobra.Command) { c.Use = s }
}

// WithShort returns an option that sets the command's Short field.
func WithShort(s string) Option {
	return func(c *cobra.Command) { c.Short = s }
}

// WithAliases returns an option that sets the command's Aliases field.
func WithAliases(a []string) Option {
	return func(c *cobra.Command) { c.Aliases = a }
}

// TODO: After dependency update (https://github.com/TheThingsNetwork/lorawan-stack-migrate/issues/72)
// Add `WithGroup` option.

// WithPersistentPreRun returns an option that sets the command's PersistentPreRun field.
func WithPersistentPreRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.PersistentPreRun = f }
}

// WithPersistentPreRunE returns an option that sets the command's PersistentPreRunE field.
func WithPersistentPreRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.PersistentPreRunE = f }
}

// WithPreRun returns an option that sets the command's PreRun field.
func WithPreRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.PreRun = f }
}

// WithPreRunE returns an option that sets the command's PreRunE field.
func WithPreRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.PreRunE = f }
}

// WithRun returns an option that sets the command's Run field.
func WithRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.Run = f }
}

// WithRunE returns an option that sets the command's RunE field.
func WithRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.RunE = f }
}

// WithPostRun returns an option that sets the command's PostRun field.
func WithPostRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.PostRun = f }
}

// WithPostRunE returns an option that sets the command's PostRunE field.
func WithPostRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.PostRunE = f }
}

// WithPersistentPostRun returns an option that sets the command's PersistentPostRun field.
func WithPersistentPostRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.PersistentPostRun = f }
}

// WithPersistentPostRunE returns an option that sets the command's PersistentPostRunE field.
func WithPersistentPostRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.PersistentPostRunE = f }
}

// WithSubcommands returns an option that adds commands cmd to the command.
func WithSubcommands(cmd ...*cobra.Command) Option {
	return func(c *cobra.Command) { c.AddCommand(cmd...) }
}

// WithFlagSet returns an option that adds a flag set to the command's flags.
func WithFlagSet(fs *pflag.FlagSet) Option {
	return func(c *cobra.Command) { c.Flags().AddFlagSet(fs) }
}
