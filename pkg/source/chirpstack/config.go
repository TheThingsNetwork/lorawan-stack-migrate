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

package chirpstack

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

type config struct {
	ctx context.Context

	token    string
	url      string
	ca       string
	insecure bool
	tls      *tls.Config

	frequencyPlanID string
	joinEUI         *types.EUI64
	exportVars      bool
	exportSession   bool
}

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.String("chirpstack.api-url", os.Getenv("CHIRPSTACK_API_URL"), "ChirpStack API URL")
	flags.String("chirpstack.api-token", os.Getenv("CHIRPSTACK_API_TOKEN"), "ChirpStack API Token")
	flags.String("chirpstack.api-ca", os.Getenv("CHIRPSTACK_API_CA"), "(optional) CA for TLS")
	flags.Bool("chirpstack.api-insecure", os.Getenv("CHIRPSTACK_API_INSECURE") == "1", "Do not connect to ChirpStack over TLS")
	flags.Bool("chirpstack.export-vars", false, "Export device variables from ChirpStack")
	flags.Bool("chirpstack.export-session", true, "Export device session keys from ChirpStack")
	flags.String("chirpstack.join-eui", os.Getenv("JOIN_EUI"), "JoinEUI of exported devices")
	flags.String("chirpstack.frequency-plan-id", os.Getenv("FREQUENCY_PLAN_ID"), "Frequency Plan ID of exported devices")
	return flags
}

func buildConfig(ctx context.Context, flags *pflag.FlagSet) (config, error) {
	stringFlag := func(f string) string {
		s, _ := flags.GetString(f)
		return s
	}
	boolFlag := func(f string) bool {
		s, _ := flags.GetBool(f)
		return s
	}

	c := config{
		ctx: ctx,

		token:    stringFlag("chirpstack.api-token"),
		url:      stringFlag("chirpstack.api-url"),
		ca:       stringFlag("chirpstack.api-ca"),
		insecure: boolFlag("chirpstack.api-insecure"),

		frequencyPlanID: stringFlag("chirpstack.frequency-plan-id"),
		exportVars:      boolFlag("chirpstack.export-vars"),
		exportSession:   boolFlag("chirpstack.export-session"),
	}

	if c.token == "" {
		return config{}, errNoAPIToken.New()
	}
	if c.url == "" {
		return config{}, errNoAPIURL.New()
	}
	if c.frequencyPlanID == "" {
		return config{}, errNoFrequencyPlan.New()
	}

	c.joinEUI = &types.EUI64{}
	strJoinEUI := stringFlag("chirpstack.join-eui")
	if err := c.joinEUI.UnmarshalText([]byte(strJoinEUI)); err != nil {
		return config{}, errInvalidJoinEUI.WithAttributes("join_eui", strJoinEUI)
	}

	if !c.insecure || c.ca != "" {
		c.tls = &tls.Config{}
		rootCAs := c.tls.RootCAs
		if rootCAs == nil {
			var err error
			if rootCAs, err = x509.SystemCertPool(); err != nil {
				rootCAs = x509.NewCertPool()
			}
		}
		if c.ca != "" {
			pemBytes, err := ioutil.ReadFile(c.ca)
			if err != nil {
				return config{}, errRead.WithAttributes("file", c.ca)
			}
			rootCAs.AppendCertsFromPEM(pemBytes)
		}
	}

	return c, nil
}
