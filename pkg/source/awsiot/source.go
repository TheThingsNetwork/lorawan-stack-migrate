// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

package awsiot

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iotwireless"
	"github.com/aws/aws-sdk-go-v2/service/iotwireless/types"
	"go.thethings.network/lorawan-stack-migrate/pkg/iterator"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/awsiot/config"
	"go.thethings.network/lorawan-stack-migrate/pkg/util"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttntypes "go.thethings.network/lorawan-stack/v3/pkg/types"
)

// Source implements the Source interface.
type Source struct {
	ctx context.Context

	config *config.Config
}

func createNewSource(cfg *config.Config) source.CreateSource {
	return func(ctx context.Context, rootCfg source.Config) (source.Source, error) {
		if err := cfg.Initialize(rootCfg); err != nil {
			return nil, err
		}

		s := &Source{
			ctx:    ctx,
			config: cfg,
		}
		return s, nil
	}
}

func (s Source) getDevice(id string) (*DeviceIdentifiers, *Device, error) {
	resp, err := s.config.Client.GetWirelessDevice(s.ctx, &iotwireless.GetWirelessDeviceInput{
		IdentifierType: types.WirelessDeviceIdTypeWirelessDeviceId,
		Identifier:     aws.String(id),
	})
	if err != nil {
		return nil, nil, err
	}
	deviceIds := &DeviceIdentifiers{
		Id:          resp.Id,
		Name:        resp.Name,
		Description: resp.Description,
	}
	return deviceIds, &Device{resp.LoRaWAN}, nil
}

func (s Source) getDeviceProfile(id *string) (*Profile, error) {
	resp, err := s.config.Client.GetDeviceProfile(s.ctx, &iotwireless.GetDeviceProfileInput{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return &Profile{resp.LoRaWAN}, nil
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) {
	devIds, awsDev, err := s.getDevice(devID)
	if err != nil {
		return nil, err
	}

	endDev := &ttnpb.EndDevice{
		Name:        aws.ToString(devIds.Name),
		Description: aws.ToString(devIds.Description),
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: s.config.AppID},
			DeviceId:       aws.ToString(devIds.Id),
		},
		FrequencyPlanId: s.config.FrequencyPlanID,
		MacSettings:     &ttnpb.MACSettings{},
		RootKeys:        &ttnpb.RootKeys{},
	}
	endDev.Ids.DevEui, err = util.UnmarshalTextToBytes(&ttntypes.EUI64{}, aws.ToString(awsDev.DevEui))
	if err != nil {
		return nil, errInvalidDevEUI.WithAttributes("dev_eui", aws.ToString(awsDev.DevEui)).WithCause(err)
	}

	profile, err := s.getDeviceProfile(awsDev.DeviceProfileId)
	if err != nil {
		return nil, err
	}
	if err := profile.SetFields(endDev); err != nil {
		return nil, err
	}

	if endDev.SupportsJoin {
		if err := awsDev.SetOTAADevice(endDev); err != nil {
			return nil, err
		}
	} else {
		if err := awsDev.SetABPDevice(endDev); err != nil {
			return nil, err
		}
	}
	return endDev, nil
}

// Iterator implements source.Source.
func (s Source) Iterator(bool) iterator.Iterator {
	return iterator.NewReaderIterator(os.Stdin, '\n')
}

func (s Source) rangeThings(things []types.WirelessDeviceStatistics, f func(source.Source, string) error) error {
	for _, t := range things {
		id := aws.ToString(t.Id)
		if err := f(s, id); err != nil {
			return err
		}
	}
	return nil
}

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(appID string, f func(source.Source, string) error) error {
	resp, err := s.config.Client.ListWirelessDevices(s.ctx,
		&iotwireless.ListWirelessDevicesInput{
			WirelessDeviceType: types.WirelessDeviceTypeLoRaWAN,
			MaxResults:         100,
		})
	if err != nil {
		return err
	}
	s.rangeThings(resp.WirelessDeviceList, f)
	for resp.NextToken != nil {
		resp, err = s.config.Client.ListWirelessDevices(s.ctx,
			&iotwireless.ListWirelessDevicesInput{
				NextToken:  resp.NextToken,
				MaxResults: 100,
			})
		if err != nil {
			return err
		}
		s.rangeThings(resp.WirelessDeviceList, f)
	}
	return nil
}

// Close implements the Source interface.
func (s Source) Close() error {
	return nil
}
