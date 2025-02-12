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
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iotwireless/types"
	"go.thethings.network/lorawan-stack-migrate/pkg/util"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttntypes "go.thethings.network/lorawan-stack/v3/pkg/types"
)

type Device struct{ *types.LoRaWANDevice }

func (d Device) SetFields(dev *ttnpb.EndDevice, noSession bool) (err error) {
	if dev.Ids == nil {
		dev.Ids = &ttnpb.EndDeviceIdentifiers{}
	}
	dev.Ids.DevEui, err = util.UnmarshalTextToBytes(&ttntypes.EUI64{}, aws.ToString(d.DevEui))
	if err != nil {
		return errInvalidDevEUI.WithAttributes("dev_eui", aws.ToString(d.DevEui)).WithCause(err)
	}

	if dev.RootKeys == nil {
		dev.RootKeys = &ttnpb.RootKeys{}
	}

	abp, otaa := d.sessionKeys()

	if otaa.appKey != nil {
		if dev.RootKeys.AppKey == nil {
			dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{}
		}
		dev.RootKeys.AppKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(otaa.appKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(otaa.appKey)).WithCause(err)
		}
	}
	if otaa.nwkKey != nil {
		if dev.RootKeys.NwkKey == nil {
			dev.RootKeys.NwkKey = &ttnpb.KeyEnvelope{}
		}
		dev.RootKeys.NwkKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(otaa.nwkKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(otaa.nwkKey)).WithCause(err)
		}
	}
	dev.Ids.JoinEui, err = util.UnmarshalTextToBytes(&ttntypes.EUI64{}, aws.ToString(otaa.joinEUI))
	if err != nil {
		return errInvalidJoinEUI.WithAttributes("join_eui", aws.ToString(otaa.joinEUI)).WithCause(err)
	}

	// If we are not exporting session keys, we can return early.
	if noSession {
		return nil
	}

	if dev.Session == nil {
		dev.Session = &ttnpb.Session{
			Keys: &ttnpb.SessionKeys{
				AppSKey:     &ttnpb.KeyEnvelope{},
				FNwkSIntKey: &ttnpb.KeyEnvelope{},
			},
		}
	}
	if abp.appSKey != nil {
		dev.Session.Keys.AppSKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(abp.appSKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(abp.appSKey)).WithCause(err)
		}
	} else {
		dev.Session.Keys.AppSKey.Key = random.Bytes(16)
	}
	if abp.fNwkSIntKey != nil {
		dev.Session.Keys.FNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(abp.fNwkSIntKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(abp.fNwkSIntKey)).WithCause(err)
		}
	} else {
		dev.Session.Keys.FNwkSIntKey.Key = random.Bytes(16)
	}
	if abp.nwkSKey != nil {
		if dev.Session.Keys.NwkSEncKey == nil {
			dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{}
		}
		dev.Session.Keys.NwkSEncKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(abp.nwkSKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(abp.nwkSKey)).WithCause(err)
		}
	}
	if abp.sNwkSIntKey != nil {
		if dev.Session.Keys.SNwkSIntKey == nil {
			dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{}
		}
		dev.Session.Keys.SNwkSIntKey.Key, err = util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(abp.sNwkSIntKey))
		if err != nil {
			return errInvalidKey.WithAttributes("key", aws.ToString(abp.sNwkSIntKey)).WithCause(err)
		}
	}

	if abp.devAddr != nil {
		dev.Session.DevAddr, err = util.UnmarshalTextToBytes(&ttntypes.DevAddr{}, aws.ToString(abp.devAddr))
		if err != nil {
			return errInvalidDevAddr.WithAttributes("dev_addr", aws.ToString(abp.devAddr)).WithCause(err)
		}
	} else {
		dev.Session.DevAddr = random.Bytes(4)
		dev.Session.Keys.SessionKeyId = random.Bytes(16)
	}

	return nil
}

type abpKeys struct {
	appSKey, fNwkSIntKey, nwkSKey, sNwkSIntKey, devAddr *string
}

type otaaKeys struct {
	joinEUI, appKey, nwkKey *string
}

func (d Device) sessionKeys() (abp abpKeys, otaa otaaKeys) {
	if v := d.AbpV1_0_x; v != nil {
		abp = abpKeys{
			devAddr:     v.DevAddr,
			appSKey:     v.SessionKeys.AppSKey,
			fNwkSIntKey: v.SessionKeys.NwkSKey,
		}
	}
	if v := d.AbpV1_1; v != nil {
		abp = abpKeys{
			devAddr:     v.DevAddr,
			appSKey:     v.SessionKeys.AppSKey,
			fNwkSIntKey: v.SessionKeys.FNwkSIntKey,
			nwkSKey:     v.SessionKeys.NwkSEncKey,
			sNwkSIntKey: v.SessionKeys.SNwkSIntKey,
		}
	}
	if o := d.OtaaV1_0_x; o != nil {
		otaa = otaaKeys{
			appKey:  o.AppKey,
			joinEUI: o.AppEui,
		}
		if eui := o.AppEui; eui != nil {
			otaa.joinEUI = eui
		}
	}
	if o := d.OtaaV1_1; o != nil {
		otaa = otaaKeys{
			appKey:  o.AppKey,
			joinEUI: o.JoinEui,
			nwkKey:  o.NwkKey,
		}
	}
	return abp, otaa
}
