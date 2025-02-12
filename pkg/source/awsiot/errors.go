// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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
	errInvalidMACVersion       = errors.DefineInvalidArgument("invalid_mac_version", "invalid MAC version `{mac_version}`")
	errInvalidPHYForMACVersion = errors.DefineInvalidArgument("invalid_phy_for_mac_version", "invalid PHY version `{phy_version}` for MAC version `{mac_version}`")
	errInvalidDevAddr          = errors.DefineInvalidArgument("invalid_dev_addr", "invalid DevAddr `{dev_addr}`")
	errInvalidDevEUI           = errors.DefineInvalidArgument("invalid_dev_eui", "invalid DevEUI `{dev_eui}`")
	errInvalidJoinEUI          = errors.DefineInvalidArgument("invalid_join_eui", "invalid JoinEUI `{join_eui}`")
	errInvalidKey              = errors.DefineInvalidArgument("invalid_key", "invalid key `{key}`")
)
