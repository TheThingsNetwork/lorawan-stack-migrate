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
