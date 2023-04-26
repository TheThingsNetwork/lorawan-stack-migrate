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

package tts

import (
	"bytes"
	"strconv"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	getEndDeviceFromIS = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.EndDeviceRegistry/Get"].Allowed
	getEndDeviceFromNS = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.NsEndDeviceRegistry/Get"].Allowed
	getEndDeviceFromAS = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.AsEndDeviceRegistry/Get"].Allowed
	getEndDeviceFromJS = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.JsEndDeviceRegistry/Get"].Allowed

	claimAuthenticationCodePaths = []string{
		"claim_authentication_code",
		"claim_authentication_code.value",
		"claim_authentication_code.valid_from",
		"claim_authentication_code.valid_to",
	}
)

func validateDeviceIds(a, b *ttnpb.EndDeviceIdentifiers) error {
	if a == nil || b == nil {
		return nil
	}

	type pair struct {
		x, y []byte
		name string
	}
	pairs := []pair{
		{[]byte(a.DeviceId), []byte(b.DeviceId), "device_id"},
		{a.DevEui, b.DevEui, "dev_eui"},
		{a.JoinEui, b.JoinEui, "join_eui"},
		{a.DevAddr, b.DevAddr, "dev_addr"},
	}
	if x, y := a.ApplicationIds, b.ApplicationIds; x != nil && y != nil {
		pairs = append(pairs, pair{[]byte(x.ApplicationId), []byte(y.ApplicationId), "application_ids.application_id"})
	}

	isEmpty := func(s []byte) bool {
		return len(s) == 0
	}
	for _, s := range pairs {
		if isEmpty(s.x) || isEmpty(s.y) {
			continue
		}
		if bytes.Equal(s.x, s.y) {
			continue
		}
		return errDeviceIdentifiersMismatch.WithAttributes("field", s.name, "a", s.x, "b", s.y)
	}
	return nil
}

func nonImplicitPaths(paths ...string) []string {
	nonImplicitPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "ids" || strings.HasPrefix(path, "ids.") {
			continue
		}
		if path == "created_at" || path == "updated_at" {
			continue
		}
		nonImplicitPaths = append(nonImplicitPaths, path)
	}
	return nonImplicitPaths
}

func splitEndDeviceGetPaths(paths ...string) (is, ns, as, js []string) {
	is = ttnpb.AllowedFields(paths, getEndDeviceFromIS)
	ns = ttnpb.AllowedFields(paths, getEndDeviceFromNS)
	as = ttnpb.AllowedFields(paths, getEndDeviceFromAS)
	js = ttnpb.AllowedFields(paths, getEndDeviceFromJS)
	return
}

func updateDeviceTimestamps(dev, src *ttnpb.EndDevice) {
	if dev.CreatedAt == nil || (src.CreatedAt != nil && ttnpb.StdTime(src.CreatedAt).Before(*ttnpb.StdTime(dev.CreatedAt))) {
		dev.CreatedAt = src.CreatedAt
	}
	if dev.UpdatedAt == nil || (src.UpdatedAt != nil && ttnpb.StdTime(src.UpdatedAt).After(*ttnpb.StdTime(dev.UpdatedAt))) {
		dev.UpdatedAt = src.UpdatedAt
	}
}

func clearDeviceSession(dev *ttnpb.EndDevice) error {
	return dev.SetFields(nil,
		"activated_at",
		"mac_state",
		"last_dev_status_received_at",
		"last_seen_at",
		"pending_mac_state",
		"pending_session",
		"session",
	)
}

func withPagination() (limit, page uint32, opt grpc.CallOption, getTotal func() uint64) {
	limit = 50
	page = 1
	responseHeaders := metadata.MD{}
	opt = grpc.Header(&responseHeaders)
	getTotal = func() (total uint64) {
		totalHeader := responseHeaders.Get("x-total-count")
		if len(totalHeader) > 0 {
			total, _ = strconv.ParseUint(totalHeader[len(totalHeader)-1], 10, 64)
		}
		return total
	}
	return
}
