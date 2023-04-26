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

package tts

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errRead = errors.DefinePermissionDenied("read", "failed to read `{file}`")

	errDeviceIdentifiersMismatch = errors.Define("device_identifiers_mismatch", "device identifiers fields {field} do not match with values {a} and {b}")

	errNoAppID                        = errors.DefineInvalidArgument("no_app_id", "no app id")
	errNoAppAPIKey                    = errors.DefineInvalidArgument("no_app_api_key", "no app api key")
	errNoIdentityServerGRPCAddress    = errors.DefineInvalidArgument("no_identity_server_grpc_address", "no identity server grpc address")
	errNoJoinServerGRPCAddress        = errors.DefineInvalidArgument("no_join_server_grpc_address", "no join server grpc address")
	errNoApplicationServerGRPCAddress = errors.DefineInvalidArgument("no_application_server_grpc_address", "no application server grpc address")
	errNoNetworkServerGRPCAddress     = errors.DefineInvalidArgument("no_network_server_grpc_address", "no network server grpc address")
)
