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

package ttnv3

import (
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack-migrate/cmd"
)

const sourceName = "ttnv3"

// ttnv3Cmd represents the ttnv3 source.
var ttnv3Cmd = &cobra.Command{
	Use:          sourceName + " ...",
	Short:        "Export devices from The Things Stack",
	SilenceUsage: true,
}

func init() {
	cmd.RootCmd.AddCommand(ttnv3Cmd)
}
