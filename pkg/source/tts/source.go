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

package tts

import (
	"context"
	"os"

	"go.thethings.network/lorawan-stack-migrate/pkg/iterator"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/tts/api"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/tts/config"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Source implements the Source interface.
type Source struct {
	ctx context.Context

	config *config.Config
}

func createNewSource(cfg *config.Config) source.CreateSource {
	return func(ctx context.Context, rootCfg source.Config) (source.Source, error) {
		return Source{
			ctx:    ctx,
			config: cfg,
		}, cfg.Initialize(rootCfg)
	}
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) {
	if s.config.AppID == "" {
		return nil, errNoAppID.New()
	}

	basePaths := ttnpb.EndDeviceFieldPathsNested

	if !s.config.ExportCACs {
		basePaths = ttnpb.ExcludeFields(basePaths, claimAuthenticationCodePaths...)
	}

	isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceGetPaths(ttnpb.BottomLevelFields(basePaths)...)
	if len(nsPaths) > 0 {
		isPaths = ttnpb.AddFields(isPaths, "network_server_address")
	}
	if len(asPaths) > 0 {
		isPaths = ttnpb.AddFields(isPaths, "application_server_address")
	}
	if len(jsPaths) > 0 {
		isPaths = ttnpb.AddFields(isPaths, "join_server_address")
	}
	is, err := api.Dial(s.ctx, s.config.ServerConfig.IdentityServerGRPCAddress)
	if err != nil {
		return nil, err
	}
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: s.config.AppID},
		DeviceId:       devID,
	}
	dev, err := ttnpb.NewEndDeviceRegistryClient(is).Get(s.ctx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIds: ids,
		FieldMask:    ttnpb.FieldMask(isPaths...),
	})
	if err != nil {
		return nil, err
	}
	if dev.ClaimAuthenticationCode.GetValue() != "" || !s.config.ExportCACs {
		// ClaimAuthenticationCode is already retrieved from the IS. We can unset the related JS paths
		jsPaths = ttnpb.ExcludeFields(jsPaths, claimAuthenticationCodePaths...)
	}
	res, err := s.getEndDevice(ids, nsPaths, asPaths, jsPaths)
	if err != nil {
		return nil, err
	}
	if err := validateDeviceIds(dev.Ids, res.Ids); err != nil {
		return nil, err
	}
	paths := ttnpb.AddFields(nsPaths, ttnpb.AddFields(asPaths, ttnpb.AddFields(jsPaths, "ids.dev_addr")...)...)
	if err := dev.SetFields(res, paths...); err != nil {
		return nil, err
	}
	// Clear all "{server}_address" fields from exported device
	if err := dev.SetFields(nil, "application_server_address", "join_server_address", "network_server_address"); err != nil {
		return nil, err
	}

	if s.config.NoSession {
		if err := clearDeviceSession(dev); err != nil {
			return nil, err
		}
	}
	switch {
	case s.config.DeleteSourceDevice:
		if err := s.deleteEndDevice(dev.GetIds()); err != nil {
			return nil, err
		}
	case !s.config.DryRun:
		d := &ttnpb.EndDevice{
			Ids:         dev.GetIds(),
			MacSettings: &ttnpb.MACSettings{ScheduleDownlinks: &ttnpb.BoolValue{Value: false}},
		}
		nsPaths := []string{"mac_settings.schedule_downlinks.value"}
		if _, err := s.setEndDevice(d, nil, nsPaths, nil, nil, nil); err != nil {
			return nil, err
		}
	}

	updateDeviceTimestamps(dev, res)
	return dev, nil
}

// Iterator implements source.Source.
func (s Source) Iterator(bool) iterator.Iterator {
	return iterator.NewReaderIterator(os.Stdin, '\n')
}

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(appID string, f func(source.Source, string) error) error {
	s.config.AppID = appID

	is, err := api.Dial(s.ctx, s.config.ServerConfig.IdentityServerGRPCAddress)
	if err != nil {
		return err
	}
	limit, page, opt, getTotal := withPagination()
	for {
		res, err := ttnpb.NewEndDeviceRegistryClient(is).List(s.ctx, &ttnpb.ListEndDevicesRequest{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: appID},
			FieldMask:      ttnpb.FieldMask("ids.device_id"),
			Limit:          limit,
			Page:           page,
		}, opt)
		if err != nil {
			return err
		}
		for _, d := range res.EndDevices {
			if err := f(s, d.Ids.DeviceId); err != nil {
				return err
			}
		}
		if total := getTotal(); uint64(page)*uint64(limit) >= total {
			break
		}
		page++
	}
	return nil
}

// Close implements the Source interface.
func (s Source) Close() error {
	api.CloseAll()
	return nil
}

