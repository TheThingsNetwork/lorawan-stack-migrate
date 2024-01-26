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
	"encoding/json"
	"strconv"

	"github.com/TheThingsNetwork/go-utils/random"
	"go.thethings.network/lorawan-stack-migrate/pkg/util"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Devices is a list of devices.
type Devices map[types.EUI64]Device

// UnmarshalJSON implements json.Unmarshaler.
func (d Devices) UnmarshalJSON(b []byte) error {
	var devs []Device
	if err := json.Unmarshal(b, &devs); err != nil {
		return err
	}
	for _, dev := range devs {
		var devEUI types.EUI64
		if err := devEUI.UnmarshalText([]byte(dev.DevEui)); err != nil {
			return err
		}
		d[devEUI] = dev
	}
	return nil
}

// Device is a LoRaWAN end device exported from Wanesy Management Center.
type Device struct {
	Activation          string `json:"activation"`
	AdrEnabled          string `json:"adrEnabled"`
	Altitude            string `json:"altitude"`
	AppEui              string `json:"appEui"`
	AppKey              string `json:"appKey"`
	AppSKey             string `json:"appSKey"`
	CfList              string `json:"cfList"`
	ClassType           string `json:"classType"`
	ClusterID           string `json:"clusterId"`
	ClusterName         string `json:"clusterName"`
	Country             string `json:"country"`
	DevAddr             string `json:"devAddr"`
	DevEui              string `json:"devEui"`
	DevNonceCounter     string `json:"devNonceCounter"`
	DwellTime           string `json:"dwellTime"`
	FNwkSIntKey         string `json:"fNwkSIntKey"`
	FcntDown            string `json:"fcntDown"`
	FcntUp              string `json:"fcntUp"`
	Geolocation         string `json:"geolocation"`
	LastDataDownMessage string `json:"lastDataDownMessage"`
	LastDataUpDr        string `json:"lastDataUpDr"`
	LastDataUpMessage   string `json:"lastDataUpMessage"`
	Latitude            string `json:"latitude"`
	Longitude           string `json:"longitude"`
	MacVersion          string `json:"macVersion"`
	Name                string `json:"name"`
	NwkSKey             string `json:"nwkSKey"`
	PingSlotDr          string `json:"pingSlotDr"`
	PingSlotFreq        string `json:"pingSlotFreq"`
	Profile             string `json:"profile"`
	RegParamsRevision   string `json:"regParamsRevision"`
	RfRegion            string `json:"rfRegion"`
	Rx1Delay            string `json:"rx1Delay"`
	Rx1DrOffset         string `json:"rx1DrOffset"`
	Rx2Dr               string `json:"rx2Dr"`
	Rx2Freq             string `json:"rx2Freq"`
	RxWindows           string `json:"rxWindows"`
	SNwkSIntKey         string `json:"sNwkSIntKey"`
	Status              string `json:"status"`
}

