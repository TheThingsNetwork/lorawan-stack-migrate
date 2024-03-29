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

package export

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errExport                = errors.Define("export", "export device `{device_id}`")
	errFormat                = errors.DefineCorruption("format", "format device `{device_id}`")
	errInvalidFields         = errors.DefineInvalidArgument("invalid_fields", "invalid fields for device `{device_id}`")
	errDevIDExceedsMaxLength = errors.Define("dev_id_exceeds_max_length", "device ID `{id}` exceeds max length")
	errAppIDExceedsMaxLength = errors.Define("app_id_exceeds_max_length", "application ID `{id}` exceeds max length")
	errNoExportedIDorEUI     = errors.Define("no_exported_id_or_eui", "device `{device_id}` has no exported ID or EUI")
)
