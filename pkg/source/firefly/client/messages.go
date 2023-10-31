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
	Address               string    `json:"address"`
	AdrLimit              int       `json:"adr_limit"`
	ApplicationKey        string    `json:"application_key"`
	ApplicationSessionKey string    `json:"application_session_key"`
	ClassC                bool      `json:"class_c"`
	Deduplicate           bool      `json:"deduplicate"`
	Description           string    `json:"description"`
	DeviceClassID         int       `json:"device_class_id"`
	EUI                   string    `json:"eui"`
	FrameCounter          int       `json:"frame_counter"`
	InsertedAt            string    `json:"inserted_at"`
	Location              *Location `json:"location"`
	Name                  string    `json:"name"`
	NetworkSessionKey     string    `json:"network_session_key"`
	OTAA                  bool      `json:"otaa"`
	OrganizationID        int       `json:"organization_id"`
	OverrideLocation      bool      `json:"override_location"`
	Region                string    `json:"region"`
	Rx2DataRate           int       `json:"rx2_data_rate"`
	SkipFCntCheck         bool      `json:"skip_fcnt_check"`
	Tags                  []string  `json:"tags"`
	UpdatedAt             string    `json:"updated_at"`
}

// WithIncrementKeys returns the device with last byte of AppKey and AppSKey incremented by one.
func (d Device) WithIncrementKeys() Device {
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
	return ret
}

// Packet is a LoRaWAN packet.
type Packet struct {
	FCnt int `json:"fcnt"`
}
