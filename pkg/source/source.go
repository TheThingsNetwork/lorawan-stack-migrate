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

package source

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type Config struct {
	DryRun, Verbose bool
	FrequencyPlansURL,
	Source string
}

var RootConfig Config

// Source is a source for end devices.
type Source interface {
	// ExportDevice retrieves an end device from the source and returns it as a ttnpb.EndDevice.
	ExportDevice(devID string) (*ttnpb.EndDevice, error)
	// RangeDevices calls a function for all matching devices of an application.
	RangeDevices(appID string, f func(s Source, devID string) error) error
	// Close cleans up and terminates any open connections.
	Close() error
}

// CreateSource is a function that constructs a new Source.
type CreateSource func(ctx context.Context, rootCfg Config) (Source, error)

// Registration contains information for a registered Source.
type Registration struct {
	Name,
	Description string

	Create  CreateSource
	FlagSet *pflag.FlagSet
}

var (
	errNotRegistered     = errors.DefineInvalidArgument("not_registered", "source `{source}` is not registered")
	errAlreadyRegistered = errors.DefineInvalidArgument("already_registered", "source `{source}` is already registered")
	errNoSource          = errors.DefineInvalidArgument("no_source", "no source")

	registeredSources map[string]Registration
)

// RegisterSource registers a new Source.
func RegisterSource(r Registration) error {
	if _, ok := registeredSources[r.Name]; ok {
		return errAlreadyRegistered.WithAttributes("source", r.Name)
	}
	registeredSources[r.Name] = r
	return nil
}

// NewSource creates a new Source from parsed flags.
func NewSource(ctx context.Context) (Source, error) {
	if RootConfig.Source == "" {
		return nil, errNoSource.New()
	}
	if registration, ok := registeredSources[RootConfig.Source]; ok {
		return registration.Create(ctx, RootConfig)
	}
	return nil, errNotRegistered.WithAttributes("source", RootConfig.Source)
}

func addPrefix(name, prefix string) string {
	if prefix == "" {
		return name
	}
	// Append separator
	prefix += "."
	// Remove prefix if already present
	name = strings.TrimPrefix(name, prefix)
	return prefix + name
}

// AllFlagSets returns flags for all configured sources prefixed with source names.
func AllFlagSets() *pflag.FlagSet {
	fs := new(pflag.FlagSet)
	for _, r := range registeredSources {
		if r.FlagSet == nil {
			continue
		}
		r.FlagSet.VisitAll(func(a *pflag.Flag) {
			b := *a // Avoid modifying source flags
			b.Name = addPrefix(b.Name, r.Name)
			b.Annotations = make(map[string][]string)
			for k, v := range a.Annotations {
				b.Annotations[k] = v
			}
			fs.AddFlag(&b)
		})
	}
	fs.StringVar(&RootConfig.Source,
		"source",
		"",
		fmt.Sprintf("source (%s)", strings.Join(Names(), "|")),
	)
	return fs
}

// FlagSet returns a flag set for a given source.
func FlagSet(s string) (*pflag.FlagSet, error) {
	src, ok := registeredSources[s]
	if !ok {
		return nil, errNotRegistered.WithAttributes("source", s)
	}
	return src.FlagSet, nil
}

// Sources returns a map of registered Sources and their descriptions.
func Sources() map[string]string {
	sources := make(map[string]string)
	for _, registration := range registeredSources {
		sources[registration.Name] = registration.Description
	}
	return sources
}

// Names returns a slice of registered Sources names.
func Names() []string {
	var names []string
	for k := range registeredSources {
		names = append(names, k)
	}
	return names
}

func init() {
	registeredSources = make(map[string]Registration)
}
