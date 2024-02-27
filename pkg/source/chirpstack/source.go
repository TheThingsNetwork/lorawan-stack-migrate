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

package chirpstack

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	csv4api "github.com/chirpstack/chirpstack/api/go/v4/api"
	"github.com/chirpstack/chirpstack/api/go/v4/common"
	"go.thethings.network/lorawan-stack-migrate/pkg/iterator"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/chirpstack/config"
	"go.thethings.network/lorawan-stack-migrate/pkg/util"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	encoderFormat = `%s

function Encoder(payload, fport) {
	return Encode(fport, payload, null);
}`
	decoderFormat = `%s

function Decoder(bytes, fport) {
	return Decode(fport, bytes, null);
}`
)

// Source implements the Source interface.
type Source struct {
	*config.Config

	ctx context.Context

	applications map[string]*csv4api.Application
	devProfiles  map[string]*csv4api.DeviceProfile
}

func createNewSource(cfg *config.Config) source.CreateSource {
	return func(ctx context.Context, src source.Config) (source.Source, error) {
		if err := cfg.Initialize(src); err != nil {
			return nil, err
		}
		return &Source{
			ctx:          ctx,
			Config:       cfg,
			applications: make(map[string]*csv4api.Application),
			devProfiles:  make(map[string]*csv4api.DeviceProfile),
		}, nil
	}
}

// Iterator implements source.Source.
func (s Source) Iterator(bool) iterator.Iterator {
	return iterator.NewReaderIterator(os.Stdin, '\n')
}

// RangeDevices implements the Source interface.
func (p *Source) RangeDevices(id string, f func(source.Source, string) error) error {
	app, err := p.getApplication(id)
	if err != nil {
		return err
	}
	client := csv4api.NewDeviceServiceClient(p.ClientConn)
	offset := uint32(0)
	for {
		devices, err := client.List(p.ctx, &csv4api.ListDevicesRequest{
			ApplicationId: app.Id,
			Limit:         limit,
			Offset:        offset,
		})
		if err != nil {
			return err
		}
		for _, devListItem := range devices.Result {
			if err := f(p, devListItem.DevEui); err != nil {
				return err
			}
		}

		if offset += limit; offset > devices.TotalCount {
			break
		}
	}
	return nil
}

