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
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const dialTimeout = 10 * time.Second

type Config struct {
	src source.Config

	apiKey, caCertPath, url, joinEUI string
	flags                            *pflag.FlagSet
	FPStore                          *frequencyplans.Store
	insecure                         bool

	ClientConn *grpc.ClientConn

	ExportVars,
	ExportSession bool
	FrequencyPlanID string
	JoinEUI         *types.EUI64
}

func New() *Config {
	config := &Config{
		flags: &pflag.FlagSet{},
	}

	config.flags.StringVar(&config.url,
		"api-url",
		"",
		"ChirpStack API URL")
	config.flags.StringVar(&config.apiKey,
		"api-key",
		"",
		"ChirpStack API key")
	config.flags.StringVar(&config.caCertPath,
		"ca-cert-path",
		"",
		"(optional) Path to the CA certificate file for ChirpStack API TLS connections")
	config.flags.BoolVar(&config.insecure,
		"insecure",
		false,
		"Do not connect to ChirpStack over TLS")
	config.flags.BoolVar(&config.ExportVars,
		"export-vars",
		false,
		"Export device variables from ChirpStack")
	config.flags.BoolVar(&config.ExportSession,
		"export-session",
		false,
		"Export device session keys from ChirpStack")
	config.flags.StringVar(&config.joinEUI,
		"join-eui",
		"",
		"JoinEUI of exported devices")
	config.flags.StringVar(&config.FrequencyPlanID,
		"frequency-plan-id",
		"",
		"Frequency Plan ID of exported devices")

	return config
}

func (c *Config) Initialize(src source.Config) error {
	c.src = src

	if c.apiKey = os.Getenv("CHIRPSTACK_API_KEY"); c.apiKey == "" {
		return errNoAPIToken.New()
	}
	if c.url = os.Getenv("CHIRPSTACK_API_URL"); c.url == "" {
		return errNoAPIURL.New()
	}
	if c.FrequencyPlanID = os.Getenv("FREQUENCY_PLAN_ID"); c.FrequencyPlanID == "" {
		return errNoFrequencyPlan.New()
	}
	if c.joinEUI = os.Getenv("JOIN_EUI"); c.joinEUI == "" {
		return errNoJoinEUI.New()
	}
	c.JoinEUI = &types.EUI64{}
	if err := c.JoinEUI.UnmarshalText([]byte(c.joinEUI)); err != nil {
		return errInvalidJoinEUI.WithAttributes("join_eui", c.joinEUI)
	}
	c.caCertPath = os.Getenv("CHIRPSTACK_CA_CERT_PATH")
	c.insecure = os.Getenv("CHIRPSTACK_INSECURE") == "true"
	c.ExportSession = os.Getenv("EXPORT_SESSION") == "true"
	c.ExportVars = os.Getenv("EXPORT_VARS") == "true"

	err := c.dialGRPC(
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(token(c.apiKey)),
	)
	if err != nil {
		return err
	}
	fpFetcher, err := fetch.FromHTTP(http.DefaultClient, src.FrequencyPlansURL)
	if err != nil {
		return err
	}
	c.FPStore = frequencyplans.NewStore(fpFetcher)
	return nil
}

// Flags returns the flags for the configuration.
func (c *Config) Flags() *pflag.FlagSet {
	return c.flags
}

func (c *Config) dialGRPC(opts ...grpc.DialOption) error {
	if c.insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig, err := generateTLSConfig(c.caCertPath)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	var err error
	c.ClientConn, err = grpc.DialContext(ctx, c.url, opts...)
	if err != nil {
		return err
	}
	return nil
}

// GenerateTLSConfig generates a TLS configuration.
func generateTLSConfig(caPath string) (cfg *tls.Config, err error) {
	cfg = http.DefaultTransport.(*http.Transport).TLSClientConfig
	if cfg == nil {
		cfg = &tls.Config{}
	}
	if cfg.RootCAs == nil {
		if cfg.RootCAs, err = x509.SystemCertPool(); err != nil {
			cfg.RootCAs = x509.NewCertPool()
		}
	}
	if caPath == "" {
		return cfg, nil
	}
	pemBytes, err := os.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	cfg.RootCAs.AppendCertsFromPEM(pemBytes)
	return cfg, nil
}
