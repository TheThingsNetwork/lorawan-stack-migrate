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

package config

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errNoAppID           = errors.DefineInvalidArgument("no_app_id", "no app id")
	errNoFrequencyPlanID = errors.DefineInvalidArgument("no_frequency_plan_id", "no frequency plan ID")
)
