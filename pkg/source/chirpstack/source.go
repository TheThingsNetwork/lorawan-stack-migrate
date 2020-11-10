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

package chirpstack

import (
	"context"
	"fmt"
	"math"
	"time"

	csapi "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	config

	cc *grpc.ClientConn

	applications map[int64]*csapi.Application
	devProfiles  map[string]*csapi.DeviceProfile
	svcProfiles  map[string]*csapi.ServiceProfile
}

// NewSource creates a new ChirpStack Source.
func NewSource(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) {
	p := &Source{}

	var err error
	p.config, err = buildConfig(ctx, flags)
	if err != nil {
		return nil, err
	}

	dialOpts := []grpc.DialOption{
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(token(p.token)),
	}
	if p.insecure && p.ca == "" {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	if p.tls != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(p.tls)))
	}
	p.cc, err = grpc.Dial(p.url, dialOpts...)
	if err != nil {
		return nil, err
	}

	log.FromContext(p.ctx).WithFields(p.logFields()).Info("Initialized ChirpStack source")

	p.applications = make(map[int64]*csapi.Application)
	p.devProfiles = make(map[string]*csapi.DeviceProfile)
	p.svcProfiles = make(map[string]*csapi.ServiceProfile)
	return p, nil
}

// RangeDevices implements the Source interface.
func (p *Source) RangeDevices(id string, f func(source.Source, string) error) error {
	app, err := p.getApplication(id)
	if err != nil {
		return err
	}
	client := csapi.NewDeviceServiceClient(p.cc)
	offset := int64(0)
	for {
		devices, err := client.List(p.ctx, &csapi.ListDeviceRequest{
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
	dev.EndDeviceIdentifiers.DevEUI = &types.EUI64{}
	dev.EndDeviceIdentifiers.JoinEUI = &types.EUI64{}
	dev.Attributes = make(map[string]string)
	dev.MACSettings = &ttnpb.MACSettings{}
	dev.MACState = &ttnpb.MACState{}
	dev.RootKeys = &ttnpb.RootKeys{}
	dev.Formatters = &ttnpb.MessagePayloadFormatters{}

	csdev, err := p.getDevice(devEui)
	if err != nil {
		return nil, err
	}
	app, err := p.getApplicationByID(csdev.ApplicationId)
	if err != nil {
		return nil, err
	}
	svcProfile, err := p.getServiceProfile(app.ServiceProfileId)
	if err != nil {
		return nil, err
	}
	devProfile, err := p.getDeviceProfile(csdev.DeviceProfileId)
	if err != nil {
		return nil, err
	}

	// Identifiers
	if err := dev.EndDeviceIdentifiers.DevEUI.UnmarshalText([]byte(devEui)); err != nil {
		return nil, errInvalidDevEUI.WithAttributes("dev_eui", devEui).WithCause(err)
	}
	dev.EndDeviceIdentifiers.JoinEUI = p.joinEUI
	dev.ApplicationIdentifiers.ApplicationID = fmt.Sprintf("chirpstack-%d", csdev.ApplicationId)
	dev.EndDeviceIdentifiers.DeviceID = "eui-" + devEui

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
	if p.exportVars {
		for key, value := range csdev.Variables {
			dev.Attributes["var-"+key] = value
		}
	}

	// Service Profile
	if svcProfile.DevStatusReqFreq > 0 {
		// ChirpStack frequency is requests/day. TTS is time.Duration of interval
		d := time.Duration(24) * time.Hour / time.Duration(svcProfile.DevStatusReqFreq)
		dev.MACSettings.StatusTimePeriodicity = &d
	}

	// Frequency Plan
	dev.FrequencyPlanID = p.frequencyPlanID

	// General
	switch devProfile.MacVersion {
	case "1.0.0":
		dev.LoRaWANVersion = ttnpb.MAC_V1_0
		dev.LoRaWANPHYVersion = ttnpb.PHY_V1_0
	case "1.0.1":
		dev.LoRaWANVersion = ttnpb.MAC_V1_0_1
		dev.LoRaWANPHYVersion = ttnpb.PHY_V1_0_1
	case "1.0.2":
		dev.LoRaWANVersion = ttnpb.MAC_V1_0_2
		switch devProfile.RegParamsRevision {
		case "A":
			dev.LoRaWANPHYVersion = ttnpb.PHY_V1_0_2_REV_A
		case "B":
			dev.LoRaWANPHYVersion = ttnpb.PHY_V1_0_2_REV_B
		default:
			return nil, errInvalidPHYVersion.WithAttributes("phy_version", devProfile.RegParamsRevision)
		}
	case "1.0.3":
		dev.LoRaWANVersion = ttnpb.MAC_V1_0_3
		dev.LoRaWANPHYVersion = ttnpb.PHY_V1_0_3_REV_A
	case "1.1.0":
		dev.LoRaWANVersion = ttnpb.MAC_V1_1
		switch devProfile.RegParamsRevision {
		case "A":
			dev.LoRaWANPHYVersion = ttnpb.PHY_V1_1_REV_A
		case "B":
			dev.LoRaWANPHYVersion = ttnpb.PHY_V1_1_REV_B
		default:
			return nil, errInvalidPHYVersion.WithAttributes("phy_version", devProfile.RegParamsRevision)
		}
	default:
		return nil, errInvalidMACVersion.WithAttributes("mac_version", devProfile.MacVersion)
	}
	if devProfile.MaxEirp > 0 {
		dev.MACState.DesiredParameters.MaxEIRP = float32(devProfile.MaxEirp)
	}

	// Join (OTAA/ABP)
	dev.SupportsJoin = devProfile.SupportsJoin
	if !dev.SupportsJoin {
		if devProfile.RxFreq_2 > 0 {
			dev.MACSettings.Rx2Frequency = &pbtypes.UInt64Value{
				Value: uint64(devProfile.RxFreq_2),
			}
		}
		if devProfile.RxDelay_1 > 0 {
			dev.MACSettings.Rx1Delay = &ttnpb.RxDelayValue{
				Value: ttnpb.RxDelay(devProfile.RxDelay_1),
			}
		}
		if devProfile.RxDrOffset_1 >= 0 {
			dev.MACSettings.DesiredRx1DataRateOffset = &pbtypes.UInt32Value{
				Value: devProfile.RxDrOffset_1,
			}
		}
		if devProfile.RxDatarate_2 >= 0 {
			dev.MACSettings.DesiredRx2DataRateIndex = &ttnpb.DataRateIndexValue{
				Value: ttnpb.DataRateIndex(devProfile.RxDatarate_2),
			}
		}
		for _, freq := range devProfile.FactoryPresetFreqs {
			dev.MACSettings.FactoryPresetFrequencies = append(dev.MACSettings.FactoryPresetFrequencies, uint64(freq))
		}
	}

	// Class B
	dev.SupportsClassB = devProfile.SupportsClassB
	if dev.SupportsClassB {
		if devProfile.ClassBTimeout > 0 {
			timeout := time.Duration(devProfile.ClassBTimeout) * time.Second
			dev.MACSettings.ClassBTimeout = &timeout
		}
		// ChirpStack API returns 2^(seconds + 5)
		dev.MACSettings.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
			Value: ttnpb.PingSlotPeriod(math.Log2(float64(devProfile.PingSlotPeriod)) - 5),
		}

		if devProfile.PingSlotFreq > 0 {
			dev.MACSettings.DesiredPingSlotFrequency = &pbtypes.UInt64Value{
				Value: uint64(devProfile.PingSlotFreq),
			}
		}
		dev.MACSettings.DesiredPingSlotDataRateIndex = &ttnpb.DataRateIndexValue{
			Value: ttnpb.DataRateIndex(devProfile.PingSlotDr),
		}
	}

	// Class C
	dev.SupportsClassC = devProfile.SupportsClassC
	if dev.SupportsClassC {
		if devProfile.ClassCTimeout > 0 {
			timeout := time.Duration(devProfile.ClassCTimeout) * time.Second
			dev.MACSettings.ClassCTimeout = &timeout
		}
	}

	// Root Keys
	rootKeys, err := p.getRootKeys(devEui)
	if err == nil {
		switch dev.LoRaWANVersion {
		case ttnpb.MAC_V1_1:
			dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{
				Key: &types.AES128Key{},
			}
			if err := dev.RootKeys.AppKey.Key.UnmarshalText([]byte(rootKeys.AppKey)); err != nil {
				return nil, errInvalidKey.WithAttributes(rootKeys.AppKey).WithCause(err)
			}
			dev.RootKeys.NwkKey = &ttnpb.KeyEnvelope{
				Key: &types.AES128Key{},
			}
			if err := dev.RootKeys.NwkKey.Key.UnmarshalText([]byte(rootKeys.NwkKey)); err != nil {
				return nil, errInvalidKey.WithAttributes(rootKeys.NwkKey).WithCause(err)
			}
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2, ttnpb.MAC_V1_0_3, ttnpb.MAC_V1_0_4:
			// For LoRaWAN v1.0.x, ChirpStack stores AppKey as NwkKey
			dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{
				Key: &types.AES128Key{},
			}
			if err := dev.RootKeys.AppKey.Key.UnmarshalText([]byte(rootKeys.NwkKey)); err != nil {
				return nil, errInvalidKey.WithAttributes(rootKeys.NwkKey).WithCause(err)
			}
		}
	}

	// Payload formatters
	switch devProfile.PayloadCodec {
	case "CAYENNE_LPP":
		dev.Formatters.UpFormatter = ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP
		dev.Formatters.DownFormatter = ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP
	case "CUSTOM_JS":
		if devProfile.PayloadEncoderScript != "" {
			dev.Formatters.UpFormatter = ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT
			dev.Formatters.UpFormatterParameter = fmt.Sprintf(encoderFormat, devProfile.PayloadEncoderScript)
		}
		if devProfile.PayloadDecoderScript != "" {
			dev.Formatters.DownFormatter = ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT
			dev.Formatters.DownFormatterParameter = fmt.Sprintf(decoderFormat, devProfile.PayloadDecoderScript)
		}
	}

	// Configuration
	if csdev.SkipFCntCheck {
		dev.MACSettings.ResetsFCnt = &pbtypes.BoolValue{
			Value: csdev.SkipFCntCheck,
		}
	}

	// Session
	if p.exportSession {
		activation, err := p.getActivation(devEui)
		if err == nil {
			dev.Session = &ttnpb.Session{}

			if err := dev.Session.DevAddr.UnmarshalText([]byte(activation.DevAddr)); err != nil {
				return nil, errInvalidDevAddr.WithAttributes("dev_addr", activation.DevAddr).WithCause(err)
			}

			// This cannot be empty
			dev.Session.StartedAt = time.Now()

			dev.Session.AppSKey = &ttnpb.KeyEnvelope{
				Key: &types.AES128Key{},
			}
			if err := dev.Session.AppSKey.Key.UnmarshalText([]byte(activation.AppSKey)); err != nil {
				return nil, errInvalidKey.WithAttributes(activation.AppSKey).WithCause(err)
			}
			dev.Session.FNwkSIntKey = &ttnpb.KeyEnvelope{
				Key: &types.AES128Key{},
			}
			if err := dev.Session.FNwkSIntKey.Key.UnmarshalText([]byte(activation.FNwkSIntKey)); err != nil {
				return nil, errInvalidKey.WithAttributes(activation.FNwkSIntKey).WithCause(err)
			}
			switch dev.LoRaWANVersion {
			case ttnpb.MAC_V1_1:
				dev.Session.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{},
				}
				if err := dev.Session.NwkSEncKey.Key.UnmarshalText([]byte(activation.NwkSEncKey)); err != nil {
					return nil, errInvalidKey.WithAttributes(activation.NwkSEncKey).WithCause(err)
				}
				dev.Session.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{},
				}
				if err := dev.Session.SNwkSIntKey.Key.UnmarshalText([]byte(activation.SNwkSIntKey)); err != nil {
					return nil, errInvalidKey.WithAttributes(activation.SNwkSIntKey).WithCause(err)
				}
			default:
			}

			dev.Session.SessionKeyID = generateBytes(16)
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
	return p.cc.Close()
}
