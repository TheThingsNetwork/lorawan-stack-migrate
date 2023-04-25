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

type Option func(*cobra.Command)

func New(opts ...Option) *cobra.Command {
	cmd := new(cobra.Command)
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

func WithUse(s string) Option {
	return func(c *cobra.Command) { c.Use = s }
}

func WithShort(s string) Option {
	return func(c *cobra.Command) { c.Short = s }
}

func WithAliases(a []string) Option {
	return func(c *cobra.Command) { c.Aliases = a }
}

func WithPersistentPreRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.PersistentPreRun = f }
}

func WithPersistentPreRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.PersistentPreRunE = f }
}

func WithPreRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.PreRun = f }
}

func WithPreRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.PreRunE = f }
}

func WithRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.Run = f }
}

func WithRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.RunE = f }
}

func WithPostRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.PostRun = f }
}

func WithPostRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.PostRunE = f }
}

func WithPersistentPostRun(f CobraRun) Option {
	return func(c *cobra.Command) { c.PersistentPostRun = f }
}

func WithPersistentPostRunE(f CobraRunE) Option {
	return func(c *cobra.Command) { c.PersistentPostRunE = f }
}

func WithSubcommands(cmd ...*cobra.Command) Option {
	return func(c *cobra.Command) { c.AddCommand(cmd...) }
}

func WithFlagSet(fs *pflag.FlagSet) Option {
	return func(c *cobra.Command) { c.Flags().AddFlagSet(fs) }
}
