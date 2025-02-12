package awsiot

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errNoValuesInCSV           = errors.DefineInvalidArgument("no_values_in_csv", "no values in CSV file")
	errNoAppID                 = errors.DefineInvalidArgument("no_app_id", "no app id")
	errNoCSVFileProvided       = errors.DefineInvalidArgument("no_csv_file_provided", "no csv file provided")
	errNoJoinEUI               = errors.DefineInvalidArgument("no_join_eui", "no join eui")
	errNoDeviceFound           = errors.DefineInvalidArgument("no_device_found", "no device with eui `{eui}` found")
	errNoFrequencyPlanID       = errors.DefineInvalidArgument("no_frequency_plan_id", "no frequency plan ID")
	errInvalidMACVersion       = errors.DefineInvalidArgument("invalid_mac_version", "invalid MAC version `{mac_version}`")
	errInvalidPHYForMACVersion = errors.DefineInvalidArgument("invalid_phy_for_mac_version", "invalid PHY version `{phy_version}` for MAC version `{mac_version}`")
)
