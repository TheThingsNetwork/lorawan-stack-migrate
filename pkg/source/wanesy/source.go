// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wanesy

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"

	"go.thethings.network/lorawan-stack-migrate/pkg/iterator"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/client"
)

type Source struct {
	*Config
	*client.Client

	imported Devices
}

func createNewSource(cfg *Config) source.CreateSource {
	return func(ctx context.Context, src source.Config) (source.Source, error) {
		if err := cfg.Initialize(src); err != nil {
			return nil, err
		}
		devs, err := cfg.ImportDevices()
		if err != nil {
			return nil, err
		}
		return Source{
			Config:   cfg,
			imported: devs,
		}, nil
	}
}

// Iterator implements source.Source.
func (s Source) Iterator(isApplication bool) iterator.Iterator {
	// if !isApplication {
	// 	return iterator.NewReaderIterator(os.Stdin, '\n')
	// }
	// if s.all {
	// 	// The Firefly LNS does not group devices by an application.
	// 	// When the "all" flag is set, we get all devices accessible by the API key.
	// 	// We use a dummy "all" App ID to fallthrough to the RangeDevices method,
	// 	// where the appID argument is unused.
	// 	return iterator.NewListIterator(
	// 		[]string{"all"},
	// 	)
	// }
	return iterator.NewNoopIterator()
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devEUIString string) (*ttnpb.EndDevice, error) {
	var devEUI, joinEUI types.EUI64
	if err := devEUI.UnmarshalText([]byte(devEUIString)); err != nil {
		return nil, err
	}
	wmcdev, ok := s.imported[devEUI]
	if !ok {
		return nil, errNoDeviceFound.WithAttributes("eui", devEUIString)
	}

	if err := joinEUI.UnmarshalText([]byte(wmcdev.AppEui)); err != nil {
		return nil, err
	}
	v3dev, err := wmcdev.EndDevice(s.fpStore, s.appID, s.frequencyPlanID)
	if err != nil {
		return nil, err
	}
	return v3dev, nil
}

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(_ string, f func(source.Source, string) error) error {
	var (
		devs []client.Device
		err  error
	)
	s.src.Logger.Debugw("Firefly LNS does not group devices by an application. Get all devices accessible by the API key")
	devs, err = s.GetAllDevices()
	if err != nil {
		return err
	}
	for _, d := range devs {
		if err := f(s, d.EUI); err != nil {
			return err
		}
	}
	return nil
}

// Close implements the Source interface.
func (s Source) Close() error { return nil }
