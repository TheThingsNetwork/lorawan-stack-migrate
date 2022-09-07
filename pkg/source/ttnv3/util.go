package ttnv3

import (
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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
	nonImplicitPaths := nonImplicitPaths(paths...)
	is = ttnpb.AllowedFields(nonImplicitPaths, getEndDeviceFromIS)
	ns = ttnpb.AllowedFields(nonImplicitPaths, getEndDeviceFromNS)
	as = ttnpb.AllowedFields(nonImplicitPaths, getEndDeviceFromAS)
	js = ttnpb.AllowedFields(nonImplicitPaths, getEndDeviceFromJS)
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
