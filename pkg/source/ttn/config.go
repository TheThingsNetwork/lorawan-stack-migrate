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

package ttn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/spf13/pflag"
)

const (
	clientName = "ttn-lw-migrate"
)

type config struct {
	sdkConfig ttnsdk.ClientConfig

	debug bool

	appAccessKey string
	appID        string

	frequencyPlanID   string
	withFrameCounters bool
}

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.Bool("ttn.without-frame-counters", false, "Do not export device frame counters (faster)")
	flags.String("ttn.frequency-plan-id", os.Getenv("FREQUENCY_PLAN_ID"), "Frequency Plan ID of exported devices")
	flags.String("ttn.app-id", os.Getenv("TTN_APP_ID"), "TTN Application ID")
	flags.String("ttn.app-access-key", os.Getenv("TTN_APP_ACCESS_KEY"), "TTN Application Access Key (with 'devices' permissions)")
	flags.String("ttn.ca-cert", os.Getenv("TTN_CA_CERT"), "(only for private networks) CA for TLS")
	flags.String("ttn.handler-address", os.Getenv("TTN_HANDLER_ADDRESS"), "(only for private networks) Address for the Handler")
	flags.String("ttn.account-server-address", os.Getenv("TTN_ACCOUNT_SERVER_ADDRESS"), "(only for private networks) Address for the Account Server")
	flags.String("ttn.account-server-client-id", os.Getenv("TTN_ACCOUNT_SERVER_CLIENT_ID"), "(only for private networks) Client ID for the Account Server")
	flags.String("ttn.account-server-client-secret", os.Getenv("TTN_ACCOUNT_SERVER_CLIENT_SECRET"), "(only for private networks) Client secret for the Account Server")
	flags.String("ttn.discovery-server-address", os.Getenv("TTN_DISCOVERY_SERVER_ADDRESS"), "(only for private networks) Address for the Discovery Server")
	flags.Bool("ttn.discovery-server-insecure", false, "(only for private networks) Not recommended")

	return flags
}

func getConfig(ctx context.Context, flags *pflag.FlagSet) (config, error) {
	stringFlag := func(f string) string {
		s, _ := flags.GetString(f)
		return s
	}
	boolFlag := func(f string) bool {
		s, _ := flags.GetBool(f)
		return s
	}

	cfg := ttnsdk.NewCommunityConfig(clientName)
	if f := stringFlag("ttn.account-server-address"); f != "" {
		cfg.AccountServerAddress = f
	}
	if f := stringFlag("ttn.account-server-client-id"); f != "" {
		cfg.AccountServerClientID = f
	}
	if f := stringFlag("ttn.account-server-client-secret"); f != "" {
		cfg.AccountServerClientSecret = f
	}
	if f := stringFlag("ttn.handler-address"); f != "" {
		cfg.HandlerAddress = f
	}
	if f := stringFlag("ttn.discovery-server-address"); f != "" {
		cfg.DiscoveryServerAddress = f
	}
	cfg.DiscoveryServerInsecure = boolFlag("ttn.discovery-server-insecure")

	if ca := stringFlag("ttn.ca-cert"); ca != "" {
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

	appAccessKey := stringFlag("ttn.app-access-key")
	if appAccessKey == "" {
		return config{}, errNoAppAccessKey.New()
	}
	appID := stringFlag("ttn.app-id")
	if appID == "" {
		return config{}, errNoAppID.New()
	}
	frequencyPlanID := stringFlag("ttn.frequency-plan-id")
	if frequencyPlanID == "" {
		return config{}, errNoFrequencyPlanID.New()
	}

	logger := apex.Stdout()
	if boolFlag("verbose") {
		logger.MustParseLevel("debug")
	}
	ttnlog.Set(logger)

	return config{
		sdkConfig: cfg,

		appID:           appID,
		appAccessKey:    appAccessKey,
		frequencyPlanID: frequencyPlanID,

		withFrameCounters: !boolFlag("ttn.without-frame-counters"),
	}, nil
}
