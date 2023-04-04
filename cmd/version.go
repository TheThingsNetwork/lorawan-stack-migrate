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

package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack-migrate/pkg/version"
)

func printVar(k, v string) {
	fmt.Printf("%-20s %s\n", k+":", v)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s: %s\n", cmd.Root().Short, cmd.Root().Name())
		printVar("Version", version.Version)
		if version.BuildDate != "" {
			printVar("Build date", version.BuildDate)
		}
		if version.GitCommit != "" {
			printVar("Git commit", version.GitCommit)
		}
		printVar("Go version", runtime.Version())
		printVar("OS/Arch", runtime.GOOS+"/"+runtime.GOARCH)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
