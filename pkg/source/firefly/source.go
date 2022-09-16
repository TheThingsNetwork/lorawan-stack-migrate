package firefly

import (
	"context"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/api"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/devices"
)

type Source struct {
	ctx context.Context

	config *config
}

// NewSource creates a ner Firefly source
func NewSource(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) {
	config, err := getConfig(flags)
	if err != nil {
		return Source{}, err
	}
	return Source{
		ctx: ctx,

		config: config,
	}, nil
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
		FrequencyPlanId: s.config.frequencyPlanID,
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: s.config.appID},
			DeviceId:       "eui-" + strings.ToLower(ffdev.EUI),
		},
		MacSettings: &ttnpb.MACSettings{
			Rx2DataRateIndex:       &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(ffdev.Rx2DataRate)},
			StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
			StatusTimePeriodicity:  pbtypes.DurationProto(0),
		},
		// MacState:       &ttnpb.MACState{},
		RootKeys:       &ttnpb.RootKeys{},
		SupportsClassC: ffdev.ClassC,
		SupportsJoin:   ffdev.ApplicationKey != "",
	}
	v3dev.Attributes = make(map[string]string)
	v3dev.Ids.DevEui, err = unmarshalTextToBytes(&types.EUI64{}, ffdev.EUI)
	if err != nil {
		return nil, err
	}
	v3dev.Ids.JoinEui, err = unmarshalTextToBytes(&types.EUI64{}, s.config.joinEUI)
	if err != nil {
		return nil, err
	}
	if v3dev.SupportsJoin {
		v3dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{}
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

	if !s.config.dryRun {
		log.FromContext(s.ctx).WithFields(log.Fields(
			"device_id", ffdev.Name,
			"device_eui", ffdev.EUI,
		)).Info("Clearing device keys")
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
	logger.WithField("app-id", appID).Debug("App ID")
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
	logger.WithField("devices", devs).Debug("Devices List")
	for _, d := range devs {
		if err := f(s, d.EUI); err != nil {
			return err
		}
	}
	return nil
}

// Close implements the Source interface.
func (s Source) Close() error { return nil }