func (s Source) getEndDevice(ids *ttnpb.EndDeviceIdentifiers, nsPaths, asPaths, jsPaths []string) (*ttnpb.EndDevice, error) {
	res := &ttnpb.EndDevice{}
	if len(jsPaths) > 0 {
		if s.config.ServerConfig.JoinServerGRPCAddress == "" {
			s.config.Logger.With("paths", jsPaths).Warn("Join Server disabled but fields specified to get")
		} else {
			js, err := api.Dial(s.ctx, s.config.ServerConfig.JoinServerGRPCAddress)
			if err != nil {
				return nil, err
			}
			jsRes, err := ttnpb.NewJsEndDeviceRegistryClient(js).Get(s.ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: ids,
				FieldMask:    ttnpb.FieldMask(jsPaths...),
			})
			if err != nil {
				s.config.Logger.With("error", err).Warn("Could not get end device from Join Server")
			} else {
				if err := validateDeviceIds(res.Ids, jsRes.Ids); err != nil {
					return nil, err
				}
				if err := res.SetFields(jsRes, ttnpb.AllowedReachableBottomLevelFields(jsPaths, getEndDeviceFromJS, jsRes.FieldIsZero)...); err != nil {
					return nil, err
				}
				updateDeviceTimestamps(res, jsRes)
			}
		}
	}
	if len(asPaths) > 0 {
		if s.config.ServerConfig.ApplicationServerGRPCAddress == "" {
			s.config.Logger.With("paths", asPaths).Warn("Application Server disabled but fields specified to get")
		} else {
			as, err := api.Dial(s.ctx, s.config.ServerConfig.ApplicationServerGRPCAddress)
			if err != nil {
				return nil, err
			}
			asRes, err := ttnpb.NewAsEndDeviceRegistryClient(as).Get(s.ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: ids,
				FieldMask:    ttnpb.FieldMask(asPaths...),
			})
			if err != nil {
				return nil, err
			}
			if err := validateDeviceIds(res.Ids, asRes.Ids); err != nil {
				return nil, err
			}
			if err := res.SetFields(asRes, ttnpb.AllowedReachableBottomLevelFields(asPaths, getEndDeviceFromAS, asRes.FieldIsZero)...); err != nil {
				return nil, err
			}
			updateDeviceTimestamps(res, asRes)
		}
	}
	if len(nsPaths) > 0 {
		if s.config.ServerConfig.NetworkServerGRPCAddress == "" {
			s.config.Logger.With("paths", nsPaths).Warn("Network Server disabled but fields specified to get")
		} else {
			ns, err := api.Dial(s.ctx, s.config.ServerConfig.NetworkServerGRPCAddress)
			if err != nil {
				return nil, err
			}
			nsRes, err := ttnpb.NewNsEndDeviceRegistryClient(ns).Get(s.ctx,
				&ttnpb.GetEndDeviceRequest{
					EndDeviceIds: ids,
					FieldMask:    ttnpb.FieldMask(nsPaths...),
				},
			)
			if err != nil {
				return nil, err
			}
			if err := validateDeviceIds(res.Ids, nsRes.Ids); err != nil {
				return nil, err
			}
			if err := res.SetFields(nsRes, ttnpb.AllowedReachableBottomLevelFields(nsPaths, getEndDeviceFromNS, nsRes.FieldIsZero)...); err != nil {
				return nil, err
			}
			updateDeviceTimestamps(res, nsRes)
		}
	}
	return res, nil
}

