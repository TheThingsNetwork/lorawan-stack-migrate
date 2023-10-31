package firefly

import (
	"context"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/api"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/devices"
)

type Source struct {
	*Config

	ctx context.Context
}

func createNewSource(cfg *Config) source.CreateSource {
	return func(ctx context.Context, src source.Config) (source.Source, error) {
		cfg.Config = src
		if err := cfg.Initialize(); err != nil {
			return nil, err
		}
		return Source{
			ctx:    ctx,
			Config: cfg,
		}, nil
	}
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devEUI string) (*ttnpb.EndDevice, error) {
	ffdev, err := devices.GetDevice(devEUI)
	if err != nil {
		return nil, err
	}
	v3dev := &ttnpb.EndDevice{
		Name:        ffdev.Name,
		Description: ffdev.Description,
		// Formatters:      &ttnpb.MessagePayloadFormatters{},
		FrequencyPlanId: s.frequencyPlanID,
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: s.appID},
			DeviceId:       "eui-" + strings.ToLower(ffdev.EUI),
		},
		MacSettings: &ttnpb.MACSettings{
			DesiredAdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADRAckLimitExponent(ffdev.AdrLimit)},
			Rx2DataRateIndex:           &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(ffdev.Rx2DataRate)},
			StatusCountPeriodicity:     wrapperspb.UInt32(0),
			StatusTimePeriodicity:      durationpb.New(0),
		},
		SupportsClassC: ffdev.ClassC,
		SupportsJoin:   ffdev.ApplicationKey != "",
	}

	if ffdev.Location != nil {
		v3dev.Locations = map[string]*ttnpb.Location{
			"user": {
				Latitude:  ffdev.Location.Lattitude,
				Longitude: ffdev.Location.Longitude,
				Source:    ttnpb.LocationSource_SOURCE_REGISTRY,
			},
		}
		logger.Debugw("Set location", "location", v3dev.Locations)
	}
	v3dev.Ids.DevEui, err = unmarshalTextToBytes(&types.EUI64{}, ffdev.EUI)
	if err != nil {
		return nil, err
	}
	v3dev.Ids.JoinEui, err = unmarshalTextToBytes(&types.EUI64{}, s.joinEUI)
	if err != nil {
		return nil, err
	}
	if v3dev.SupportsJoin {
		v3dev.RootKeys = &ttnpb.RootKeys{AppKey: &ttnpb.KeyEnvelope{}}
		v3dev.RootKeys.AppKey.Key, err = unmarshalTextToBytes(&types.AES128Key{}, ffdev.ApplicationKey)
		if err != nil {
			return nil, err
		}
	}
	hasSession := ffdev.Address != "" && ffdev.NetworkSessionKey != "" && ffdev.ApplicationSessionKey != ""
	if hasSession || !v3dev.SupportsJoin {
		v3dev.Session = &ttnpb.Session{Keys: &ttnpb.SessionKeys{AppSKey: &ttnpb.KeyEnvelope{}, NwkSEncKey: &ttnpb.KeyEnvelope{}}}
		v3dev.Session.DevAddr, err = unmarshalTextToBytes(&types.DevAddr{}, ffdev.Address)
		if err != nil {
			return nil, err
		}
		v3dev.Session.Keys.AppSKey.Key, err = unmarshalTextToBytes(&types.AES128Key{}, ffdev.ApplicationSessionKey)
		if err != nil {
			return nil, err
		}
		v3dev.Session.Keys.NwkSEncKey.Key, err = unmarshalTextToBytes(&types.AES128Key{}, ffdev.NetworkSessionKey)
		if err != nil {
			return nil, err
		}
		v3dev.Session.LastAFCntDown = uint32(ffdev.FrameCounter)
		v3dev.Session.LastNFCntDown = uint32(ffdev.FrameCounter)
		packet, err := devices.GetLastPacket()
		if err != nil {
			return nil, err
		}
		v3dev.Session.LastFCntUp = uint32(packet.FCnt)
	}

	if !s.DryRun {
		logger.Debugw("Clearing device keys", "device_id", ffdev.Name, "device_eui", ffdev.EUI)
		r, err := api.PutDeviceUpdate(devEUI, map[string]string{
			"address": "", "application_key": "", "application_session_key": "",
		})
		if err != nil {
			return nil, err
		}
		r.Body.Close()
	}

	return v3dev, nil
}

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(appID string, f func(source.Source, string) error) error {
	// req, err := http.Get(fmt.Sprintf("http://%s/api/v1/applications/%s/euis?auth=%s", s.config.apiURL, appID, s.config.apiKey))
	var (
		devs []devices.Device
		err  error
	)
	logger.With("app-id", appID).Debug("App ID")
	switch {
	case appID != "":
		devs, err = devices.GetDeviceListByAppID(appID)
		if err != nil {
			return err
		}
	default:
		devs, err = devices.GetAllDevices()
		if err != nil {
			return err
		}
	}
	logger.With("devices", devs).Debug("Devices List")
	for _, d := range devs {
		if err := f(s, d.EUI); err != nil {
			return err
		}
	}
	return nil
}

// Close implements the Source interface.
func (s Source) Close() error { return nil }
