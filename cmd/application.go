// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package cmd

import (
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

var (
	applicationsCmd = &cobra.Command{
		Use:     "application [app-id] ...",
		Aliases: []string{"applications", "app"},
		Short:   "Export all devices of an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			return exportCommand(cmd, args, func(s source.Source, item string) error {
				return s.RangeDevices(item, exportCfg.exportDev)
			})
		},
	}
)

func init() {
	applicationsCmd.Flags().AddFlagSet(source.ApplicationFlagSet())
	rootCmd.AddCommand(applicationsCmd)
}
