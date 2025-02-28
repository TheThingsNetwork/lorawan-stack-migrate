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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iotwireless/types"
	"go.thethings.network/lorawan-stack-migrate/pkg/util"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttntypes "go.thethings.network/lorawan-stack/v3/pkg/types"
)

type Device struct{ *types.LoRaWANDevice }

type DeviceIdentifiers struct {
	Id, Name, Description *string
}

// SetOTAADevice sets the OTAA device fields.
func (d Device) SetOTAADevice(dev *ttnpb.EndDevice) (err error) {
	rootKeys := d.rootKeys()
	if rootKeys.appKey != nil {
		dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{}
		dev.RootKeys.AppKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(rootKeys.appKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(rootKeys.appKey)).WithCause(err)
		}
	} else {
		return errEmptyKey.WithAttributes("key", "AppKey")
	}
	if dev.LorawanVersion == ttnpb.MACVersion_MAC_V1_1 {
		if rootKeys.nwkKey != nil {
			dev.RootKeys.NwkKey = &ttnpb.KeyEnvelope{}
			dev.RootKeys.NwkKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(rootKeys.nwkKey))
			if err != nil {
				return errInvalidKey.WithAttributes("key", aws.ToString(rootKeys.nwkKey)).WithCause(err)
			}
		} else {
			return errEmptyKey.WithAttributes("key", "NwkKey")
		}
	}
	joinEui := d.joinEui()
	dev.Ids.JoinEui, err = util.UnmarshalTextToBytes(&ttntypes.EUI64{}, aws.ToString(joinEui))
	if err != nil {
		return errInvalidJoinEUI.WithAttributes("join_eui", aws.ToString(joinEui)).WithCause(err)
	}
	return nil
}

// SetABPDevice sets the ABP device fields.
func (d Device) SetABPDevice(dev *ttnpb.EndDevice) (err error) {
	dev.Session = &ttnpb.Session{Keys: &ttnpb.SessionKeys{}}

	sessionKeys := d.sessionKeys()
	if sessionKeys.appSKey != nil {
		dev.Session.Keys.AppSKey = &ttnpb.KeyEnvelope{}
		dev.Session.Keys.AppSKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(sessionKeys.appSKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(sessionKeys.appSKey)).WithCause(err)
		}
	} else {
		return errEmptyKey.WithAttributes("key", "AppSKey")
	}
	if sessionKeys.fNwkSIntKey != nil {
		dev.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{}
		dev.Session.Keys.FNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(sessionKeys.fNwkSIntKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(sessionKeys.fNwkSIntKey)).WithCause(err)
		}
	} else {
		return errEmptyKey.WithAttributes("key", "FNwkSIntKey")
	}
	if dev.LorawanVersion == ttnpb.MACVersion_MAC_V1_1 {
		if sessionKeys.nwkSEncKey != nil {
			dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{}
			dev.Session.Keys.NwkSEncKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(sessionKeys.nwkSEncKey))
			if err != nil {
				return errInvalidKey.WithAttributes("key", aws.ToString(sessionKeys.nwkSEncKey)).WithCause(err)
			}
		} else {
			return errEmptyKey.WithAttributes("key", "NwkSEncKey")
		}
		if sessionKeys.sNwkSIntKey != nil {
			dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{}
			dev.Session.Keys.SNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(sessionKeys.sNwkSIntKey))
			if err != nil {
				return errInvalidKey.WithAttributes("key", aws.ToString(sessionKeys.sNwkSIntKey)).WithCause(err)
			}
		} else {
			return errEmptyKey.WithAttributes("key", "SNwkSIntKey")
		}
	}
	devAddr := d.devAddr()
	if devAddr != nil {
		dev.Session.DevAddr, err = util.UnmarshalTextToBytes(&ttntypes.DevAddr{}, aws.ToString(devAddr))
		if err != nil {
			return errInvalidDevAddr.WithAttributes("dev_addr", aws.ToString(devAddr)).WithCause(err)
		}
	}
	return nil
}

type sessionKeys struct {
	appSKey, fNwkSIntKey, nwkSEncKey, sNwkSIntKey *string
}

type rootKeys struct {
	appKey, nwkKey *string
}

func (d Device) sessionKeys() sessionKeys {
	k := sessionKeys{}
	if d.AbpV1_0_x != nil {
		k.appSKey = d.AbpV1_0_x.SessionKeys.AppSKey
		k.fNwkSIntKey = d.AbpV1_0_x.SessionKeys.NwkSKey
	}
	if d.AbpV1_1 != nil {
		k.appSKey = d.AbpV1_1.SessionKeys.AppSKey
		k.fNwkSIntKey = d.AbpV1_1.SessionKeys.FNwkSIntKey
		k.nwkSEncKey = d.AbpV1_1.SessionKeys.NwkSEncKey
		k.sNwkSIntKey = d.AbpV1_1.SessionKeys.SNwkSIntKey
	}
	return k
}

func (d Device) devAddr() (addr *string) {
	if d.AbpV1_0_x != nil {
		addr = d.AbpV1_0_x.DevAddr
	}
	if d.AbpV1_1 != nil {
		addr = d.AbpV1_1.DevAddr
	}
	return addr
}

func (d Device) rootKeys() (keys rootKeys) {
	if d.OtaaV1_0_x != nil {
		keys.appKey = d.OtaaV1_0_x.AppKey
	}
	if d.OtaaV1_1 != nil {
		keys.appKey = d.OtaaV1_1.AppKey
		keys.nwkKey = d.OtaaV1_1.NwkKey
	}
	return keys
}

func (d Device) joinEui() (joinEui *string) {
	if d.OtaaV1_0_x != nil {
		if d.OtaaV1_0_x.AppEui != nil {
			joinEui = d.OtaaV1_0_x.AppEui
		} else {
			joinEui = d.OtaaV1_0_x.JoinEui
		}
	}
	if d.OtaaV1_1 != nil {
		joinEui = d.OtaaV1_1.JoinEui
	}
	return joinEui
}
