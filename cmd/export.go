// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package cmd

import (
	"io"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

func exportCommand(cmd *cobra.Command, args []string, f func(s source.Source, item string) error) error {
	var iter Iterator
	switch len(args) {
	case 0:
		iter = NewEmptyIterator()
	default:
		iter = NewListIterator(args)
	}

	s, err := source.NewSource(ctx, cmd.Flags())
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			logger.WithError(err).Fatal("Failed to clean up")
		}
	}()

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
