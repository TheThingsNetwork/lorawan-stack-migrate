// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package firefly

import (
	"context"

	"github.com/TheThingsNetwork/go-utils/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.thethings.network/lorawan-stack-migrate/pkg/iterator"
	"go.thethings.network/lorawan-stack-migrate/pkg/log"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/client"
	"go.thethings.network/lorawan-stack-migrate/pkg/util"
)

type Source struct {
	*Config
	*client.Client
}

func createNewSource(cfg *Config) source.CreateSource {
	return func(ctx context.Context, src source.Config) (source.Source, error) {
		if err := cfg.Initialize(src); err != nil {
			return nil, err
		}
		client, err := cfg.NewClient(log.NewContext(ctx, src.Logger))
		if err != nil {
			return nil, err
		}
		return Source{
			Config: cfg,
			Client: client,
		}, nil
	}
}

// Iterator implements source.Source.
func (s Source) Iterator() iterator.Iterator {
	if s.all {
		// The Firefly LNS does not group devices by an application.
		// When the "all" flag is set, we get all devices accessible by the API key.
		// We use a dummy "all" App ID to fallthrough to the RangeDevices method,
		// where the appID argument is unused.
		return iterator.NewListIterator(
			[]string{"all"},
		)
	}
	return iterator.NewNoopIterator()
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devEUIString string) (*ttnpb.EndDevice, error) {
	ffdev, err := s.GetDeviceByEUI(devEUIString)
	if err != nil {
		return nil, err
	}
	if ffdev == nil {
		return nil, errNoDeviceFound.WithAttributes("eui", devEUIString)
	}

	var (
		devEUI, joinEUI types.EUI64
	)
	if err := devEUI.UnmarshalText([]byte(devEUIString)); err != nil {
		return nil, err
	}
	if err := joinEUI.UnmarshalText([]byte(s.joinEUI)); err != nil {
		return nil, err
	}

	v3dev := &ttnpb.EndDevice{
		Name:            ffdev.Name,
		Description:     ffdev.Description,
		FrequencyPlanId: s.frequencyPlanID,
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: s.appID},
			DevEui:         devEUI.Bytes(),
			JoinEui:        joinEUI.Bytes(),
		},
		MacSettings: &ttnpb.MACSettings{
			DesiredAdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADRAckLimitExponent(ffdev.AdrLimit)},
			Rx2DataRateIndex:           &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(ffdev.Rx2DataRate)},
			StatusCountPeriodicity:     wrapperspb.UInt32(0),
			StatusTimePeriodicity:      durationpb.New(0),
		},
		SupportsClassC:    ffdev.ClassC,
		SupportsJoin:      ffdev.OTAA,
		LorawanVersion:    s.derivedMacVersion,
		LorawanPhyVersion: s.derivedPhyVersion,
	}

	if ffdev.Location != nil {
		v3dev.Locations = map[string]*ttnpb.Location{
			"user": {
				Latitude:  ffdev.Location.Latitude,
				Longitude: ffdev.Location.Longitude,
				Source:    ttnpb.LocationSource_SOURCE_REGISTRY,
			},
		}
		s.src.Logger.Debugw("Set location", "location", v3dev.Locations)
	}
	v3dev.Ids.DevEui, err = util.UnmarshalTextToBytes(&types.EUI64{}, ffdev.EUI)
	if err != nil {
		return nil, err
	}
	v3dev.Ids.JoinEui, err = util.UnmarshalTextToBytes(&types.EUI64{}, s.joinEUI)
	if err != nil {
		return nil, err
	}
	if v3dev.SupportsJoin {
		v3dev.RootKeys = &ttnpb.RootKeys{AppKey: &ttnpb.KeyEnvelope{}}
		v3dev.RootKeys.AppKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, ffdev.ApplicationKey)
		if err != nil {
			return nil, err
		}
	}
	hasSession := ffdev.Address != "" && ffdev.NetworkSessionKey != "" && ffdev.ApplicationSessionKey != ""

	if hasSession || !v3dev.SupportsJoin {
		v3dev.Session = &ttnpb.Session{Keys: &ttnpb.SessionKeys{AppSKey: &ttnpb.KeyEnvelope{}, FNwkSIntKey: &ttnpb.KeyEnvelope{}}}
		v3dev.Session.DevAddr, err = util.UnmarshalTextToBytes(&types.DevAddr{}, ffdev.Address)
		if err != nil {
			return nil, err
		}
		// This cannot be empty
		v3dev.Session.StartedAt = timestamppb.Now()
		v3dev.Session.Keys.SessionKeyId = random.Bytes(16)

		v3dev.Session.Keys.AppSKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, ffdev.ApplicationSessionKey)
		if err != nil {
			return nil, err
		}
		v3dev.Session.Keys.FNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, ffdev.NetworkSessionKey)
		if err != nil {
			return nil, err
		}
		switch v3dev.LorawanVersion {
		case ttnpb.MACVersion_MAC_V1_1:
			v3dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{}
			v3dev.Session.Keys.NwkSEncKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, ffdev.ApplicationSessionKey)
			if err != nil {
				return nil, err
			}
			v3dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{}
			v3dev.Session.Keys.SNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, ffdev.NetworkSessionKey)
			if err != nil {
				return nil, err
			}
		}

		// Set FrameCounters
		packet, err := s.GetLastPacket(devEUIString)
		if err != nil {
			return nil, err
		}
		v3dev.Session.LastFCntUp = uint32(packet.FCnt)
		v3dev.Session.LastAFCntDown = uint32(ffdev.FrameCounter)
		v3dev.Session.LastNFCntDown = uint32(ffdev.FrameCounter)
	}

	if s.invalidateKeys {
		s.src.Logger.Debugw("Increment the last byte of the device keys", "device_id", ffdev.Name, "device_eui", ffdev.EUI)
		// Increment the last byte of the device keys.
		// This makes it easier to rollback a migration if needed.
		updated := ffdev.WithIncrementedKeys()
		err := s.UpdateDeviceByEUI(devEUIString, updated)
		if err != nil {
			return nil, err
		}
	}

	return v3dev, nil
}

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(_ string, f func(source.Source, string) error) error {
	var (
		devs []client.Device
		err  error
	)
	s.src.Logger.Debugw("Firefly LNS does not group devices by an application. Get all devices accessible by the API key")
	devs, err = s.GetAllDevices()
	if err != nil {
		return err
	}
	for _, d := range devs {
		if err := f(s, d.EUI); err != nil {
			return err
		}
	}
	return nil
}

// Close implements the Source interface.
func (s Source) Close() error { return nil }
