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
	"os"

	"github.com/spf13/pflag"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

func New() (*Config, *pflag.FlagSet) {
	var (
		cfg   = new(Config)
		flags = new(pflag.FlagSet)
	)

	flags.StringVar(&cfg.APIKey,
		"api-key",
		os.Getenv("LORIOT_API_KEY"),
		"Loriot API Key")
	flags.StringVar(&cfg.URL, "api-url",
		os.Getenv("LORIOT_API_URL"),
		"Loriot API URL")
	flags.StringVar(&cfg.AppID,
		"app-id",
		os.Getenv("LORIOT_APP_ID"),
		"Loriot APP ID")
	flags.BoolVar(&cfg.Insecure,
		"insecure",
		os.Getenv("LORIOT_INSECURE") == "1",
		"Do not connect to Loriot over TLS")

	return cfg, flags
}

type Config struct {
	source.Config

	APIKey, URL string
	Insecure    bool

	AppID string
}
