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
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
)

var (
	logger    *log.Logger
	ctx       context.Context
	rootCfg   = source.Config()
	exportCfg = exportConfig{}
	rootCmd   = &cobra.Command{
		Use:   "ttn-lw-migrate",
		Short: "Migrate from other LoRaWAN network servers to The Things Stack",

		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logLevel := log.InfoLevel
			if rootCfg.Verbose {
				logLevel = log.DebugLevel
			}
			logHandler, err := log.NewZap("console")
			if err != nil {
				return err
			}
			logger = log.NewLogger(
				logHandler,
				log.WithLevel(logLevel),
			)
			ctx = log.NewContext(context.Background(), logger)

			rpclog.ReplaceGrpcLogger(logger)
			return nil
		},
	}
)

// Execute runs the root command and returns the exit code.
func Execute() int {
	if len(os.Args) >= 2 && source.SetActiveSource(os.Args[1]) {
		rootCmd.SetArgs(os.Args[2:])
	}

	// NOTE: The flag set contents depend on the active source.
	// As such, they should be initialized after the active source has
	// been determined.
	applicationsCmd.Flags().AddFlagSet(source.ApplicationFlagSet())
	devicesCmd.Flags().AddFlagSet(source.DevicesFlagSet())

	rootCmd.AddCommand(applicationsCmd)
	rootCmd.AddCommand(devicesCmd)

	if err := rootCmd.Execute(); err != nil {
		printStack(os.Stderr, err)
		return 1
	}
	return 0
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&rootCfg.DryRun,
		"dry-run",
		false,
		"Do everything except resetting root and session keys of exported devices")
	rootCmd.PersistentFlags().BoolVar(&rootCfg.Verbose,
		"verbose",
		false,
		"Verbose output")
	rootCmd.PersistentFlags().StringVar(&rootCfg.FrequencyPlansURL,
		"frequency-plans-url",
		"https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans/master",
		"URL for fetching frequency plans")
}
