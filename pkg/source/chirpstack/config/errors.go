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

package config

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errNoAPIToken      = errors.DefineInvalidArgument("no_api_token", "no API token")
	errNoAPIURL        = errors.DefineInvalidArgument("no_api_url", "no API URL")
	errNoFrequencyPlan = errors.DefineInvalidArgument("no_frequency_plan", "no Frequency Plan")

	errInvalidJoinEUI = errors.DefineInvalidArgument("invalid_join_eui", "invalid JoinEUI `{join_eui}`")
)