func (s Source) setEndDevice(device *ttnpb.EndDevice, isPaths, nsPaths, asPaths, jsPaths, unsetPaths []string) (*ttnpb.EndDevice, error) {
	var res ttnpb.EndDevice
	if err := res.SetFields(device, "ids", "created_at", "updated_at"); err != nil {
		return nil, err
	}
	if len(isPaths) > 0 {
		is, err := api.Dial(s.ctx, s.config.ServerConfig.IdentityServerGRPCAddress)
		if err != nil {
			return nil, err
		}
		isDevice := &ttnpb.EndDevice{}
		s.config.Logger.With("paths", isPaths).Debug("Set end device on Identity Server")
		isDevice.SetFields(device, ttnpb.AddFields(ttnpb.ExcludeFields(isPaths, unsetPaths...), "ids")...)
		isRes, err := ttnpb.NewEndDeviceRegistryClient(is).Update(s.ctx, &ttnpb.UpdateEndDeviceRequest{
			EndDevice: isDevice,
			FieldMask: ttnpb.FieldMask(isPaths...),
		})
		if err != nil {
			return nil, err
		}
		if err := res.SetFields(isRes, isPaths...); err != nil {
			return nil, err
		}
		updateDeviceTimestamps(&res, isRes)
	}
	if len(jsPaths) > 0 {
		if s.config.ServerConfig.JoinServerGRPCAddress == "" {
			s.config.Logger.With("paths", jsPaths).Warn("Join Server disabled but fields specified to set")
		} else {
			js, err := api.Dial(s.ctx, s.config.ServerConfig.JoinServerGRPCAddress)
			if err != nil {
				return nil, err
			}
			jsDevice := &ttnpb.EndDevice{}
			s.config.Logger.With("paths", jsPaths).Debug("Set end device on Join Server")
			if err := jsDevice.SetFields(device, ttnpb.AddFields(ttnpb.ExcludeFields(jsPaths, unsetPaths...), "ids")...); err != nil {
				return nil, err
			}
			jsRes, err := ttnpb.NewJsEndDeviceRegistryClient(js).Set(s.ctx, &ttnpb.SetEndDeviceRequest{
				EndDevice: jsDevice,
				FieldMask: ttnpb.FieldMask(jsPaths...),
			})
			if err != nil {
				return nil, err
			}
			if err := res.SetFields(jsDevice, jsPaths...); err != nil {
				return nil, err
			}
			updateDeviceTimestamps(&res, jsRes)
		}
	}
	if len(nsPaths) > 0 {
		if s.config.ServerConfig.NetworkServerGRPCAddress == "" {
			s.config.Logger.With("paths", nsPaths).Warn("Network Server disabled but fields specified to set")
		} else {
			ns, err := api.Dial(s.ctx, s.config.ServerConfig.NetworkServerGRPCAddress)
			if err != nil {
				return nil, err
			}
			nsDevice := &ttnpb.EndDevice{}
			s.config.Logger.With("paths", nsPaths).Debug("Set end device on Network Server")
			if err := nsDevice.SetFields(device, ttnpb.AddFields(ttnpb.ExcludeFields(nsPaths, unsetPaths...), "ids")...); err != nil {
				return nil, err
			}
			nsRes, err := ttnpb.NewNsEndDeviceRegistryClient(ns).Set(s.ctx, &ttnpb.SetEndDeviceRequest{
				EndDevice: nsDevice,
				FieldMask: ttnpb.FieldMask(nsPaths...),
			})
			if err != nil {
				return nil, err
			}
			if err := res.SetFields(nsRes, nsPaths...); err != nil {
				return nil, err
			}
			updateDeviceTimestamps(&res, nsRes)
		}
	}
	if len(asPaths) > 0 {
		if s.config.ServerConfig.ApplicationServerGRPCAddress == "" {
			s.config.Logger.With("paths", asPaths).Warn("Application Server disabled but fields specified to set")
		} else {
			as, err := api.Dial(s.ctx, s.config.ServerConfig.ApplicationServerGRPCAddress)
			if err != nil {
				return nil, err
			}
			asDevice := &ttnpb.EndDevice{}
			s.config.Logger.With("paths", asPaths).Debug("Set end device on Application Server")
			if err := asDevice.SetFields(device, ttnpb.AddFields(ttnpb.ExcludeFields(asPaths, unsetPaths...), "ids")...); err != nil {
				return nil, err
			}
			asRes, err := ttnpb.NewAsEndDeviceRegistryClient(as).Set(s.ctx, &ttnpb.SetEndDeviceRequest{
				EndDevice: asDevice,
				FieldMask: ttnpb.FieldMask(asPaths...),
			})
			if err != nil {
				return nil, err
			}
			updateDeviceTimestamps(&res, asRes)
		}
	}
	return &res, nil
}

func (s Source) deleteEndDevice(ids *ttnpb.EndDeviceIdentifiers) error {
	if address := s.config.ServerConfig.ApplicationServerGRPCAddress; address != "" {
		as, err := api.Dial(s.ctx, address)
		if err != nil {
			return err
		}
		if _, err = ttnpb.NewAsEndDeviceRegistryClient(as).Delete(s.ctx, ids); err != nil {
			return err
		}
	}
	if address := s.config.ServerConfig.NetworkServerGRPCAddress; address != "" {
		ns, err := api.Dial(s.ctx, address)
		if err != nil {
			return err
		}
		if _, err = ttnpb.NewNsEndDeviceRegistryClient(ns).Delete(s.ctx, ids); err != nil {
			return err
		}
	}
	if address := s.config.ServerConfig.JoinServerGRPCAddress; address != "" {
		js, err := api.Dial(s.ctx, address)
		if err != nil {
			return err
		}
		if _, err = ttnpb.NewJsEndDeviceRegistryClient(js).Delete(s.ctx, ids); err != nil {
			return err
		}
	}
	if address := s.config.ServerConfig.IdentityServerGRPCAddress; address != "" {
		is, err := api.Dial(s.ctx, address)
		if err != nil {
			return err
		}
		if _, err = ttnpb.NewEndDeviceRegistryClient(is).Delete(s.ctx, ids); err != nil {
			return err
		}
	}

	return nil
}
