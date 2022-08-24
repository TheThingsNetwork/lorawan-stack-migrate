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

	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	ttntypes "github.com/TheThingsNetwork/ttn/core/types"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Source implements the Source interface.
type Source struct {
	ctx context.Context

	config config
	mgr    ttnsdk.DeviceManager
	client ttnsdk.Client
}

// NewSource creates a new TTNv2 Source.
func NewSource(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) {
	config, err := getConfig(flags)
	if err != nil {
		return nil, err
	}

	s := &Source{
		ctx:    ctx,
		config: config,
		client: config.sdkConfig.NewClient(config.appID, config.appAccessKey),
	}
	mgr, err := s.client.ManageDevices()
	if err != nil {
		return nil, err
	}
	s.mgr = newDeviceManager(ctx, mgr)
	return s, nil
}

// ExportDevice implements the source.Source interface.
func (s *Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) {
	dev, err := s.mgr.Get(devID)
	if err != nil {
		if err, ok := errors.From(err); ok {
			return nil, err
		}
		return nil, err
	}

	v3dev := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{},
		},
	}
	v3dev.Ids.DeviceId = dev.DevID
	v3dev.Ids.JoinEui = dev.AppEUI.Bytes()
	v3dev.Ids.DevEui = dev.DevEUI.Bytes()
	v3dev.Ids.ApplicationIds.ApplicationId = s.config.appID

	v3dev.Name = dev.DevID
	v3dev.Description = dev.Description
	v3dev.Attributes = dev.Attributes

	v3dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_2
	v3dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2_REV_B
	v3dev.FrequencyPlanId = s.config.frequencyPlanID

	v3dev.MacSettings = &ttnpb.MACSettings{
		StatusTimePeriodicity:  pbtypes.DurationProto(0),
		StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
	}
	if dev.Uses32BitFCnt {
		v3dev.MacSettings.Supports_32BitFCnt = &ttnpb.BoolValue{
			Value: dev.Uses32BitFCnt,
		}
	}
	if dev.DisableFCntCheck {
		v3dev.MacSettings.ResetsFCnt = &ttnpb.BoolValue{
			Value: dev.DisableFCntCheck,
		}
	}

	deviceSupportsJoin := dev.AppKey != nil && !dev.AppKey.IsEmpty()
	deviceHasSession := dev.DevAddr != nil && !dev.DevAddr.IsEmpty() && dev.NwkSKey != nil && !dev.NwkSKey.IsEmpty() && dev.AppSKey != nil && !dev.AppSKey.IsEmpty()
	if deviceSupportsJoin {
		// OTAA devices
		v3dev.SupportsJoin = true
		v3dev.RootKeys = &ttnpb.RootKeys{
			AppKey: &ttnpb.KeyEnvelope{
				Key: dev.AppKey.Bytes(),
			},
		}
	}

	if dev.Latitude != 0 || dev.Longitude != 0 {
		v3dev.Locations = map[string]*ttnpb.Location{
			"user": {
				Latitude:  float64(dev.Latitude),
				Longitude: float64(dev.Longitude),
				Altitude:  dev.Altitude,
				Source:    ttnpb.LocationSource_SOURCE_REGISTRY,
			},
		}
	}

	if s.config.withSession && deviceHasSession || !deviceSupportsJoin {
		v3dev.Session = &ttnpb.Session{
			Keys: &ttnpb.SessionKeys{
				AppSKey:     &ttnpb.KeyEnvelope{},
				FNwkSIntKey: &ttnpb.KeyEnvelope{},
			},
			LastFCntUp:    dev.FCntUp,
			LastNFCntDown: dev.FCntDown,
			StartedAt:     pbtypes.TimestampNow(),
		}
		if deviceSupportsJoin {
			v3dev.Session.Keys.SessionKeyId = generateBytes(16)
		}
		v3dev.Session.DevAddr = dev.DevAddr.Bytes()
		v3dev.Session.Keys.AppSKey.Key = dev.AppSKey.Bytes()
		v3dev.Session.Keys.FNwkSIntKey.Key = dev.NwkSKey.Bytes()

		if v3dev.MacState, err = mac.NewState(v3dev, s.config.fpStore, &ttnpb.MACSettings{}); err != nil {
			return nil, err
		}
		// Ensure MAC state matches v2 configuration.
		v3dev.MacState.CurrentParameters = v3dev.MacState.DesiredParameters
		v3dev.MacState.DeviceClass = ttnpb.Class_CLASS_A
		v3dev.MacState.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_2
		v3dev.MacState.CurrentParameters.Rx1Delay = ttnpb.RxDelay_RX_DELAY_1
	}

	log.FromContext(s.ctx).WithFields(log.Fields(
		"device_id", dev.DevID,
		"dev_eui", dev.DevEUI,
	)).Info("Clearing device keys")
	if !s.config.dryRun {
		dev.AppKey = &ttntypes.AppKey{}
		if s.config.withSession {
			dev.AppSKey = &ttntypes.AppSKey{}
			dev.NwkSKey = &ttntypes.NwkSKey{}
			dev.DevAddr = &ttntypes.DevAddr{}
		}
		if err := s.mgr.Set(dev); err != nil {
			return nil, err
		}
	}

	// For OTAA devices with a session, set current parameters instead of MAC settings.
	if !deviceSupportsJoin {
		v3dev.MacSettings.Rx1Delay = &ttnpb.RxDelayValue{Value: ttnpb.RxDelay_RX_DELAY_1}

		if s.config.resetsToFrequencyPlan {
			macState, err := mac.NewState(v3dev, s.config.fpStore, &ttnpb.MACSettings{})
			if err != nil {
				return nil, err
			}
			channels := macState.DesiredParameters.GetChannels()
			freqs := make([]uint64, 0, len(channels))
			for _, channel := range channels {
				if channel.EnableUplink && channel.UplinkFrequency > 0 {
					freqs = append(freqs, channel.UplinkFrequency)
				}
			}

			v3dev.MacSettings.FactoryPresetFrequencies = freqs
		}
	}

	return v3dev, nil
}

// RangeDevices implements the source.Source interface.
func (s *Source) RangeDevices(_ string, f func(source.Source, string) error) error {
	devices, err := s.mgr.List(0, 0)
	if err != nil {
		if err, ok := errors.From(err); ok {
			return err
		}
		return err
	}

	for _, dev := range devices {
		if err := f(s, dev.DevID); err != nil {
			return err
		}
	}
	return nil
}

// Close implements the Source interface.
func (s *Source) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func generateBytes(length int) []byte {
	return random.Bytes(length)
}
