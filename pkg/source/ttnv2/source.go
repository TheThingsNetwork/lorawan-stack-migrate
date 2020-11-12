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

package ttnv2

import (
	"context"
	"time"

	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

const (
	// cooldown between consecutive DeviceManager.Get calls, in order to avoid rate limits.
	cooldown = 10 * time.Millisecond
)

// Source implements the Source interface.
type Source struct {
	ctx context.Context

	config config
	mgr    ttnsdk.DeviceManager
	client ttnsdk.Client

	devices map[string]*ttnsdk.Device
}

// NewSource creates a new TTNv2 Source.
func NewSource(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) {
	config, err := getConfig(ctx, flags)
	if err != nil {
		return nil, err
	}

	return &Source{
		ctx:     ctx,
		config:  config,
		devices: make(map[string]*ttnsdk.Device),
	}, nil
}

func (s *Source) getDeviceManager(appID string) (ttnsdk.DeviceManager, error) {
	if s.mgr == nil {
		if s.client == nil {
			s.client = s.config.sdkConfig.NewClient(appID, s.config.appAccessKey)
		}
		var err error
		s.mgr, err = s.client.ManageDevices()
		if err != nil {
			return nil, errors.FromGRPCError(err)
		}
	}
	return s.mgr, nil
}

// ExportDevice implements the source.Source interface.
func (s *Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) {
	dev := s.devices[devID]
	if s.config.withSession {
		mgr, err := s.getDeviceManager(s.config.appID)
		if err != nil {
			return nil, err
		}
		dev, err = mgr.Get(devID)
		if err != nil {
			return nil, errors.FromGRPCError(err)
		}
	}

	v3dev := &ttnpb.EndDevice{}
	v3dev.DeviceID = dev.DevID
	v3dev.ApplicationID = s.config.appID

	v3dev.Name = dev.DevID
	v3dev.Description = dev.Description
	v3dev.Attributes = dev.Attributes

	v3dev.JoinEUI = &types.EUI64{}
	if err := v3dev.JoinEUI.Unmarshal(dev.AppEUI.Bytes()); err != nil {
		return nil, err
	}
	v3dev.DevEUI = &types.EUI64{}
	if err := v3dev.DevEUI.Unmarshal(dev.DevEUI.Bytes()); err != nil {
		return nil, err
	}

	v3dev.LoRaWANVersion = ttnpb.MAC_V1_0_2
	v3dev.LoRaWANPHYVersion = ttnpb.PHY_V1_0_2_REV_B
	v3dev.FrequencyPlanID = s.config.frequencyPlanID

	v3dev.MACSettings = &ttnpb.MACSettings{
		Rx1Delay: &ttnpb.RxDelayValue{
			Value: ttnpb.RX_DELAY_1,
		},
	}
	if dev.Uses32BitFCnt {
		v3dev.MACSettings.Supports32BitFCnt = &pbtypes.BoolValue{
			Value: dev.Uses32BitFCnt,
		}
	}
	if dev.DisableFCntCheck {
		v3dev.MACSettings.ResetsFCnt = &pbtypes.BoolValue{
			Value: dev.DisableFCntCheck,
		}
	}

	if dev.AppKey != nil && !dev.AppKey.IsEmpty() {
		v3dev.SupportsJoin = true
		v3dev.RootKeys = &ttnpb.RootKeys{}
		v3dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{
			Key: &types.AES128Key{},
		}
		if err := v3dev.RootKeys.AppKey.Key.Unmarshal(dev.AppKey.Bytes()); err != nil {
			return nil, err
		}
	}

	if dev.Latitude != 0 || dev.Longitude != 0 {
		v3dev.Locations = map[string]*ttnpb.Location{
			"user": {
				Latitude:  float64(dev.Latitude),
				Longitude: float64(dev.Longitude),
				Altitude:  dev.Altitude,
				Source:    ttnpb.SOURCE_REGISTRY,
			},
		}
	}

	if s.config.withSession && dev.DevAddr != nil && dev.NwkSKey != nil && dev.AppSKey != nil {
		v3dev.Session = &ttnpb.Session{
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{},
				},
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{},
				},
			},
			LastFCntUp:    dev.FCntUp,
			LastNFCntDown: dev.FCntDown,
			StartedAt:     time.Now(),
		}
		if err := v3dev.Session.DevAddr.Unmarshal(dev.DevAddr.Bytes()); err != nil {
			return nil, err
		}
		if err := v3dev.Session.SessionKeys.AppSKey.Key.Unmarshal(dev.AppSKey.Bytes()); err != nil {
			return nil, err
		}
		if err := v3dev.Session.SessionKeys.FNwkSIntKey.Key.Unmarshal(dev.NwkSKey.Bytes()); err != nil {
			return nil, err
		}
	}

	return v3dev, nil
}

// RangeDevices implements the source.Source interface.
func (s *Source) RangeDevices(appID string, f func(source.Source, string) error) error {
	mgr, err := s.getDeviceManager(appID)
	if err != nil {
		return err
	}
	devices, err := mgr.List(0, 0)
	if err != nil {
		return errors.FromGRPCError(err)
	}

	for _, dev := range devices {
		s.devices[dev.DevID] = dev.AsDevice()
		if err := f(s, dev.DevID); err != nil {
			return err
		}
	}
	return nil
}

// Close implements the Source interface.
func (s *Source) Close() error {
	return s.client.Close()
}
