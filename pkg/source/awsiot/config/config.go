// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package config

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iotwireless"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
)

type Config struct {
	source.Config

	Client *iotwireless.Client

	AppID           string
	FrequencyPlanID string
	NoSession       bool

	flags   *pflag.FlagSet
	fpStore *frequencyplans.Store
}

func New() *Config {
	c := &Config{
		flags: new(pflag.FlagSet),
	}

	c.flags.StringVar(&c.AppID,
		"app-id",
		os.Getenv("APP_ID"),
		"Application ID for the exported devices")
	c.flags.StringVar(&c.FrequencyPlanID,
		"frequency-plan-id",
		os.Getenv("FREQUENCY_PLAN_ID"),
		"Frequency Plan ID for the exported devices")
	c.flags.BoolVar(&c.NoSession,
		"no-session",
		os.Getenv("NO_SESSION") == "true",
		"TTS export devices without session")

	return c
}

func (c *Config) Initialize(rootCfg source.Config) error {
	c.Config = rootCfg

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return err
	}
	c.Client = iotwireless.NewFromConfig(cfg)

	fpFetcher, err := fetch.FromHTTP(http.DefaultClient, c.FrequencyPlansURL)
	if err != nil {
		return err
	}
	c.fpStore = frequencyplans.NewStore(fpFetcher)

	return nil
}

// Flags returns the flags for the configuration.
func (c *Config) Flags() *pflag.FlagSet {
	return c.flags
}

func (c *Config) FPStore() *frequencyplans.Store {
	return c.fpStore
}
