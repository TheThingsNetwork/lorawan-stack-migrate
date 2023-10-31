// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package firefly

import (
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/client"
)

type Config struct {
	client.Config

	src source.Config

	appID           string
	frequencyPlanID string
	joinEUI         string
	macVersion      string

	flags *pflag.FlagSet
}

// New returns a new Firefly configuration.
func New() *Config {
	config := &Config{
		flags: &pflag.FlagSet{},
	}
	config.flags.StringVar(&config.Host,
		"host",
		os.Getenv("FIREFLY_HOST"),
		"Host of the Firefly API")
	config.flags.StringVar(&config.APIKey,
		"api-key",
		os.Getenv("FIREFLY_API_KEY"),
		"Key to access the Firefly API")
	config.flags.StringVar(&config.joinEUI,
		"join-eui",
		os.Getenv("JOIN_EUI"),
		"JoinEUI for the exported devices")
	config.flags.StringVar(&config.frequencyPlanID,
		"frequency-plan-id",
		os.Getenv("FREQUENCY_PLAN_ID"),
		"Frequency Plan ID for the exported devices")
	config.flags.StringVar(&config.macVersion,
		"mac-version",
		os.Getenv("MAC_VERSION"),
		"MAC version for the exported devices")
	config.flags.StringVar(&config.appID,
		"app-id",
		os.Getenv("APP_ID"),
		"Application ID for the exported devices")

	return config
}

var logger *zap.SugaredLogger

// Initialize the configuration.
func (c *Config) Initialize() error {
	cfg := zap.NewProductionConfig()
	if c.src.Verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapLogger, err := cfg.Build()
	if err != nil {
		return err
	}
	logger = zapLogger.Sugar()

	if c.appID == "" {
		return errNoAppID.New()
	}
	if c.Host == "" {
		return errNoHost.New()
	}
	if c.APIKey == "" {
		return errNoAPIKey.New()
	}
	if c.joinEUI == "" {
		return errNoJoinEUI.New()
	}
	if c.frequencyPlanID == "" {
		return errNoFrequencyPlanID.New()
	}
	return nil
}

// Flags returns the flags for the configuration.
func (c *Config) Flags() *pflag.FlagSet {
	return c.flags
}
