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
	ttntypes "github.com/TheThingsNetwork/ttn/core/types"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// backoff to use before next try when client receives an error of type ResourceExhausted.
const backoff = time.Second

// Source implements the Source interface.
type Source struct {
	ctx context.Context

	config config
	mgr    ttnsdk.DeviceManager
	client ttnsdk.Client
}

// NewSource creates a new TTNv2 Source.
func NewSource(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) {
	config, err := getConfig(ctx, flags)
	if err != nil {
		return nil, err
	}

	s := &Source{
		ctx:    ctx,
		config: config,
		client: config.sdkConfig.NewClient(config.appID, config.appAccessKey),
	}
	s.mgr, err = s.client.ManageDevices()
	if err != nil {
		return nil, err
	}
	s.mgr = &deviceManagerWithRetry{s.mgr, ctx, 15, backoff}
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

	v3dev.MACSettings = &ttnpb.MACSettings{}
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

	deviceSupportsJoin := dev.AppKey != nil && !dev.AppKey.IsEmpty()
	deviceHasSession := dev.DevAddr != nil && !dev.DevAddr.IsEmpty() && dev.NwkSKey != nil && !dev.NwkSKey.IsEmpty() && dev.AppSKey != nil && !dev.AppSKey.IsEmpty()
	if deviceSupportsJoin {
		// OTAA devices
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

	if s.config.withSession && deviceHasSession {
		v3dev.Session = &ttnpb.Session{
			SessionKeys: ttnpb.SessionKeys{
				SessionKeyID: generateBytes(16),
				AppSKey:      &ttnpb.KeyEnvelope{Key: &types.AES128Key{}},
				FNwkSIntKey:  &ttnpb.KeyEnvelope{Key: &types.AES128Key{}},
				NwkSEncKey:   &ttnpb.KeyEnvelope{Key: &types.AES128Key{}},
				SNwkSIntKey:  &ttnpb.KeyEnvelope{Key: &types.AES128Key{}},
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
		if err := v3dev.Session.SessionKeys.NwkSEncKey.Key.Unmarshal(dev.NwkSKey.Bytes()); err != nil {
			return nil, err
		}
		if err := v3dev.Session.SessionKeys.SNwkSIntKey.Key.Unmarshal(dev.NwkSKey.Bytes()); err != nil {
			return nil, err
		}
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
	if deviceHasSession {
		var err error
		if v3dev.MACState, err = mac.NewState(v3dev, s.config.fpStore, ttnpb.MACSettings{}); err != nil {
			return nil, err
		}
		// Ensure MAC state matches v2 configuration.
		v3dev.MACState.CurrentParameters = v3dev.MACState.DesiredParameters
		v3dev.MACState.DeviceClass = ttnpb.CLASS_A
		v3dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_0_2
		v3dev.MACState.CurrentParameters.Rx1Delay = ttnpb.RX_DELAY_1
	}
	if !deviceSupportsJoin {
		v3dev.MACSettings = &ttnpb.MACSettings{
			Rx1Delay: &ttnpb.RxDelayValue{Value: ttnpb.RX_DELAY_1},
		}
	}

	return v3dev, nil
}

// RangeDevices implements the source.Source interface.
func (s *Source) RangeDevices(appID string, f func(source.Source, string) error) error {
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
	b := make([]byte, length)
	random.Read(b)
	return b
}
