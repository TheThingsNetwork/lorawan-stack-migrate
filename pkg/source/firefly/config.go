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

	flags *pflag.FlagSet
}

var logger *zap.SugaredLogger

// NewConfig returns a new Firefly configuration.
func NewConfig() *Config {
	config := &Config{
		flags: &pflag.FlagSet{},
	}
	config.flags.StringVar(&config.Host,
		"host",
		os.Getenv("FIREFLY_HOST"),
		"Host of the Firefly API. Don't use the scheme (http/https). Port is optional")
	config.flags.StringVar(&config.CACertPath,
		"ca-cert-path",
		os.Getenv("FIREFLY_CA_CERT_PATH"),
		"(optional) Path to the CA certificate for the Firefly API")
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
		`LoRaWAN MAC version for the exported devices.
Supported options are 1.0.0, 1.0.1, 1.0.2a, 1.0.2b, 1.0.3, 1.1.0a, 1.1.0b`)
	config.flags.StringVar(&config.appID,
		"app-id",
		os.Getenv("APP_ID"),
		"Application ID for the exported devices")
	config.flags.BoolVar(&config.invalidateKeys,
		"invalidate-keys",
		(os.Getenv("INVALIDATE_KEYS") == "true"),
		`Invalidate the root and/or session keys of the devices on the Firefly server.
This is necessary to prevent both networks from communicating with the same device.
The last byte of the keys will be incremented by 0x01. This enables an easy rollback if necessary.
Setting this flag to false would result in a dry run,
where the devices are exported but they are still valid on the firefly server
		`)
	config.flags.BoolVar(&config.UseHTTP,
		"use-http",
		(os.Getenv("FIREFLY_USE_HTTP") == "true"),
		"(optional) Use HTTP instead of HTTPS for the Firefly API. Only for testing")
	config.flags.BoolVar(&config.all,
		"all",
		(os.Getenv("ALL") == "true"),
		"Export all devices that the API key has access to. This is only used by the application command")
	return config
}

// Initialize the configuration.
func (c *Config) Initialize(src source.Config) error {
	c.src = src
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

	return nil
}

// Flags returns the flags for the configuration.
func (c *Config) Flags() *pflag.FlagSet {
	return c.flags
}
