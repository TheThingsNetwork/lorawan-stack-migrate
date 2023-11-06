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

package commands

import (
	"io"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack-migrate/pkg/export"
	"go.thethings.network/lorawan-stack-migrate/pkg/iterator"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

func Export(cmd *cobra.Command, args []string, f func(s source.Source, item string) error) error {
	s, err := source.NewSource(cmd.Context())
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.FromContext(cmd.Context()).WithError(err).Fatal("Failed to clean up")
		}
	}()

	var iter iterator.Iterator
	switch len(args) {
	case 0:
		iter = s.Iterator()
	default:
		iter = iterator.NewListIterator(args)
	}

	for {
		item, err := iter.Next()
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}
		if item == "" {
			continue
		}

		if err := f(s, item); err != nil {
			return err
		}
	}
}

func ExportApplication() CobraRunE {
	return func(cmd *cobra.Command, args []string) error {
		return Export(cmd, args, func(s source.Source, item string) error {
			return s.RangeDevices(item, export.FromContext(cmd.Context()).ExportDev)
		})
	}
}

func ExportDevices() CobraRunE {
	return func(cmd *cobra.Command, args []string) error {
		return Export(cmd, args, export.FromContext(cmd.Context()).ExportDev)
	}
}
