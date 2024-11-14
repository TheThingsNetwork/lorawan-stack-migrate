package awsiot

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iotwireless/types"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Profile struct{ *types.LoRaWANDeviceProfile }

func splitMacVersion(v string) (string, string) {
	sep := " "
	split := strings.Split(v, sep)
	return split[0], strings.Join(split[1:], sep)
}

func (p Profile) macVersion() (mac ttnpb.MACVersion, phy ttnpb.PHYVersion, _ error) {
	regParamsRevision := aws.ToString(p.RegParamsRevision)

	switch v := aws.ToString(p.MacVersion); v {
	case "1.0.0":
		mac = ttnpb.MACVersion_MAC_V1_0
		phy = ttnpb.PHYVersion_TS001_V1_0
	case "1.0.1":
		mac = ttnpb.MACVersion_MAC_V1_0_1
		phy = ttnpb.PHYVersion_TS001_V1_0_1
	case "1.0.2":
		mac = ttnpb.MACVersion_MAC_V1_0_2
		switch regParamsRevision {
		case "A":
			phy = ttnpb.PHYVersion_RP001_V1_0_2
		case "B":
			phy = ttnpb.PHYVersion_RP001_V1_0_2_REV_B
		default:
			return mac, phy, errInvalidPHYForMACVersion.WithAttributes(
				"phy_version",
				regParamsRevision,
				"mac_version",
				v,
			)
		}
	case "1.0.3":
		mac = ttnpb.MACVersion_MAC_V1_0_3
		phy = ttnpb.PHYVersion_RP001_V1_0_3_REV_A
	case "1.0.4":
		mac = ttnpb.MACVersion_MAC_V1_0_4
		phy = ttnpb.PHYVersion_RP002_V1_0_4
	case "1.1.0":
		mac = ttnpb.MACVersion_MAC_V1_1
		switch regParamsRevision {
		case "A":
			phy = ttnpb.PHYVersion_RP001_V1_1_REV_A
		case "B":
			phy = ttnpb.PHYVersion_RP001_V1_1_REV_B
		default:
			return mac, phy, errInvalidPHYForMACVersion.WithAttributes(
				"phy_version",
				regParamsRevision,
				"mac_version",
				v,
			)
		}
	default:
		return mac, phy, errInvalidMACVersion.WithAttributes(
			"mac_version",
			v,
		)
	}
	return mac, phy, nil
}

func (p Profile) supportsJoin() bool {
	if v := p.SupportsJoin; v != nil {
		return *v
	}
	mode, _ := splitMacVersion(aws.ToString(p.MacVersion))
	return mode == "OTAA"
}

func (p Profile) SetFields(dev *ttnpb.EndDevice, fpStore *frequencyplans.Store) (err error) {
	dev.LorawanVersion, dev.LorawanPhyVersion, err = p.macVersion()
	if err != nil {
		return err
	}
	dev.SupportsClassB = p.SupportsClassB
	dev.SupportsClassC = p.SupportsClassC
	dev.SupportsJoin = p.supportsJoin()

	m := dev.MacSettings
	if m == nil {
		m = new(ttnpb.MACSettings)
	}

	if v := p.ClassBTimeout; v != nil {
		m.ClassBTimeout = durationpb.New(time.Duration(aws.ToInt32(v)))
	}
	if v := p.ClassCTimeout; v != nil {
		m.ClassCTimeout = durationpb.New(time.Duration(aws.ToInt32(v)))
	}
	if v := p.MaxDutyCycle; v != nil {
		m.DesiredMaxDutyCycle = &ttnpb.AggregatedDutyCycleValue{Value: ttnpb.AggregatedDutyCycle(aws.ToInt32(v))}
	}
	if v := p.MaxEirp; v != nil {
		m.DesiredMaxEirp = &ttnpb.DeviceEIRPValue{Value: ttnpb.DeviceEIRP(aws.ToInt32(v))}
	}
	if v := p.PingSlotDr; v != nil {
		m.PingSlotDataRateIndex = &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(aws.ToInt32(v))}
	}
	if v := p.PingSlotFreq; v != nil {
		m.PingSlotFrequency = &ttnpb.ZeroableFrequencyValue{Value: uint64(aws.ToInt32(v))}
	}
	if v := p.PingSlotPeriod; v != nil {
		m.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{Value: ttnpb.PingSlotPeriod(aws.ToInt32(v))}
	}
	if v := p.RxDataRate2; v != nil {
		m.Rx2DataRateIndex = &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(aws.ToInt32(v))}
	}
	if v := p.RxDelay1; v != nil {
		m.Rx1Delay = &ttnpb.RxDelayValue{Value: ttnpb.RxDelay(aws.ToInt32(v))}
	}
	if v := p.RxDrOffset1; v != nil {
		m.Rx1DataRateOffset = &ttnpb.DataRateOffsetValue{Value: ttnpb.DataRateOffset(aws.ToInt32(v))}
	}
	if v := p.RxFreq2; v != nil {
		m.Rx2Frequency = &ttnpb.FrequencyValue{Value: uint64(aws.ToInt32(v))}
	}
	m.Supports_32BitFCnt = &ttnpb.BoolValue{Value: p.Supports32BitFCnt}

	if dev.MacState, err = mac.NewState(dev, fpStore, dev.MacSettings); err != nil {
		return err
	}
	dev.MacState.CurrentParameters = dev.MacState.DesiredParameters

	return nil
}
