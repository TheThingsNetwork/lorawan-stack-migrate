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

	config *Config
	mgr    ttnsdk.DeviceManager
	client ttnsdk.Client
}

func createNewSource(cfg *Config) source.CreateSource {
	return func(ctx context.Context, rootCfg source.Config) (source.Source, error) {
		return NewSource(ctx, cfg, rootCfg)
	}
}

// NewSource creates a new TTNv2 Source.
func NewSource(ctx context.Context, cfg *Config, rootCfg source.Config) (source.Source, error) {
	s := &Source{
		ctx:    ctx,
		config: cfg,
		client: cfg.sdkConfig.NewClient(cfg.appID, cfg.appAccessKey),
	}
	mgr, err := s.client.ManageDevices()
	if err != nil {
		return nil, err
	}
	s.mgr = newDeviceManager(ctx, mgr)
	return s, cfg.Initialize(rootCfg)
}

// ExportDevice implements the source.Source interface.
func (s *Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) {
	if s.config.appID == "" {
		return nil, errNoAppID.New()
	}

	dev, err := s.mgr.Get(devID)
	if err != nil {
		if err, ok := errors.From(err); ok {
			return nil, err
		}
		return nil, err
	}

	v3dev := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: s.config.appID,
			},
			DeviceId: dev.DevID,
			DevEui:   dev.DevEUI.Bytes(),
			JoinEui:  dev.AppEUI.Bytes(),
		},
		MacSettings: &ttnpb.MACSettings{
			StatusTimePeriodicity:  pbtypes.DurationProto(0),
			StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
		},
		Name:              dev.DevID,
		Description:       dev.Description,
		Attributes:        dev.Attributes,
		LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_2,
		LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
		FrequencyPlanId:   s.config.frequencyPlanID,
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
			AppKey: &ttnpb.KeyEnvelope{Key: dev.AppKey.Bytes()},
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
				AppSKey:     &ttnpb.KeyEnvelope{Key: dev.AppSKey.Bytes()},
				FNwkSIntKey: &ttnpb.KeyEnvelope{Key: dev.NwkSKey.Bytes()},
			},
			DevAddr:       dev.DevAddr.Bytes(),
			LastFCntUp:    dev.FCntUp,
			LastNFCntDown: dev.FCntDown,
			StartedAt:     pbtypes.TimestampNow(),
		}
		if deviceSupportsJoin {
			v3dev.Session.Keys.SessionKeyId = generateBytes(16)
		}
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
func (s *Source) RangeDevices(appID string, f func(source.Source, string) error) error {
	s.config.appID = appID

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
