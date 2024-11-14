// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

func (d Device) SetFields(dev *ttnpb.EndDevice, noSession bool) (err error) {
	if dev.Ids == nil {
		dev.Ids = &ttnpb.EndDeviceIdentifiers{}
	}

	eui, err := util.UnmarshalTextToBytes(&ttntypes.EUI64{}, aws.ToString(d.DevEui))
	if err != nil {
		return err
	}
	dev.Ids.DevEui = eui

	if dev.Session == nil {
		dev.Session = &ttnpb.Session{}
	}
	if dev.Session.Keys == nil {
		dev.Session.Keys = &ttnpb.SessionKeys{}
	}

	if noSession {
		return nil
	}

	apb, otaa := d.sessionKeys()
	keys := dev.Session.Keys
	if err := unmarshalKeys([]struct {
		envelope *ttnpb.KeyEnvelope
		key      *string
	}{
		{keys.AppSKey, apb.appSKey},
		{keys.NwkSEncKey, apb.nwkSKey},
		{keys.FNwkSIntKey, apb.fNwkSIntKey},
		{keys.SNwkSIntKey, apb.sNwkSIntKey},
		{dev.RootKeys.AppKey, otaa.appKey},
		{dev.RootKeys.NwkKey, otaa.nwkKey},
	}); err != nil {
		return err
	}

	var b []byte
	if b, err = util.UnmarshalTextToBytes(&ttntypes.DevAddr{}, aws.ToString(apb.devAddr)); err != nil {
		return err
	}
	dev.Session.DevAddr = b
	if b, err = util.UnmarshalTextToBytes(&ttntypes.EUI64{}, aws.ToString(otaa.joinEUI)); err != nil {
		return err
	}
	dev.Ids.JoinEui = b

	return nil
}

type apbKeys struct {
	appSKey, fNwkSIntKey, nwkSKey, sNwkSIntKey, devAddr *string
}

type otaaKeys struct {
	joinEUI, appKey, nwkKey *string
}

func (d Device) sessionKeys() (apb apbKeys, otaa otaaKeys) {
	if v := d.AbpV1_0_x; v != nil {
		apb = apbKeys{
			devAddr: v.DevAddr,
			appSKey: v.SessionKeys.AppSKey,
			nwkSKey: v.SessionKeys.NwkSKey,
		}
	}
	if v := d.AbpV1_1; v != nil {
		apb = apbKeys{
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
	return apb, otaa
}

func unmarshalKeys(data []struct {
	envelope *ttnpb.KeyEnvelope
	key      *string
},
) (err error) {
	for _, v := range data {
		if v.envelope == nil {
			v.envelope = &ttnpb.KeyEnvelope{}
		}
		b, err := util.UnmarshalTextToBytes(&ttntypes.AES128Key{}, aws.ToString(v.key))
		if err != nil {
			return err
		}
		v.envelope.Key = b
	}
	return nil
}
