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
	"go.uber.org/zap"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/ttnv3/api"
)

var logger *zap.SugaredLogger

type serverConfig struct {
	defaultGRPCAddress,
	ApplicationServerGRPCAddress,
	IdentityServerGRPCAddress,
	JoinServerGRPCAddress,
	NetworkServerGRPCAddress string
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

func New() (*Config, source.FlagSets) {
	var (
		config  = &Config{}
		devices = &pflag.FlagSet{}
		shared  = &pflag.FlagSet{}
	)

	devices.StringVar(&config.AppID,
		"app-id",
		os.Getenv("TTNV3_APP_ID"),
		"TTS Application ID")

	shared.StringVar(&config.appAPIKey,
		"app-api-key",
		os.Getenv("TTNV3_APP_API_KEY"),
		"TTS Application Access Key (with 'devices' permissions)")

	shared.StringVar(&config.caPath,
		"ca-file",
		os.Getenv("TTNV3_CA_FILE"),
		"TTS Path to a CA file (optional)")
	shared.BoolVar(&config.insecure,
		"insecure",
		false,
		"TTS allow TCP connection")

	shared.StringVar(&config.ServerConfig.defaultGRPCAddress,
		"default-grpc-address",
		os.Getenv("TTNV3_DEFAULT_GRPC_ADDRESS"),
		"TTS default GRPC Address (optional)")
	shared.StringVar(&config.ServerConfig.ApplicationServerGRPCAddress,
		"appplication-server-grpc-address",
		os.Getenv("TTNV3_APPLICATION_SERVER_GRPC_ADDRESS"),
		"TTS Application Server GRPC Address")
	shared.StringVar(&config.ServerConfig.IdentityServerGRPCAddress,
		"identity-server-grpc-address",
		os.Getenv("TTNV3_IDENTITY_SERVER_GRPC_ADDRESS"),
		"TTS Identity Server GRPC Address")
	shared.StringVar(&config.ServerConfig.JoinServerGRPCAddress,
		"join-server-grpc-address",
		os.Getenv("TTNV3_JOIN_SERVER_GRPC_ADDRESS"),
		"TTS Join Server GRPC Address")
	shared.StringVar(&config.ServerConfig.NetworkServerGRPCAddress,
		"network-server-grpc-address",
		os.Getenv("TTNV3_NETWORK_SERVER_GRPC_ADDRESS"),
		"TTS Network Server GRPC Address")

	shared.BoolVar(&config.NoSession,
		"no-session",
		false,
		"TTS export devices without session")
	shared.BoolVar(&config.DeleteSourceDevice,
		"delete-source-device",
		false,
		"TTS delete exported devices")

	return config, source.FlagSets{
		Devices: devices,
		Shared:  shared,
	}
}

type Config struct {
	source.RootConfig

	ServerConfig serverConfig

	caPath, appAPIKey,
	AppID string

	insecure,
	DeleteSourceDevice,
	DryRun, NoSession bool
}

func (c *Config) Initialize(rootConfig source.RootConfig) error {
	c.RootConfig = rootConfig

	var err error
	logger, err = NewLogger(c.Verbose)
	if err != nil {
		return err
	}

	if c.appAPIKey == "" {
		return errNoAppAPIKey.New()
	}
	api.SetAuth("bearer", c.appAPIKey)

	switch {
	case c.insecure:
		api.SetInsecure(true)
		logger.Warn("Using insecure connection to API")

	default:
		if c.caPath != "" {
			setCustomCA(c.caPath)
		}
	}

	c.ServerConfig.applyDefaults()
	if err := c.ServerConfig.anyFieldEmpty(); err != nil {
		return err
	}

	// deleteSourceDevice not allowed when dryRun
	if c.DryRun && c.DeleteSourceDevice {
		logger.Warn("Cannot delete source devices during a dry run.")
		c.DeleteSourceDevice = false
	}

	return nil
}

func NewLogger(verbose bool) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	if verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
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
