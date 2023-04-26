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
	"fmt"
	"io"
	"os"
	"strings"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func toJSON(dev *ttnpb.EndDevice) ([]byte, error) {
	return jsonpb.TTN().Marshal(dev)
}

const (
	maxIDLength = 36
)

var (
	errExport                = errors.Define("export", "export device `{device_id}`")
	errFormat                = errors.DefineCorruption("format", "format device `{device_id}`")
	errInvalidFields         = errors.DefineInvalidArgument("invalid_fields", "invalid fields for device `{device_id}`")
	errDevIDExceedsMaxLength = errors.Define("dev_id_exceeds_max_length", "device ID `{id}` exceeds max length")
	errAppIDExceedsMaxLength = errors.Define("app_id_exceeds_max_length", "application ID `{id}` exceeds max length")

	sanitizeID = strings.NewReplacer("_", "-")
)

type ExportConfig struct {
	EUIForID    bool
	DevIDPrefix string
}

func (cfg ExportConfig) ExportDev(s source.Source, devID string) error {
	dev, err := s.ExportDevice(devID)
	if err != nil {
		return errExport.WithAttributes("device_id", devID).WithCause(err)
	}
	oldID := dev.Ids.DeviceId

	if eui := dev.Ids.DevEui; cfg.EUIForID && eui != nil {
		dev.Ids.DeviceId = strings.ToLower(string(eui))
	}
	if cfg.DevIDPrefix != "" {
		dev.Ids.DeviceId = fmt.Sprintf("%s-%s", cfg.DevIDPrefix, dev.Ids.DeviceId)
	}

	dev.Ids.DeviceId = sanitizeID.Replace(dev.Ids.DeviceId)
	if id := dev.Ids.DeviceId; len(id) > maxIDLength {
		return errDevIDExceedsMaxLength.WithAttributes("id", id)
	}

	if dev.Ids.DeviceId != oldID {
		if dev.Attributes == nil {
			dev.Attributes = make(map[string]string)
		}
		dev.Attributes["old-id"] = oldID
	}

	dev.Ids.ApplicationIds.ApplicationId = sanitizeID.Replace(dev.Ids.ApplicationIds.ApplicationId)
	if id := dev.Ids.ApplicationIds.ApplicationId; len(id) > maxIDLength {
		return errAppIDExceedsMaxLength.WithAttributes("id", id)
	}

	if err := dev.ValidateFields(); err != nil {
		return errInvalidFields.WithAttributes(
			"device_id", dev.Ids.DeviceId,
			"dev_eui", dev.Ids.DevEui,
		).WithCause(err)
	}
	b, err := toJSON(dev)
	if err != nil {
		return errFormat.WithAttributes(
			"device_id", dev.Ids.DeviceId,
			"dev_eui", dev.Ids.DevEui,
		).WithCause(err)
	}
	_, err = fmt.Fprintln(os.Stdout, string(b))
	return err
}

// printStack prints the error stack to w.
func printStack(w io.Writer, err error) {
	for i, err := range errors.Stack(err) {
		if i == 0 {
			fmt.Fprintln(w, err)
		} else {
			fmt.Fprintf(w, "--- %s\n", err)
		}
		for k, v := range errors.Attributes(err) {
			fmt.Fprintf(os.Stderr, "    %s=%v\n", k, v)
		}
		if ttnErr, ok := errors.From(err); ok {
			if correlationID := ttnErr.CorrelationID(); correlationID != "" {
				fmt.Fprintf(os.Stderr, "    correlation_id=%s\n", ttnErr.CorrelationID())
			}
		}
	}
}
