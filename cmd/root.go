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
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
)

var (
	logger  *log.Logger
	ctx     context.Context
	rootCmd = &cobra.Command{
		Use:   "ttn-lw-migrate",
		Short: "Migrate from other LoRaWAN network servers to The Things Stack",

		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logLevel := log.InfoLevel
			if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
				logLevel = log.DebugLevel
			}

			logger = log.NewLogger(
				log.WithLevel(logLevel),
				log.WithHandler(log.NewCLI(os.Stderr)),
			)
			ctx = log.NewContext(context.Background(), logger)

			rpclog.ReplaceGrpcLogger(logger)
			return nil
		},
	}
)

// Execute runs the root command and returns the exit code.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		printStack(os.Stderr, err)
		return 1
	}
	return 0
}

func init() {
	rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")
}
