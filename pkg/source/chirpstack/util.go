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
	"strconv"

	csapi "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	log "go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const limit = int64(100)

func generateBytes(length int) []byte {
	return random.Bytes(length)
}

func (p *Source) getDeviceProfile(id string) (*csapi.DeviceProfile, error) {
	if profile, ok := p.devProfiles[id]; ok {
		return profile, nil
	}

	client := csapi.NewDeviceProfileServiceClient(p.cc)
	resp, err := client.Get(p.ctx, &csapi.GetDeviceProfileRequest{
		Id: id,
	})
	if err != nil {
		return nil, errAPI.WithCause(err)
	}
	p.devProfiles[id] = resp.DeviceProfile
	return resp.DeviceProfile, nil
}

func (p *Source) getServiceProfile(id string) (*csapi.ServiceProfile, error) {
	if profile, ok := p.svcProfiles[id]; ok {
		return profile, nil
	}

	client := csapi.NewServiceProfileServiceClient(p.cc)
	resp, err := client.Get(p.ctx, &csapi.GetServiceProfileRequest{
		Id: id,
	})
	if err != nil {
		return nil, errAPI.WithCause(err)
	}
	p.svcProfiles[id] = resp.ServiceProfile
	return resp.ServiceProfile, nil
}

func (p *Source) getDevice(devEui string) (*csapi.Device, error) {
	client := csapi.NewDeviceServiceClient(p.cc)

	resp, err := client.Get(p.ctx, &csapi.GetDeviceRequest{
		DevEui: devEui,
	})
	if err != nil {
		return nil, errAPI.WithCause(err)
	}
	return resp.Device, nil
}

func (p *Source) getApplication(application string) (*csapi.Application, error) {
	appID, err := strconv.ParseInt(application, 10, 64)
	if err != nil {
		appID, err = p.getApplicationIDByName(application)
		if err != nil {
			return nil, err
		}
	}
	app, err := p.getApplicationByID(appID)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			appID, err = p.getApplicationIDByName(application)
			if err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
		return p.getApplicationByID(appID)
	}
	return app, nil
}

func (p *Source) getApplicationByID(id int64) (*csapi.Application, error) {
	if app, ok := p.applications[id]; ok {
		return app, nil
	}

	client := csapi.NewApplicationServiceClient(p.cc)
	resp, err := client.Get(p.ctx, &csapi.GetApplicationRequest{
		Id: id,
	})
	if err != nil {
		return nil, errAPI.WithCause(err)
	}

	p.applications[id] = resp.Application
	return resp.Application, nil
}

func (p *Source) getApplicationIDByName(name string) (int64, error) {
	client := csapi.NewApplicationServiceClient(p.cc)
	offset := int64(0)
	for {
		resp, err := client.List(p.ctx, &csapi.ListApplicationRequest{
			Limit:  limit,
			Offset: offset,
			Search: name,
		})
		if err != nil {
			return 0, err
		}
		for _, appListItem := range resp.Result {
			if appListItem.Name == name {
				return appListItem.Id, nil
			}
		}

		if offset += limit; offset > resp.TotalCount {
			return 0, errAppNotFound.WithAttributes("app", name)
		}
	}
}

func (p *Source) getRootKeys(devEui string) (*csapi.DeviceKeys, error) {
	client := csapi.NewDeviceServiceClient(p.cc)
	resp, err := client.GetKeys(context.Background(), &csapi.GetDeviceKeysRequest{
		DevEui: devEui,
	})
	if err != nil {
		log.FromContext(p.ctx).WithField("dev_eui", devEui).WithError(err).Debug("No root keys")
		return nil, err
	}
	return resp.DeviceKeys, err
}

func (p *Source) getActivation(devEui string) (*csapi.DeviceActivation, error) {
	client := csapi.NewDeviceServiceClient(p.cc)
	resp, err := client.GetActivation(context.Background(), &csapi.GetDeviceActivationRequest{
		DevEui: devEui,
	})
	if err != nil {
		log.FromContext(p.ctx).WithField("dev_eui", devEui).WithError(err).Debug("No session keys")
		return nil, err
	}
	return resp.DeviceActivation, err
}

func unmarshalTextToBytes(
	unmarshaller interface {
		UnmarshalText([]byte) error
		Bytes() []byte
	},
	source string,
) ([]byte, error) {
	err := unmarshaller.UnmarshalText([]byte(source))
	return unmarshaller.Bytes(), err
}
