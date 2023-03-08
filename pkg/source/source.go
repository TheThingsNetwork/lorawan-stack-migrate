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
type CreateSource func(ctx context.Context, flags *pflag.FlagSet) (Source, error)

// Registration contains information for a registered Source.
type Registration struct {
	Name,
	Description string

	Create CreateSource
	FlagSet,
	ApplicationFlagSet,
	DevicesFlagSet *pflag.FlagSet
}

var (
	errNotRegistered     = errors.DefineInvalidArgument("not_registered", "source `{source}` is not registered")
	errAlreadyRegistered = errors.DefineInvalidArgument("already_registered", "source `{source}` is already registered")
	errNoSource          = errors.DefineInvalidArgument("no_source", "no source")

	registeredSources map[string]Registration
	activeSource      string
)

// SetActiveSource changes the active source to s, if it exists.
func SetActiveSource(s string) bool {
	if _, found := registeredSources[s]; found {
		activeSource = s
		return true
	}
	return false
}

// RegisterSource registers a new Source.
func RegisterSource(r Registration) error {
	if _, ok := registeredSources[r.Name]; ok {
		return errAlreadyRegistered.WithAttributes("source", r.Name)
	}
	registeredSources[r.Name] = r
	return nil
}

// NewSource creates a new Source from parsed flags.
func NewSource(ctx context.Context, flags *pflag.FlagSet) (Source, error) {
	if activeSource == "" {
		return nil, errNoSource.New()
	}
	if registration, ok := registeredSources[activeSource]; ok {
		return registration.Create(ctx, flags)
	}
	return nil, errNotRegistered.WithAttributes("source", activeSource)
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

// FlagSet returns common flags for ActiveSource.
// If there is no active source returns flags for all configured sources.
func FlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	switch activeSource {
	case "":
		for _, r := range registeredSources {
			fs := &pflag.FlagSet{}
			// Append source names to flag names to prevent duplicate names
			r.FlagSet.VisitAll(func(f *pflag.Flag) {
				f.Name = addPrefix(f.Name, strings.ToLower(r.Name))
				fs.AddFlag(f)
			})
			flags.AddFlagSet(fs)
		}
		flags.StringVar(&activeSource, "source", "", fmt.Sprintf("source (%s)", strings.Join(Names(), "|")))
		flags.MarkDeprecated("source",
			fmt.Sprintf("use positional `source` argument instead `ttn-lw-migrate (%s) [command] ... [flags]`",
				strings.Join(Names(), "|"),
			),
		)

	default:
		r := registeredSources[activeSource]
		flags.AddFlagSet(r.FlagSet)
	}

	return flags
}

// ApplicationFlagSet returns `application` command flags for ActiveSource.
// If there is no active source returns flags for all configured sources.
func ApplicationFlagSet() *pflag.FlagSet {
	flags := FlagSet()

	switch activeSource {
	case "":
		for _, r := range registeredSources {
			if r.ApplicationFlagSet == nil {
				continue
			}
			fs := &pflag.FlagSet{}
			// Append source names to flag names to prevent duplicate names
			r.ApplicationFlagSet.VisitAll(func(f *pflag.Flag) {
				f.Name = addPrefix(f.Name, strings.ToLower(r.Name))
				fs.AddFlag(f)
			})
			flags.AddFlagSet(fs)
		}

	default:
		r := registeredSources[activeSource]
		flags.AddFlagSet(r.ApplicationFlagSet)
	}

	return flags
}

// DevicesFlagSet returns `devices` command flags for ActiveSource.
// If there is no active source returns flags for all configured sources.
func DevicesFlagSet() *pflag.FlagSet {
	flags := FlagSet()

	switch activeSource {
	case "":
		for _, r := range registeredSources {
			if r.DevicesFlagSet == nil {
				continue
			}
			fs := &pflag.FlagSet{}
			// Append source names to flag names to prevent duplicate names
			r.DevicesFlagSet.VisitAll(func(f *pflag.Flag) {
				f.Name = addPrefix(f.Name, strings.ToLower(r.Name))
				fs.AddFlag(f)
			})
			flags.AddFlagSet(fs)
		}

	default:
		r := registeredSources[activeSource]
		flags.AddFlagSet(r.DevicesFlagSet)
	}

	return flags
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