// EndDevice converts a Wanesy device to a TTS device.
func (d Device) EndDevice(fpStore *frequencyplans.Store, applicationID, frequencyPlanID string) (*ttnpb.EndDevice, error) {
	var devEUI types.EUI64
	if err := devEUI.UnmarshalText([]byte(d.DevEui)); err != nil {
		return nil, err
	}
	ret := &ttnpb.EndDevice{
		Name:            d.Name,
		FrequencyPlanId: frequencyPlanID,
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: applicationID},
			DevEui:         devEUI.Bytes(),
		},
		MacSettings:    &ttnpb.MACSettings{},
		SupportsClassC: d.ClassType == "C",
		SupportsClassB: d.ClassType == "B",
		SupportsJoin:   d.Activation == "OTAA",
	}
	if ret.SupportsJoin {
		var (
			joinEUI types.EUI64
			err     error
		)
		ret.RootKeys = &ttnpb.RootKeys{AppKey: &ttnpb.KeyEnvelope{}}
		ret.RootKeys.AppKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, d.AppKey)
		if err != nil {
			return nil, err
		}
		if err := joinEUI.UnmarshalText([]byte(d.AppEui)); err != nil {
			return nil, err
		}
		ret.Ids.JoinEui = joinEUI.Bytes()
	}
	if d.Rx2Dr != "NULL" {
		s, err := strconv.ParseUint(d.Rx2Dr, 16, 32)
		if err != nil {
			return nil, err
		}
		ret.MacSettings.DesiredRx2DataRateIndex = &ttnpb.DataRateIndexValue{
			Value: ttnpb.DataRateIndex(int32(s)),
		}
	}
	switch d.MacVersion {
	case "1.0.0":
		ret.LorawanVersion = ttnpb.MACVersion_MAC_V1_0
		ret.LorawanPhyVersion = ttnpb.PHYVersion_TS001_V1_0
	case "1.0.1":
		ret.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_1
		ret.LorawanPhyVersion = ttnpb.PHYVersion_TS001_V1_0_1
	case "1.0.2":
		ret.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_2
		switch d.RegParamsRevision {
		case "A":
			ret.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2
		case "B":
			ret.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_2_REV_B
		default:
			return nil, errInvalidPHYForMACVersion.WithAttributes(
				"phy_version",
				d.RegParamsRevision,
				"mac_version",
				d.MacVersion,
			)
		}
	case "1.0.3":
		ret.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_3
		ret.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_0_3_REV_A
	case "1.0.4":
		ret.LorawanVersion = ttnpb.MACVersion_MAC_V1_0_4
		ret.LorawanPhyVersion = ttnpb.PHYVersion_RP002_V1_0_4
	case "1.1.0":
		ret.LorawanVersion = ttnpb.MACVersion_MAC_V1_1
		switch d.RegParamsRevision {
		case "A":
			ret.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_1_REV_A
		case "B":
			ret.LorawanPhyVersion = ttnpb.PHYVersion_RP001_V1_1_REV_B
		default:
			return nil, errInvalidPHYForMACVersion.WithAttributes(
				"phy_version",
				d.RegParamsRevision,
				"mac_version",
				d.MacVersion,
			)
		}
	default:
		return nil, errInvalidMACVersion.WithAttributes("mac_version", d.MacVersion)
	}

	if d.Longitude != "NULL" && d.Latitude != "NULL" && d.Altitude != "NULL" {
		latitude, _ := strconv.ParseFloat(d.Latitude, 64)
		longitude, _ := strconv.ParseFloat(d.Longitude, 64)
		altitude, err := strconv.ParseUint(d.Rx2Dr, 16, 32)
		if err != nil {
			return nil, err
		}
		ret.Locations = map[string]*ttnpb.Location{
			"user": {
				Latitude:  latitude,
				Longitude: longitude,
				Altitude:  int32(altitude),
				Source:    ttnpb.LocationSource_SOURCE_REGISTRY,
			},
		}
	}

	// Copy session information if available.
	hasSession := d.DevAddr != "NULL" && d.NwkSKey != "NULL" && d.AppKey != ""
	if hasSession || !ret.SupportsJoin {
		var err error
		ret.Session = &ttnpb.Session{Keys: &ttnpb.SessionKeys{AppSKey: &ttnpb.KeyEnvelope{}, FNwkSIntKey: &ttnpb.KeyEnvelope{}}}
		ret.Session.DevAddr, err = util.UnmarshalTextToBytes(&types.DevAddr{}, d.DevAddr)
		if err != nil {
			return nil, err
		}
		if ret.SupportsJoin {
			ret.Session.StartedAt = timestamppb.Now()
			ret.Session.Keys.SessionKeyId = random.Bytes(16)
		}
		ret.Session.Keys.AppSKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, d.AppSKey)
		if err != nil {
			return nil, err
		}
		ret.Session.Keys.FNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, d.NwkSKey)
		if err != nil {
			return nil, err
		}
		switch ret.LorawanVersion {
		case ttnpb.MACVersion_MAC_V1_1:
			ret.Session.Keys.FNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, d.FNwkSIntKey)
			if err != nil {
				return nil, err
			}
			ret.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{}
			ret.Session.Keys.NwkSEncKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, d.NwkSKey)
			if err != nil {
				return nil, err
			}
			ret.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{}
			ret.Session.Keys.SNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&types.AES128Key{}, d.SNwkSIntKey)
			if err != nil {
				return nil, err
			}
		}

		// Set FrameCounters
		s, err := strconv.ParseUint(d.FcntUp, 16, 32)
		if err != nil {
			return nil, err
		}
		ret.Session.LastFCntUp = uint32(s)
		s, err = strconv.ParseUint(d.FcntDown, 16, 32)
		if err != nil {
			return nil, err
		}
		ret.Session.LastAFCntDown = uint32(s)
		ret.Session.LastNFCntDown = uint32(s)

		// Create a MACState.
		if ret.MacState, err = mac.NewState(ret, fpStore, ret.MacSettings); err != nil {
			return nil, err
		}
		ret.MacState.CurrentParameters = ret.MacState.DesiredParameters
		ret.MacState.CurrentParameters.Rx1Delay = ttnpb.RxDelay_RX_DELAY_1 // Fallback
		if d.Rx1Delay != "NULL" {
			s, err = strconv.ParseUint(d.Rx1Delay, 16, 32)
			if err != nil {
				return nil, err
			}
			ret.MacState.CurrentParameters.Rx1Delay = ttnpb.RxDelay(s)
		}
	}

	return ret, nil
}
