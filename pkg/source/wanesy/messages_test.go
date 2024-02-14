// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package wanesy_test

import (
	"net/http"
	"testing"

	"github.com/smarty/assertions"
	"github.com/smarty/assertions/should"
	. "go.thethings.network/lorawan-stack-migrate/pkg/source/wanesy"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func TestImportAndExport(t *testing.T) {
	a := assertions.New(t)
	devices, err := ImportDevices("./testdata/test.csv")
	a.So(err, should.BeNil)
	a.So(len(devices), should.Equal, 1)
	var devEUI types.EUI64
	if err = devEUI.UnmarshalText([]byte("1111111111111111")); err != nil {
		t.Fatalf("Failed to unmarshal EUI: %v", err)
	}
	dev, ok := devices[devEUI]
	if !ok {
		t.FailNow()
	}
	a.So(dev.AppEui, should.Equal, "1111111111111111")
	a.So(dev.FCntUp, should.Equal, "20")
	a.So(dev.FCntDown, should.Equal, "10")
	a.So(dev.DevAddr, should.Equal, "01234567")

	fpFetcher, err := fetch.FromHTTP(
		http.DefaultClient,
		"https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans/master",
	)
	if err != nil {
		t.Fatalf("Failed to create fetcher: %v", err)
	}

	v3Device, err := dev.EndDevice(frequencyplans.NewStore(fpFetcher), "test-app", "EU_863_870")
	if err != nil {
		t.Fatalf("Failed to convert device: %v", err)
	}
	a.So(v3Device, should.NotBeNil)

	// Check converted fields.
	a.So(v3Device.Ids.DevEui, should.Resemble, devEUI.Bytes())
	a.So(v3Device.Session, should.NotBeNil)
	a.So(v3Device.Session.DevAddr, should.Resemble, []byte{0x01, 0x23, 0x45, 0x67})
	a.So(v3Device.Session.LastFCntUp, should.Equal, 20)
	a.So(v3Device.Session.LastAFCntDown, should.Equal, 10)
	a.So(v3Device.LorawanPhyVersion, should.Equal, ttnpb.PHYVersion_PHY_V1_0_2_REV_A)
	a.So(v3Device.LorawanVersion, should.Equal, ttnpb.MACVersion_MAC_V1_0_2)
	a.So(v3Device.Session.Keys, should.NotBeNil)
}
