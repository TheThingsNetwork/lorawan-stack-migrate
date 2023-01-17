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
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func New() (*Config, *pflag.FlagSet) {
	var (
		config = &Config{}
		flags  = &pflag.FlagSet{}
	)

	flags.StringVar(&config.url,
		"api-url",
		os.Getenv("CHIRPSTACK_API_URL"),
		"ChirpStack API URL")
	flags.StringVar(&config.token,
		"api-token",
		os.Getenv("CHIRPSTACK_API_TOKEN"),
		"ChirpStack API Token")
	flags.StringVar(&config.caPath,
		"api-ca",
		os.Getenv("CHIRPSTACK_API_CA"),
		"(optional) CA for TLS")
	flags.BoolVar(&config.insecure,
		"api-insecure",
		os.Getenv("CHIRPSTACK_API_INSECURE") == "1",
		"Do not connect to ChirpStack over TLS")
	flags.BoolVar(&config.ExportVars,
		"export-vars",
		false,
		"Export device variables from ChirpStack")
	flags.BoolVar(&config.ExportSession,
		"export-session",
		true,
		"Export device session keys from ChirpStack")
	flags.StringVar(&config.joinEUI,
		"join-eui",
		os.Getenv("JOIN_EUI"),
		"JoinEUI of exported devices")
	flags.StringVar(&config.FrequencyPlanID,
		"frequency-plan-id",
		os.Getenv("FREQUENCY_PLAN_ID"),
		"Frequency Plan ID of exported devices")

	return config, flags
}

type Config struct {
	source.Config

	ClientConn *grpc.ClientConn

	token, caPath, url,
	FrequencyPlanID string

	joinEUI string
	JoinEUI *types.EUI64

	insecure,
	ExportVars,
	ExportSession bool
}

func (c *Config) Initialize() error {
	if c.token == "" {
		return errNoAPIToken.New()
	}
	if c.url == "" {
		return errNoAPIURL.New()
	}
	if c.FrequencyPlanID == "" {
		return errNoFrequencyPlan.New()
	}

	c.JoinEUI = &types.EUI64{}
	if err := c.JoinEUI.UnmarshalText([]byte(c.joinEUI)); err != nil {
		return errInvalidJoinEUI.WithAttributes("join_eui", c.joinEUI)
	}

	if !c.insecure || c.caPath != "" {
		if err := setCustomCA(c.caPath); err != nil {
			return err
		}
	}

	err := c.dialGRPC(
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(token(c.token)),
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) dialGRPC(opts ...grpc.DialOption) error {
	if c.insecure && c.caPath == "" {
		opts = append(opts, grpc.WithInsecure())
	}
	if tls := http.DefaultTransport.(*http.Transport).TLSClientConfig; tls != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tls)))
	}
	var err error
	c.ClientConn, err = grpc.Dial(c.url, opts...)
	if err != nil {
		return err
	}
	return nil
}

func setCustomCA(path string) error {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	rootCAs := http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs
	if rootCAs == nil {
		if rootCAs, err = x509.SystemCertPool(); err != nil {
			rootCAs = x509.NewCertPool()
		}
	}
	rootCAs.AppendCertsFromPEM(pemBytes)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: rootCAs}
	return nil
}
