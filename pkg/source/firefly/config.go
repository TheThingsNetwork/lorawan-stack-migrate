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
	"net/http"
	"os"

	"github.com/spf13/pflag"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/client"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type Config struct {
	client.Config
	src source.Config

	appID           string
	frequencyPlanID string
	joinEUI         string
	macVersion      string
	invalidateKeys  bool
	all             bool

	derivedMacVersion ttnpb.MACVersion
	derivedPhyVersion ttnpb.PHYVersion

	flags   *pflag.FlagSet
	fpStore *frequencyplans.Store
}

// NewConfig returns a new Firefly configuration.
func NewConfig() *Config {
	config := &Config{
		flags: &pflag.FlagSet{},
	}

	config.flags.StringVar(&config.Host,
		"host",
		"",
		"Host of the Firefly API. Don't use the scheme (http/https). Port is optional")
	config.flags.StringVar(&config.CACertPath,
		"ca-cert-path",
		"",
		"(optional) Path to the CA certificate for the Firefly API")
	config.flags.StringVar(&config.APIKey,
		"api-key",
		"",
		"Key to access the Firefly API")
	config.flags.StringVar(&config.joinEUI,
		"join-eui",
		"",
		"JoinEUI for the exported devices")
	config.flags.StringVar(&config.frequencyPlanID,
		"frequency-plan-id",
		"",
		"Frequency Plan ID for the exported devices")
	config.flags.StringVar(&config.macVersion,
		"mac-version",
		"",
		`LoRaWAN MAC version for the exported devices.
Supported options are 1.0.0, 1.0.1, 1.0.2a, 1.0.2b, 1.0.3, 1.1.0a, 1.1.0b`)
	config.flags.StringVar(&config.appID,
		"app-id",
		"",
		"Application ID for the exported devices")
	config.flags.BoolVar(&config.invalidateKeys,
		"invalidate-keys",
		false,
		`Invalidate the root and/or session keys of the devices on the Firefly server.
This is necessary to prevent both networks from communicating with the same device.
The last byte of the keys will be incremented by 0x01. This enables an easy rollback if necessary.
Setting this flag to false would result in a dry run,
where the devices are exported but they are still valid on the firefly server
		`)
	config.flags.BoolVar(&config.UseHTTP,
		"use-http",
		false,
		"(optional) Use HTTP instead of HTTPS for the Firefly API. Only for testing")
	config.flags.BoolVar(&config.all,
		"all",
		false,
		"Export all devices that the API key has access to. This is only used by the application command")
	return config
}

// Initialize the configuration.
func (c *Config) Initialize(src source.Config) error {
	c.src = src

	if appID := os.Getenv("APP_ID"); appID == "" {
		c.appID = appID
	}
	if frequencyPlanID := os.Getenv("FREQUENCY_PLAN_ID"); frequencyPlanID == "" {
		c.frequencyPlanID = frequencyPlanID
	}
	if joinEUI := os.Getenv("JOIN_EUI"); joinEUI == "" {
		c.joinEUI = joinEUI
	}
	if invalidateKeys := os.Getenv("INVALIDATE_KEYS"); invalidateKeys == "true" {
		c.invalidateKeys = true
	}
	if all := os.Getenv("ALL"); all == "true" {
		c.all = true
	}

	if host := os.Getenv("FIREFLY_HOST"); host == "" {
		c.Host = host
	}
	if apiKey := os.Getenv("FIREFLY_API_KEY"); apiKey == "" {
		c.APIKey = apiKey
	}
	if caCertPath := os.Getenv("FIREFLY_CA_CERT_PATH"); caCertPath == "" {
		c.CACertPath = caCertPath
	}
	if useHTTP := os.Getenv("FIREFLY_USE_HTTP"); useHTTP == "true" {
		c.UseHTTP = true
	}
	if macVersion := os.Getenv("MAC_VERSION"); macVersion == "" {
		c.macVersion = macVersion
	}

	if c.appID == "" {
		return errNoAppID.New()
	}
	if c.frequencyPlanID == "" {
		return errNoFrequencyPlanID.New()
	}
	if c.joinEUI == "" {
		return errNoJoinEUI.New()
	}

	if c.Host == "" {
		return errNoHost.New()
	}
	if c.APIKey == "" {
		return errNoAPIKey.New()
	}

	switch c.macVersion {
	case "1.0.0":
		c.derivedMacVersion = ttnpb.MACVersion_MAC_V1_0
		c.derivedPhyVersion = ttnpb.PHYVersion_TS001_V1_0
	case "1.0.1":
		c.derivedMacVersion = ttnpb.MACVersion_MAC_V1_0_1
		c.derivedPhyVersion = ttnpb.PHYVersion_TS001_V1_0_1
	case "1.0.2a":
		c.derivedMacVersion = ttnpb.MACVersion_MAC_V1_0_2
		c.derivedPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2
	case "1.0.2b":
		c.derivedMacVersion = ttnpb.MACVersion_MAC_V1_0_2
		c.derivedPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2_REV_B
	case "1.0.3":
		c.derivedMacVersion = ttnpb.MACVersion_MAC_V1_0_3
		c.derivedPhyVersion = ttnpb.PHYVersion_RP001_V1_0_3_REV_A
	case "1.1.0a":
		c.derivedMacVersion = ttnpb.MACVersion_MAC_V1_1
		c.derivedPhyVersion = ttnpb.PHYVersion_RP001_V1_1_REV_A
	case "1.1.0b":
		c.derivedMacVersion = ttnpb.MACVersion_MAC_V1_1
		c.derivedPhyVersion = ttnpb.PHYVersion_RP001_V1_1_REV_B
	default:
		return errInvalidMACVersion.WithAttributes("mac_version", c.macVersion)
	}

	fpFetcher, err := fetch.FromHTTP(http.DefaultClient, src.FrequencyPlansURL)
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
