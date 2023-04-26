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

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func installHook(name string) (err error) {
	fmt.Printf("Installing %s hook\n", name)
	f, err := os.OpenFile(filepath.Join(".git", "hooks", name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	_, err = fmt.Fprintf(f, "HOOK=\"%s\" ARGS=\"$@\" go run %s\n", name, filepath.Join(".hooks", "run-hooks.go"))
	if err != nil {
		return err
	}
	return nil
}

var gitHooks = []string{"commit-msg"}

// InstallHooks installs git hooks that help developers follow our best practices.
func InstallHooks() error {
	for _, hook := range gitHooks {
		if err := installHook(hook); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := InstallHooks(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
