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

package ttnv2

import (
	"go.thethings.network/lorawan-stack-migrate/pkg/commands"
	_ "go.thethings.network/lorawan-stack-migrate/pkg/source/ttnv2"
)

const sourceName = "ttnv2"

// TTNv2Cmd represents the ttnv2 source.
var TTNv2Cmd = commands.Source(sourceName, "Export devices from TTN V2")
