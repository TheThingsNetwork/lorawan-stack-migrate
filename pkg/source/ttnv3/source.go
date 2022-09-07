package ttnv3

import (
	"context"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/ttnv3/api"
)

// Source implements the Source interface.
type Source struct {
	ctx context.Context

	config *config
}

// NewSource creates a new TTNv3 cource
func NewSource(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) {
	config, err := getConfig(flags)
	if err != nil {
		return Source{}, err
	}
	s := Source{
		ctx:    ctx,
		config: config,
	}
	return s, nil
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) {
	isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceGetPaths(ttnpb.BottomLevelFields(ttnpb.EndDeviceFieldPathsNested)...)
	if len(nsPaths) > 0 {
		isPaths = ttnpb.AddFields(isPaths, "network_server_address")
	}
	if len(asPaths) > 0 {
		isPaths = ttnpb.AddFields(isPaths, "application_server_address")
	}
	if len(jsPaths) > 0 {
		isPaths = ttnpb.AddFields(isPaths, "join_server_address")
	}
	is, err := api.Dial(s.ctx, s.config.identityServerGRPCAddress)
	if err != nil {
		return nil, err
	}
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: s.config.appID},
		DeviceId:       devID,
	}
	dev, err := ttnpb.NewEndDeviceRegistryClient(is).Get(s.ctx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIds: ids,
		FieldMask:    ttnpb.FieldMask(isPaths...),
	})
	if err != nil {
		return nil, err
	}
	if dev.ClaimAuthenticationCode.GetValue() != "" {
		// ClaimAuthenticationCode is already retrieved from the IS. We can unset the related JS paths
		jsPaths = ttnpb.ExcludeFields(jsPaths, claimAuthenticationCodePaths...)
	}
	res, err := s.getEndDevice(ids, nsPaths, asPaths, jsPaths, true)
	if err != nil {
		return nil, err
	}
	paths := ttnpb.AddFields(nsPaths, ttnpb.AddFields(asPaths, ttnpb.AddFields(jsPaths, "ids.dev_addr")...)...)
	if err := dev.SetFields(res, paths...); err != nil {
		return nil, err
	}
	updateDeviceTimestamps(dev, res)
	return dev, nil
}

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(appID string, f func(source.Source, string) error) error {
	is, err := api.Dial(s.ctx, s.config.identityServerGRPCAddress)
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
	if err := api.CloseConnections(); err != nil {
		return err
	}
	return nil
}

func (s Source) getEndDevice(ids *ttnpb.EndDeviceIdentifiers, nsPaths, asPaths, jsPaths []string, continueOnError bool) (*ttnpb.EndDevice, error) {
	var res ttnpb.EndDevice
	if len(jsPaths) > 0 {
		if s.config.joinServerGRPCAddress == "" {
			logger.WithField("paths", jsPaths).Warn("Join Server disabled but fields specified to get")
		} else {
			js, err := api.Dial(s.ctx, s.config.joinServerGRPCAddress)
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not connect to Join Server")
			} else {
				jsRes, err := ttnpb.NewJsEndDeviceRegistryClient(js).Get(s.ctx, &ttnpb.GetEndDeviceRequest{
					EndDeviceIds: ids,
					FieldMask:    ttnpb.FieldMask(jsPaths...),
				})
				if err != nil {
					if !continueOnError {
						return nil, err
					}
					logger.WithError(err).Error("Could not get end device from Join Server")
				} else {
					if err := res.SetFields(jsRes, ttnpb.AllowedReachableBottomLevelFields(jsPaths, getEndDeviceFromJS, jsRes.FieldIsZero)...); err != nil {
						return nil, err
					}
					updateDeviceTimestamps(&res, jsRes)
				}
			}
		}
	}
	if len(asPaths) > 0 {
		if s.config.applicationServerGRPCAddress == "" {
			logger.WithField("paths", asPaths).Warn("Application Server disabled but fields specified to get")
		} else {
			as, err := api.Dial(s.ctx, s.config.applicationServerGRPCAddress)
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not connect to Application Server")
			} else {
				asRes, err := ttnpb.NewAsEndDeviceRegistryClient(as).Get(s.ctx, &ttnpb.GetEndDeviceRequest{
					EndDeviceIds: ids,
					FieldMask:    ttnpb.FieldMask(asPaths...),
				})
				if err != nil {
					if !continueOnError {
						return nil, err
					}
					logger.WithError(err).Error("Could not get end device from Application Server")
				} else {
					if err := res.SetFields(asRes, ttnpb.AllowedReachableBottomLevelFields(asPaths, getEndDeviceFromAS, asRes.FieldIsZero)...); err != nil {
						return nil, err
					}
					updateDeviceTimestamps(&res, asRes)
				}
			}
		}
	}
	if len(nsPaths) > 0 {
		if s.config.networkServerGRPCAddress == "" {
			logger.WithField("paths", nsPaths).Warn("Network Server disabled but fields specified to get")
		} else {
			ns, err := api.Dial(s.ctx, s.config.networkServerGRPCAddress)
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not connect to Network Server")
			} else {
				nsRes, err := ttnpb.NewNsEndDeviceRegistryClient(ns).Get(s.ctx,
					&ttnpb.GetEndDeviceRequest{
						EndDeviceIds: ids,
						FieldMask:    ttnpb.FieldMask(nsPaths...),
					},
				)
				if err != nil {
					if !continueOnError {
						return nil, err
					}
					logger.WithError(err).Error("Could not get end device from Network Server")
				} else {
					if err := res.SetFields(nsRes, ttnpb.AllowedReachableBottomLevelFields(nsPaths, getEndDeviceFromNS, nsRes.FieldIsZero)...); err != nil {
						return nil, err
					}
					updateDeviceTimestamps(&res, nsRes)
				}
			}
		}
	}
	return &res, nil
}