// ExportDevice implements the Source interface.
func (p *Source) ExportDevice(devEui string) (*ttnpb.EndDevice, error) {
	// Allocate
	dev := &ttnpb.EndDevice{}
	dev.Attributes = make(map[string]string)
	dev.Formatters = &ttnpb.MessagePayloadFormatters{}
	dev.Ids = &ttnpb.EndDeviceIdentifiers{ApplicationIds: &ttnpb.ApplicationIdentifiers{}}
	dev.MacSettings = &ttnpb.MACSettings{}
	dev.RootKeys = &ttnpb.RootKeys{}

	csdev, err := p.getDevice(devEui)
	if err != nil {
		return nil, err
	}
	devProfile, err := p.getDeviceProfile(csdev.DeviceProfileId)
	if err != nil {
		return nil, err
	}

	// Identifiers
	dev.Ids.DevEui, err = util.UnmarshalTextToBytes(&types.EUI64{}, devEui)
	if err != nil {
		return nil, errInvalidDevEUI.WithAttributes("dev_eui", devEui).WithCause(err)
	}
	dev.Ids.JoinEui = p.JoinEUI.Bytes()
	dev.Ids.ApplicationIds.ApplicationId = fmt.Sprintf("chirpstack-%d", csdev.ApplicationId)
	dev.Ids.DeviceId = "eui-" + strings.ToLower(devEui)

	// Information
	dev.Name = csdev.Name
	dev.Description = csdev.Description
	dev.Attributes["chirpstack-device-profile"] = csdev.DeviceProfileId
	for key, value := range devProfile.Tags {
		dev.Attributes[key] = value
	}
	for key, value := range csdev.Tags {
		dev.Attributes[key] = value
	}
	if p.ExportVars {
		for key, value := range csdev.Variables {
			dev.Attributes["var-"+key] = value
		}
	}

	if d := devProfile.DeviceStatusReqInterval; d > 0 {
		dev.MacSettings.StatusTimePeriodicity = durationpb.New(time.Duration(d))
	}

	// Frequency Plan
	dev.FrequencyPlanId = p.FrequencyPlanID

	// General
	switch devProfile.MacVersion {
	case common.MacVersion_LORAWAN_1_0_0:
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0
		dev.LorawanPhyVersion = ttnpb.PHYVersion_TS001_V1_0
	case common.MacVersion_LORAWAN_1_0_1:
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_1
		dev.LorawanPhyVersion = ttnpb.PHYVersion_TS001_V1_0_1
	case common.MacVersion_LORAWAN_1_0_2:
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_2
		switch devProfile.RegParamsRevision {
		case common.RegParamsRevision_A:
			dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2
		case common.RegParamsRevision_B:
			dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2_REV_B
		default:
			return nil, errInvalidPHYVersion.WithAttributes("phy_version", devProfile.RegParamsRevision)
		}
	case common.MacVersion_LORAWAN_1_0_3:
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_3
		dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_3_REV_A
	case common.MacVersion_LORAWAN_1_1_0:
		dev.LorawanVersion = ttnpb.MACVersion_MAC_V1_1
		switch devProfile.RegParamsRevision {
		case common.RegParamsRevision_A:
			dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_1_REV_A
		case common.RegParamsRevision_B:
			dev.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_1_REV_B
		default:
			return nil, errInvalidPHYVersion.WithAttributes("phy_version", devProfile.RegParamsRevision)
		}
	default:
		return nil, errInvalidMACVersion.WithAttributes("mac_version", devProfile.MacVersion)
	}

	// Join (OTAA/ABP)
	dev.SupportsJoin = devProfile.SupportsOtaa
	if !dev.SupportsJoin {
		if freq := devProfile.AbpRx2Freq; freq > 0 {
			dev.MacSettings.Rx2Frequency = &ttnpb.FrequencyValue{
				Value: uint64(freq),
			}
		}
		if delay := devProfile.AbpRx1Delay; delay > 0 {
			dev.MacSettings.Rx1Delay = &ttnpb.RxDelayValue{
				Value: ttnpb.RxDelay(delay),
			}
		}
		if offset := devProfile.AbpRx1DrOffset; offset > 0 {
			dev.MacSettings.DesiredRx1DataRateOffset = &ttnpb.DataRateOffsetValue{
				Value: ttnpb.DataRateOffset(offset),
			}
		}
		if dataRate := devProfile.AbpRx2Dr; dataRate > 0 {
			dev.MacSettings.DesiredRx2DataRateIndex = &ttnpb.DataRateIndexValue{
				Value: ttnpb.DataRateIndex(dataRate),
			}
		}
	}

	// Class B
	dev.SupportsClassB = devProfile.SupportsClassB
	if dev.SupportsClassB {
		if devProfile.ClassBTimeout > 0 {
			timeout := time.Duration(devProfile.ClassBTimeout) * time.Second
			dev.MacSettings.ClassBTimeout = durationpb.New(timeout)
		}
		// ChirpStack API returns 2^(seconds + 5)
		dev.MacSettings.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
			Value: ttnpb.PingSlotPeriod(math.Log2(float64(devProfile.ClassBPingSlotNbK)) - 5),
		}

		if devProfile.ClassBPingSlotNbK > 0 {
			dev.MacSettings.DesiredPingSlotFrequency = &ttnpb.ZeroableFrequencyValue{
				Value: uint64(devProfile.ClassBPingSlotNbK),
			}
		}
	}

	// Class C
	dev.SupportsClassC = devProfile.SupportsClassC
	if dev.SupportsClassC {
		if devProfile.ClassCTimeout > 0 {
			timeout := time.Duration(devProfile.ClassCTimeout) * time.Second
			dev.MacSettings.ClassCTimeout = durationpb.New(timeout)
		}
	}

	// Root Keys
	rootKeys, err := p.getRootKeys(devEui)
	if err == nil {
		switch dev.LorawanVersion {
		case ttnpb.MACVersion_MAC_V1_1:
			dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{}
			dev.RootKeys.AppKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, rootKeys.AppKey)
			if err != nil {
				return nil, errInvalidKey.WithAttributes(rootKeys.AppKey).WithCause(err)
			}
			dev.RootKeys.NwkKey = &ttnpb.KeyEnvelope{}
			dev.RootKeys.NwkKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, rootKeys.NwkKey)
			if err != nil {
				return nil, errInvalidKey.WithAttributes(rootKeys.NwkKey).WithCause(err)
			}
		case ttnpb.MACVersion_MAC_V1_0, ttnpb.MACVersion_MAC_V1_0_1, ttnpb.MACVersion_MAC_V1_0_2, ttnpb.MACVersion_MAC_V1_0_3, ttnpb.MACVersion_MAC_V1_0_4:
			// For LoRaWAN v1.0.x, ChirpStack stores AppKey as NwkKey
			dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{}
			dev.RootKeys.AppKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, rootKeys.NwkKey)
			if err != nil {
				return nil, errInvalidKey.WithAttributes(rootKeys.NwkKey).WithCause(err)
			}
		}
	}

	// Payload formatters
	switch devProfile.PayloadCodecRuntime {
	case csv4api.CodecRuntime_CAYENNE_LPP:
		dev.Formatters.UpFormatter = ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP
		dev.Formatters.DownFormatter = ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP
	case csv4api.CodecRuntime_JS:
		if s := devProfile.PayloadCodecScript; s != "" {
			dev.Formatters.UpFormatter = ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT
			dev.Formatters.DownFormatter = ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT
			dev.Formatters.UpFormatterParameter = fmt.Sprintf(encoderFormat, s)
			dev.Formatters.DownFormatterParameter = fmt.Sprintf(encoderFormat, s)
		}
	}

	// Configuration
	if csdev.SkipFcntCheck {
		dev.MacSettings.ResetsFCnt = &ttnpb.BoolValue{
			Value: csdev.SkipFcntCheck,
		}
	}

	// Export session information.
	if p.ExportSession {
		activation, err := p.getActivation(devEui)
		if err == nil {
			dev.Session = &ttnpb.Session{Keys: &ttnpb.SessionKeys{}, StartedAt: timestamppb.Now()}

			devAddr := &types.DevAddr{}
			if err := devAddr.UnmarshalText([]byte(activation.DevAddr)); err != nil {
				return nil, errInvalidDevAddr.WithAttributes("dev_addr", activation.DevAddr).WithCause(err)
			}
			dev.Session.DevAddr = devAddr.Bytes()

			// This cannot be empty
			dev.Session.StartedAt = timestamppb.Now()

			dev.Session.Keys.AppSKey = &ttnpb.KeyEnvelope{}
			dev.Session.Keys.AppSKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, activation.AppSKey)
			if err != nil {
				return nil, errInvalidKey.WithAttributes(activation.AppSKey).WithCause(err)
			}
			dev.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{}
			dev.Session.Keys.FNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, activation.FNwkSIntKey)
			if err != nil {
				return nil, errInvalidKey.WithAttributes(activation.FNwkSIntKey).WithCause(err)
			}
			switch dev.LorawanVersion {
			case ttnpb.MACVersion_MAC_V1_1:
				dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{}
				dev.Session.Keys.NwkSEncKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, activation.NwkSEncKey)
				if err != nil {
					return nil, errInvalidKey.WithAttributes(activation.NwkSEncKey).WithCause(err)
				}
				dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{}
				dev.Session.Keys.SNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, activation.SNwkSIntKey)
				if err != nil {
					return nil, errInvalidKey.WithAttributes(activation.SNwkSIntKey).WithCause(err)
				}
			default:
			}

			if devProfile.SupportsOtaa {
				dev.Session.Keys.SessionKeyId = generateBytes(16)
			}
			dev.Session.LastAFCntDown = activation.AFCntDown
			dev.Session.LastFCntUp = activation.FCntUp
			dev.Session.LastConfFCntDown = activation.FCntUp
			dev.Session.LastNFCntDown = activation.NFCntDown
		}
	}

	return dev, nil
}

// Close implements the Source interface.
func (p *Source) Close() error {
	return p.ClientConn.Close()
}
