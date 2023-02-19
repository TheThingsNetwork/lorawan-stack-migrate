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

package ttnv3

import (
	"go.uber.org/zap"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/ttnv3/config"
)

var logger *zap.SugaredLogger

func init() {
	cfg, flags := config.New()

	logger, _ = config.NewLogger(cfg.Verbose)

	source.RegisterSource(source.Registration{
		Name:        "ttnv3",
		Description: "Migrate from The Things Stack",
		Flags:       flags,
		Create:      createNewSource(cfg),
	})
}
