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

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	urlPrefix = "https://"
	apiURL    string
	apiKey    string

	client = &http.Client{Timeout: 10 * time.Second}
)

func SetURLPrefix(insecure bool) {
	if insecure {
		urlPrefix = "http://"
		return
	}
	urlPrefix = "https://"
}

func SetAPIURL(url string) {
	if len(strings.Split(url, "://")) <= 1 {
		url = urlPrefix + url
	}
	apiURL = url
}

func SetAPIKey(key string) {
	apiKey = key
}

func NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, apiURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	return req, nil
}

type Device struct {
	ID                string   `json:"_id"`
	Devaddr           string   `json:"devaddr"`
	Seqno             int      `json:"seqno"`
	Seqdn             int      `json:"seqdn"`
	Seqq              int      `json:"seqq"`
	AdrCnt            int      `json:"adrCnt"`
	Subscription      int      `json:"subscription"`
	Txrate            int      `json:"txrate"`
	Rxrate            int      `json:"rxrate"`
	Devclass          string   `json:"devclass"`
	Rxw               int      `json:"rxw"`
	Dutycycle         int      `json:"dutycycle"`
	Adr               bool     `json:"adr"`
	AdrMin            int      `json:"adrMin"`
	AdrMax            int      `json:"adrMax"`
	AdrFix            int      `json:"adrFix"`
	AdrCntLimit       int      `json:"adrCntLimit"`
	Seqrelax          bool     `json:"seqrelax"`
	Seqdnreset        bool     `json:"seqdnreset"`
	Nonce             int      `json:"nonce"`
	LastJoin          string   `json:"lastJoin"`
	LastSeen          int      `json:"lastSeen"`
	Rssi              float64  `json:"rssi"`
	Snr               float64  `json:"snr"`
	Freq              int      `json:"freq"`
	Sf                int      `json:"sf"`
	Bw                int      `json:"bw"`
	Gw                string   `json:"gw"`
	Appeui            string   `json:"appeui"`
	LastDevStatusSeen string   `json:"lastDevStatusSeen"`
	Bat               int      `json:"bat"`
	DevSnr            int      `json:"devSnr"`
	Lorawan           *Lorawan `json:"lorawan"`
}

type Lorawan struct {
	Major    int    `json:"major"`
	Minor    int    `json:"minor"`
	Revision string `json:"revision"`
}

func GetDevice(appID, devEUI string) (*Device, error) {
	d := new(Device)

	req, err := NewRequest("GET", fmt.Sprintf("/app/%s/device/%s", appID, devEUI), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return d, json.Unmarshal(body, d)
}

type PaginatedDevices struct {
	Page    int      `json:"page"`
	PerPage int      `json:"perPage"`
	Total   int      `json:"total"`
	Devices []Device `json:"apps"`
}

func GetPaginatedDevices(appID string, page int) (*PaginatedDevices, error) {
	req, err := NewRequest("GET", fmt.Sprintf("/app/%s/devices?page=%d", appID, page), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	r := new(PaginatedDevices)
	if err := json.Unmarshal(body, r); err != nil {
		return nil, err
	}

	return r, nil
}
