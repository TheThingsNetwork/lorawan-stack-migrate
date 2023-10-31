// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package client

import "encoding/hex"

// Location is the location of a device.
type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// Device is a Firefly device.
type Device struct {
	Address               string    `json:"address,omitempty"`
	AdrLimit              int       `json:"adr_limit,omitempty"`
	ApplicationKey        string    `json:"application_key,omitempty"`
	ApplicationSessionKey string    `json:"application_session_key,omitempty"`
	ClassC                bool      `json:"class_c,omitempty"`
	Deduplicate           bool      `json:"deduplicate,omitempty"`
	Description           string    `json:"description,omitempty"`
	DeviceClassID         int       `json:"device_class_id,omitempty"`
	EUI                   string    `json:"eui,omitempty"`
	FrameCounter          int       `json:"frame_counter,omitempty"`
	InsertedAt            string    `json:"inserted_at,omitempty"`
	Location              *Location `json:"location,omitempty"`
	Name                  string    `json:"name,omitempty"`
	NetworkSessionKey     string    `json:"network_session_key,omitempty"`
	OTAA                  bool      `json:"otaa,omitempty"`
	OrganizationID        int       `json:"organization_id,omitempty"`
	OverrideLocation      bool      `json:"override_location,omitempty"`
	Region                string    `json:"region,omitempty"`
	Rx2DataRate           int       `json:"rx2_data_rate,omitempty"`
	SkipFCntCheck         bool      `json:"skip_fcnt_check,omitempty"`
	Tags                  []string  `json:"tags,omitempty"`
	UpdatedAt             string    `json:"updated_at,omitempty"`
}

// WithIncrementedKeys returns the device with last byte of the keys incremented.
func (d Device) WithIncrementedKeys() Device {
	var ret Device
	// Increment last byte of AppKey and AppSKey
	if d.ApplicationKey != "" {
		k, err := hex.DecodeString(d.ApplicationKey)
		if err != nil {
			panic(err)
		}
		k[len(k)-1]++
		ret.ApplicationKey = hex.EncodeToString(k)
	}
	if d.ApplicationSessionKey != "" {
		k, err := hex.DecodeString(d.ApplicationSessionKey)
		if err != nil {
			panic(err)
		}
		k[len(k)-1]++
		ret.ApplicationSessionKey = hex.EncodeToString(k)
	}
	if d.NetworkSessionKey != "" {
		k, err := hex.DecodeString(d.NetworkSessionKey)
		if err != nil {
			panic(err)
		}
		k[len(k)-1]++
		ret.NetworkSessionKey = hex.EncodeToString(k)
	}

	return ret
}

// Packet is a LoRaWAN uplink packet.
type Packet struct {
	FCnt int `json:"fcnt"`
}
