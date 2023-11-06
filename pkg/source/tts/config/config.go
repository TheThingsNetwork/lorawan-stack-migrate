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

package config

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/tts/api"
)

type serverConfig struct {
	defaultGRPCAddress,
	ApplicationServerGRPCAddress,
	IdentityServerGRPCAddress,
	JoinServerGRPCAddress,
	NetworkServerGRPCAddress string
}

func (c *serverConfig) anyFieldEmpty() error {
	if c.ApplicationServerGRPCAddress == "" {
		return errNoApplicationServerGRPCAddress.New()
	}
	if c.IdentityServerGRPCAddress == "" {
		return errNoIdentityServerGRPCAddress.New()
	}
	if c.JoinServerGRPCAddress == "" {
		return errNoJoinServerGRPCAddress.New()
	}
	if c.NetworkServerGRPCAddress == "" {
		return errNoNetworkServerGRPCAddress.New()
	}
	return nil
}

func (c *serverConfig) applyDefaults() {
	applyDefault := func(adresses ...*string) {
		for _, a := range adresses {
			if *a == "" {
				*a = c.defaultGRPCAddress
			}
		}
	}

	applyDefault(
		&c.ApplicationServerGRPCAddress,
		&c.IdentityServerGRPCAddress,
		&c.JoinServerGRPCAddress,
		&c.NetworkServerGRPCAddress,
	)
}

func New() (*Config, *pflag.FlagSet) {
	var (
		config = &Config{ServerConfig: &serverConfig{}}
		flags  = &pflag.FlagSet{}
	)

	flags.StringVar(&config.AppID,
		"app-id",
		os.Getenv("TTS_APP_ID"),
		"TTS Application ID")
	flags.StringVar(&config.appAPIKey,
		"app-api-key",
		os.Getenv("TTS_APP_API_KEY"),
		"TTS Application Access Key (with 'devices' permissions)")

	flags.StringVar(&config.caPath,
		"ca-file",
		os.Getenv("TTS_CA_FILE"),
		"TTS Path to a CA file (optional)")
	flags.BoolVar(&config.insecure,
		"insecure",
		false,
		"TTS allow TCP connection")

	flags.StringVar(&config.ServerConfig.defaultGRPCAddress,
		"default-grpc-address",
		os.Getenv("TTS_DEFAULT_GRPC_ADDRESS"),
		"TTS default GRPC Address (optional)")
	flags.StringVar(&config.ServerConfig.ApplicationServerGRPCAddress,
		"application-server-grpc-address",
		os.Getenv("TTS_APPLICATION_SERVER_GRPC_ADDRESS"),
		"TTS Application Server GRPC Address")
	flags.StringVar(&config.ServerConfig.IdentityServerGRPCAddress,
		"identity-server-grpc-address",
		os.Getenv("TTS_IDENTITY_SERVER_GRPC_ADDRESS"),
		"TTS Identity Server GRPC Address")
	flags.StringVar(&config.ServerConfig.JoinServerGRPCAddress,
		"join-server-grpc-address",
		os.Getenv("TTS_JOIN_SERVER_GRPC_ADDRESS"),
		"TTS Join Server GRPC Address")
	flags.StringVar(&config.ServerConfig.NetworkServerGRPCAddress,
		"network-server-grpc-address",
		os.Getenv("TTS_NETWORK_SERVER_GRPC_ADDRESS"),
		"TTS Network Server GRPC Address")

	flags.BoolVar(&config.NoSession,
		"no-session",
		false,
		"TTS export devices without session")
	flags.BoolVar(&config.DeleteSourceDevice,
		"delete-source-device",
		false,
		"TTS delete exported devices")

	flags.BoolVar(&config.ExportCACs,
		"export-cacs",
		false,
		"Export Claim Authentication Codes (CAC)",
	)

	return config, flags
}

type Config struct {
	source.Config

	ServerConfig *serverConfig

	insecure  bool
	caPath    string
	appAPIKey string

	ExportCACs         bool
	NoSession          bool
	DeleteSourceDevice bool
	AppID              string
}

func (c *Config) Initialize(rootConfig source.Config) error {
	c.Config = rootConfig

	if c.appAPIKey == "" {
		return errNoAppAPIKey.New()
	}
	api.SetAuth("bearer", c.appAPIKey)

	switch {
	case c.insecure:
		api.SetInsecure(true)
		c.Logger.Warn("Using insecure connection to API")

	default:
		if c.caPath != "" {
			setCustomCA(c.caPath)
		}
	}

	c.ServerConfig.applyDefaults()
	if err := c.ServerConfig.anyFieldEmpty(); err != nil {
		return err
	}

	// DeleteSourceDevice is not allowed during a dry run
	if c.DryRun && c.DeleteSourceDevice {
		c.Logger.Warn("Cannot delete source devices during a dry run.")
		c.DeleteSourceDevice = false
	}

	return nil
}

func setCustomCA(path string) error {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	cfg := http.DefaultTransport.(*http.Transport).TLSClientConfig
	switch {
	case cfg == nil:
		cfg = new(tls.Config)
		fallthrough

	case cfg.RootCAs == nil:
		if cfg.RootCAs, err = x509.SystemCertPool(); err != nil {
			cfg.RootCAs = x509.NewCertPool()
		}
	}
	cfg.RootCAs.AppendCertsFromPEM(pemBytes)
	if err = api.AddCA(pemBytes); err != nil {
		return err
	}
	return nil
}
