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
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
)

const (
	clientName = "ttn-lw-migrate"
)

type config struct {
	sdkConfig ttnsdk.ClientConfig

	appAccessKey string
	appID        string

	frequencyPlanID string

	withSession bool
	dryRun      bool

	fpStore *frequencyplans.Store
}

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.String("ttnv2.frequency-plan-id", os.Getenv("FREQUENCY_PLAN_ID"), "Frequency Plan ID of exported devices")
	flags.String("ttnv2.app-id", os.Getenv("TTNV2_APP_ID"), "TTN Application ID")
	flags.String("ttnv2.app-access-key", os.Getenv("TTNV2_APP_ACCESS_KEY"), "TTN Application Access Key (with 'devices' permissions)")
	flags.String("ttnv2.ca-cert", os.Getenv("TTNV2_CA_CERT"), "(only for private networks) CA for TLS")
	flags.String("ttnv2.handler-address", os.Getenv("TTNV2_HANDLER_ADDRESS"), "(only for private networks) Address for the Handler")
	flags.String("ttnv2.account-server-address", os.Getenv("TTNV2_ACCOUNT_SERVER_ADDRESS"), "(only for private networks) Address for the Account Server")
	flags.String("ttnv2.account-server-client-id", os.Getenv("TTNV2_ACCOUNT_SERVER_CLIENT_ID"), "(only for private networks) Client ID for the Account Server")
	flags.String("ttnv2.account-server-client-secret", os.Getenv("TTNV2_ACCOUNT_SERVER_CLIENT_SECRET"), "(only for private networks) Client secret for the Account Server")
	flags.String("ttnv2.discovery-server-address", os.Getenv("TTNV2_DISCOVERY_SERVER_ADDRESS"), "(only for private networks) Address for the Discovery Server")
	flags.Bool("ttnv2.discovery-server-insecure", false, "(only for private networks) Not recommended")
	flags.Bool("ttnv2.with-session", true, "Export device session keys and frame counters")

	return flags
}

func getConfig(flags *pflag.FlagSet) (config, error) {
	stringFlag := func(f string) string {
		s, _ := flags.GetString(f)
		return s
	}
	boolFlag := func(f string) bool {
		s, _ := flags.GetBool(f)
		return s
	}

	cfg := ttnsdk.NewCommunityConfig(clientName)
	if f := stringFlag("ttnv2.account-server-address"); f != "" {
		cfg.AccountServerAddress = f
	}
	if f := stringFlag("ttnv2.account-server-client-id"); f != "" {
		cfg.AccountServerClientID = f
	}
	if f := stringFlag("ttnv2.account-server-client-secret"); f != "" {
		cfg.AccountServerClientSecret = f
	}
	if f := stringFlag("ttnv2.handler-address"); f != "" {
		cfg.HandlerAddress = f
	}
	if f := stringFlag("ttnv2.discovery-server-address"); f != "" {
		cfg.DiscoveryServerAddress = f
	}
	cfg.DiscoveryServerInsecure = boolFlag("ttnv2.discovery-server-insecure")

	if ca := stringFlag("ttnv2.ca-cert"); ca != "" {
		if cfg.TLSConfig == nil {
			cfg.TLSConfig = &tls.Config{}
		}
		rootCAs := cfg.TLSConfig.RootCAs
		if rootCAs == nil {
			var err error
			if rootCAs, err = x509.SystemCertPool(); err != nil {
				rootCAs = x509.NewCertPool()
			}
		}
		pemBytes, err := ioutil.ReadFile(ca)
		if err != nil {
			return config{}, errRead.WithAttributes("file", ca)
		}
		rootCAs.AppendCertsFromPEM(pemBytes)
	}

	appAccessKey := stringFlag("ttnv2.app-access-key")
	if appAccessKey == "" {
		return config{}, errNoAppAccessKey.New()
	}
	appID := stringFlag("ttnv2.app-id")
	if appID == "" {
		return config{}, errNoAppID.New()
	}
	frequencyPlanID := stringFlag("ttnv2.frequency-plan-id")
	if frequencyPlanID == "" {
		return config{}, errNoFrequencyPlanID.New()
	}

	logLevel := ttnapex.InfoLevel
	if boolFlag("verbose") {
		logLevel = ttnapex.DebugLevel
	}
	logger := ttnapex.Wrap(&apex.Logger{
		Level:   logLevel,
		Handler: cli.New(os.Stderr),
	})
	ttnlog.Set(logger)

	fpFetcher, err := fetch.FromHTTP(nil, stringFlag("frequency-plans-url"), true)
	if err != nil {
		return config{}, err
	}

	return config{
		sdkConfig: cfg,

		appID:           appID,
		appAccessKey:    appAccessKey,
		frequencyPlanID: frequencyPlanID,

		withSession: boolFlag("ttnv2.with-session"),

		dryRun:  boolFlag("dry-run"),
		fpStore: frequencyplans.NewStore(fpFetcher),
	}, nil
}
