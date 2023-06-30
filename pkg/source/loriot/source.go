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

package loriot

import (
	"context"
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/loriot/api"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/loriot/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

type Source struct {
	Config *config.Config
}

func createNewSource(cfg *config.Config) source.CreateSource {
	return func(ctx context.Context, rootCfg source.Config) (source.Source, error) {
		return Source{}, nil
	}
}

func unmarshalTextToBytes(
	unmarshaller interface {
		UnmarshalText([]byte) error
		Bytes() []byte
	},
	source string,
) ([]byte, error) {
	err := unmarshaller.UnmarshalText([]byte(source))
	return unmarshaller.Bytes(), err
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) {
	if s.Config.AppID == "" {
		return nil, errors.New("no app id")
	}

	dev := new(ttnpb.EndDevice)
	dev.Attributes = make(map[string]string)
	dev.Formatters = &ttnpb.MessagePayloadFormatters{}
	dev.Ids = &ttnpb.EndDeviceIdentifiers{ApplicationIds: &ttnpb.ApplicationIdentifiers{}}
	dev.MacSettings = &ttnpb.MACSettings{}
	dev.RootKeys = &ttnpb.RootKeys{}

	lDev, err := api.GetDevice(s.Config.AppID, devID)
	if err != nil {
		return nil, err
	}

	// Identifiers
	dev.Ids.ApplicationIds.ApplicationId = s.Config.AppID
	dev.Ids.DevEui, err = unmarshalTextToBytes(&types.EUI64{}, devID)
	if err != nil {
		return nil, err
	}
	dev.Ids.JoinEui, err = unmarshalTextToBytes(&types.EUI64{}, lDev.Appeui)
	if err != nil {
		return nil, err
	}

	// Information
	dev.Name = fmt.Sprintf("eui-%s", devID)

	// Frequency Plan
	dev.FrequencyPlanId = fmt.Sprint(lDev.Freq) // TODO: parse this properly

	// General
	switch v := fmt.Sprintf("%d.%d.%s", lDev.Lorawan.Major, lDev.Lorawan.Minor, strings.ToLower(lDev.Lorawan.Revision)); v {
	case "1.0.", "1.0.0":
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0
		dev.LorawanPhyVersion = ttnpb.PHYVersion_TS001_V1_0

	case "1.0.1":
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_1
		dev.LorawanPhyVersion = ttnpb.PHYVersion_TS001_V1_0_1

	case "1.0.2", "1.0.2a":
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_2
		dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2

	case "1.0.2b":
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_2
		dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2_REV_B

	case "1.0.3":
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_3
		dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_3_REV_A

	case "1.1.0", "1.1.0a":
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_1
		dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_1_REV_A

	case "1.1.0b":
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_1
		dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_1_REV_B

	default:
		return nil, errors.New("invalid lorawan version {version}").WithAttributes("version", v)
	}

	// Join (OTAA/ABP)
	switch strings.ToLower(lDev.Devclass) {
	case "a":

	case "b":
		dev.SupportsClassB = true

	case "c":
		dev.SupportsClassC = true
	}

	return dev, nil
}

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(appID string, f func(source.Source, string) error) error {
	page := 1
	for {
		p, err := api.GetPaginatedDevices(appID, page)
		if err != nil {
			return err
		}

		for _, d := range p.Devices {
			if err := f(s, d.ID); err != nil {
				return err
			}
		}

		if page*p.PerPage >= p.Total {
			break
		}
		page++
	}
	return nil
}

// Close implements the Source interface.
func (s Source) Close() error { return nil }
