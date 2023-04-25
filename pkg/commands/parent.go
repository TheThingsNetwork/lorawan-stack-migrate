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

import "github.com/spf13/cobra"

func ExecuteParentPersistentPreRun(cmd *cobra.Command, args []string) error {
	if !cmd.HasParent() {
		return nil
	}
	p := cmd.Parent()

	if f := p.PersistentPreRunE; f != nil {
		if err := f(p, args); err != nil {
			return err
		}
	} else if f := p.PersistentPreRun; f != nil {
		f(p, args)
	}
	return nil
}

func ExecuteParentPersistentPostRun(cmd *cobra.Command, args []string) error {
	if !cmd.HasParent() {
		return nil
	}
	p := cmd.Parent()

	if f := p.PersistentPostRunE; f != nil {
		if err := f(p, args); err != nil {
			return err
		}
	} else if f := p.PersistentPostRun; f != nil {
		f(p, args)
	}
	return nil
}
