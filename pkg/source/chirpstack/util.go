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

	csv4api "github.com/chirpstack/chirpstack/api/go/v4/api"
	log "go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const limit = uint32(100)

func generateBytes(length int) []byte {
	return random.Bytes(length)
}

func (p *Source) getDeviceProfile(id string) (*csv4api.DeviceProfile, error) {
	if profile, ok := p.devProfiles[id]; ok {
		return profile, nil
	}

	client := csv4api.NewDeviceProfileServiceClient(p.ClientConn)
	resp, err := client.Get(p.ctx, &csv4api.GetDeviceProfileRequest{
		Id: id,
	})
	if err != nil {
		return nil, errAPI.WithCause(err)
	}
	p.devProfiles[id] = resp.DeviceProfile
	return resp.DeviceProfile, nil
}

func (p *Source) getDevice(devEui string) (*csv4api.Device, error) {
	client := csv4api.NewDeviceServiceClient(p.ClientConn)

	resp, err := client.Get(p.ctx, &csv4api.GetDeviceRequest{
		DevEui: devEui,
	})
	if err != nil {
		return nil, errAPI.WithCause(err)
	}
	return resp.Device, nil
}

func (p *Source) getApplication(application string) (*csv4api.Application, error) {
	app, err := p.getApplicationByID(application)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			appID, err := p.getApplicationIDByName(application)
			if err != nil {
				return nil, err
			}
			return p.getApplicationByID(appID)
		default:
			return nil, err
		}
	}
	return app, nil
}

func (p *Source) getApplicationByID(id string) (*csv4api.Application, error) {
	if app, ok := p.applications[id]; ok {
		return app, nil
	}

	client := csv4api.NewApplicationServiceClient(p.ClientConn)
	resp, err := client.Get(p.ctx, &csv4api.GetApplicationRequest{
		Id: id,
	})
	if err != nil {
		return nil, errAPI.WithCause(err)
	}

	p.applications[id] = resp.Application
	return resp.Application, nil
}

func (p *Source) getApplicationIDByName(name string) (string, error) {
	client := csv4api.NewApplicationServiceClient(p.ClientConn)
	offset := uint32(0)
	for {
		resp, err := client.List(p.ctx, &csv4api.ListApplicationsRequest{
			Limit:  limit,
			Offset: offset,
			Search: name,
		})
		if err != nil {
			return "", err
		}
		for _, appListItem := range resp.Result {
			if appListItem.Name == name {
				return appListItem.Id, nil
			}
		}

		if offset += limit; offset > resp.TotalCount {
			return "", errAppNotFound.WithAttributes("app", name)
		}
	}
}

func (p *Source) getRootKeys(devEui string) (*csv4api.DeviceKeys, error) {
	client := csv4api.NewDeviceServiceClient(p.ClientConn)
	resp, err := client.GetKeys(context.Background(), &csv4api.GetDeviceKeysRequest{
		DevEui: devEui,
	})
	if err != nil {
		log.FromContext(p.ctx).WithField("dev_eui", devEui).WithError(err).Debug("No root keys")
		return nil, err
	}
	return resp.DeviceKeys, err
}

func (p *Source) getActivation(devEui string) (*csv4api.DeviceActivation, error) {
	client := csv4api.NewDeviceServiceClient(p.ClientConn)
	resp, err := client.GetActivation(context.Background(), &csv4api.GetDeviceActivationRequest{
		DevEui: devEui,
	})
	if err != nil {
		log.FromContext(p.ctx).WithField("dev_eui", devEui).WithError(err).Debug("No session keys")
		return nil, err
	}
	return resp.DeviceActivation, err
}
