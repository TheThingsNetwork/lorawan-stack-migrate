// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package wanesy

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type Config struct {
	src source.Config

	appID           string
	frequencyPlanID string
	csvPath         string
	all             bool

	derivedMacVersion ttnpb.MACVersion
	derivedPhyVersion ttnpb.PHYVersion

	flags   *pflag.FlagSet
	fpStore *frequencyplans.Store
}

// NewConfig returns a new Wanesy configuration.
func NewConfig() *Config {
	config := &Config{
		flags: &pflag.FlagSet{},
	}
	config.flags.StringVar(&config.frequencyPlanID,
		"frequency-plan-id",
		"",
		"Frequency Plan ID for the exported devices")
	config.flags.StringVar(&config.appID,
		"app-id",
		"",
		"Application ID for the exported devices")
	config.flags.StringVar(&config.csvPath,
		"csv-path",
		"",
		"Path to the CSV file exported from Wanesy Management Center")
	config.flags.BoolVar(&config.all,
		"all",
		false,
		"Export all devices in the CSV. This is only used by the application command")
	return config
}

// Initialize the configuration.
func (c *Config) Initialize(src source.Config) error {
	c.src = src

	if c.appID = os.Getenv("APP_ID"); c.appID == "" {
		return errNoAppID.New()
	}
	if c.frequencyPlanID = os.Getenv("FREQUENCY_PLAN_ID"); c.frequencyPlanID == "" {
		return errNoFrequencyPlanID.New()
	}
	if c.csvPath = os.Getenv("CSV_PATH"); c.csvPath == "" {
		return errNoCSVFileProvided.New()
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

// ImportDevices imports the devices from the provided CSV file.
func ImportDevices(csvPath string) (Devices, error) {
	raw, err := os.ReadFile(csvPath)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(strings.NewReader(string(raw)))
	readValues, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(readValues) < 2 {
		return nil, errNoValuesInCSV.New()
	}
	values := make([]map[string]string, 0)
	for i := 1; i < len(readValues); i++ {
		keys := readValues[0]
		value := make(map[string]string)
		for j := 0; j < len(keys); j++ {
			noOfcolumns := len(readValues[i])
			if j >= noOfcolumns {
				value[keys[j]] = "" // Fill empty columns.
				continue
			}
			value[keys[j]] = readValues[i][j]
		}
		values = append(values, value)
	}
	j, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	devices := make(Devices)
	err = devices.UnmarshalJSON(j)
	if err != nil {
		return nil, err
	}
	return devices, nil
}
