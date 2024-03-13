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

package ttnv2

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	"github.com/TheThingsNetwork/go-utils/handlers/cli"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	ttnapex "github.com/TheThingsNetwork/go-utils/log/apex"
	apex "github.com/apex/log"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
)

const (
	clientName = "ttn-lw-migrate"
)

func NewConfig() *Config {
	config := &Config{
		sdkConfig: ttnsdk.NewCommunityConfig(clientName),
		flags:     &pflag.FlagSet{},
	}

	config.flags.StringVar(&config.frequencyPlanID,
		"frequency-plan-id",
		"",
		"Frequency Plan ID of exported devices")
	config.flags.StringVar(&config.appID,
		"app-id",
		"",
		"TTN Application ID")
	config.flags.StringVar(&config.appAccessKey,
		"app-access-key",
		"",
		"TTN Application Access Key (with 'devices' permissions")
	config.flags.StringVar(&config.caCert,
		"ca-cert",
		"",
		"(only for private networks)")
	config.flags.StringVar(&config.sdkConfig.HandlerAddress,
		"handler-address",
		"",
		"(only for private networks) Address for the Handler")
	config.flags.StringVar(&config.sdkConfig.AccountServerAddress,
		"account-server-address",
		"",
		"(only for private networks) Address for the Account Server")
	config.flags.StringVar(&config.sdkConfig.AccountServerClientID,
		"account-server-client-id",
		"",
		"(only for private networks) Client ID for the Account Server")
	config.flags.StringVar(&config.sdkConfig.AccountServerClientSecret,
		"account-server-client-secret",
		"",
		"(only for private networks) Client secret for the Account Server")
	config.flags.StringVar(&config.sdkConfig.DiscoveryServerAddress,
		"discovery-server-address",
		"",
		"(only for private networks) Address for the Discovery Server")
	config.flags.BoolVar(&config.sdkConfig.DiscoveryServerInsecure,
		"discovery-server-insecure",
		false,
		"(only for private networks) Not recommended")
	config.flags.BoolVar(&config.withSession,
		"with-session",
		true,
		"Export device session keys and frame counters")
	config.flags.BoolVar(&config.resetsToFrequencyPlan,
		"resets-to-frequency-plan",
		false,
		"Configure preset frequencies for ABP devices so that they match the used Frequency Plan")

	return config
}

type Config struct {
	sdkConfig ttnsdk.ClientConfig

	caCert       string
	appAccessKey string
	appID        string

	frequencyPlanID string

	withSession           bool
	dryRun                bool
	resetsToFrequencyPlan bool

	flags   *pflag.FlagSet
	fpStore *frequencyplans.Store
}

func (c *Config) Initialize(rootConfig source.Config) error {
	if frequencyPlanID := os.Getenv("FREQUENCY_PLAN_ID"); frequencyPlanID != "" {
		c.frequencyPlanID = frequencyPlanID
	}
	if appID := os.Getenv("TTNV2_APP_ID"); appID != "" {
		c.appID = appID
	}
	if appAccessKey := os.Getenv("TTNV2_APP_ACCESS_KEY"); appAccessKey != "" {
		c.appAccessKey = appAccessKey
	}
	if caCert := os.Getenv("TTNV2_CA_CERT"); caCert != "" {
		c.caCert = caCert
	}
	if handlerAddress := os.Getenv("TTNV2_HANDLER_ADDRESS"); handlerAddress != "" {
		c.sdkConfig.HandlerAddress = handlerAddress
	}
	if accountServerAddress := os.Getenv("TTNV2_ACCOUNT_SERVER_ADDRESS"); accountServerAddress != "" {
		c.sdkConfig.AccountServerAddress = accountServerAddress
	}
	if accountServerClientID := os.Getenv("TTNV2_ACCOUNT_SERVER_CLIENT_ID"); accountServerClientID != "" {
		c.sdkConfig.AccountServerClientID = accountServerClientID
	}
	if accountServerClientSecret := os.Getenv("TTNV2_ACCOUNT_SERVER_CLIENT_SECRET"); accountServerClientSecret != "" {
		c.sdkConfig.AccountServerClientSecret = accountServerClientSecret
	}
	if discoveryServerAddress := os.Getenv("TTNV2_DISCOVERY_SERVER_ADDRESS"); discoveryServerAddress != "" {
		c.sdkConfig.DiscoveryServerAddress = discoveryServerAddress
	}
	if discoveryServerInsecure := os.Getenv("TTNV2_DISCOVERY_SERVER_INSECURE"); discoveryServerInsecure == "true" {
		c.sdkConfig.DiscoveryServerInsecure = true
	}
	if withSession := os.Getenv("TTNV2_WITH_SESSION"); withSession == "true" {
		c.withSession = true
	}
	if resetsToFrequencyPlan := os.Getenv("TTNV2_RESETS_TO_FREQUENCY_PLAN"); resetsToFrequencyPlan == "true" {
		c.resetsToFrequencyPlan = true
	}

	if c.caCert != "" {
		if c.sdkConfig.TLSConfig == nil {
			c.sdkConfig.TLSConfig = new(tls.Config)
		}
		rootCAs := c.sdkConfig.TLSConfig.RootCAs
		if rootCAs == nil {
			var err error
			if rootCAs, err = x509.SystemCertPool(); err != nil {
				rootCAs = x509.NewCertPool()
			}
		}
		pemBytes, err := ioutil.ReadFile(c.caCert)
		if err != nil {
			return err
		}
		rootCAs.AppendCertsFromPEM(pemBytes)
	}

	if c.appID == "" {
		return errNoAppID.New()
	}
	if c.appAccessKey == "" {
		return errNoAppAccessKey.New()
	}
	if c.frequencyPlanID == "" {
		return errNoFrequencyPlanID.New()
	}

	logLevel := ttnapex.InfoLevel
	if rootConfig.Verbose {
		logLevel = ttnapex.DebugLevel
	}
	logger := ttnapex.Wrap(&apex.Logger{
		Level:   logLevel,
		Handler: cli.New(os.Stderr),
	})
	ttnlog.Set(logger)

	fpFetcher, err := fetch.FromHTTP(nil, rootConfig.FrequencyPlansURL)
	if err != nil {
		return err
	}
	c.fpStore = frequencyplans.NewStore(fpFetcher)

	c.dryRun = rootConfig.DryRun

	return nil
}

// Flags returns the flags for the configuration.
func (c *Config) Flags() *pflag.FlagSet {
	return c.flags
}
