package devices

import (
	"encoding/json"
	"io"

	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/api"
)

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

type JSONDevice struct {
	Device Device
}

func deviceFromRequestBody(r io.ReadCloser) (*Device, error) {
	defer r.Close()
	var device JSONDevice
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&device); err != nil {
		return nil, err
	}
	return &device.Device, nil
}

func GetDevice(eui string) (*Device, error) {
	resp, err := api.GetDeviceByEUI(eui)
	if err != nil {
		return nil, err
	}
	return deviceFromRequestBody(resp.Body)
}

type JSONDevices struct {
	Devices []Device
}

func devicesListFromRequestBody(r io.ReadCloser) ([]Device, error) {
	defer r.Close()
	var devices JSONDevices
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&devices); err != nil {
		return nil, err
	}
	return devices.Devices, nil
}

func GetAllDevices() ([]Device, error) {
	resp, err := api.GetDeviceList()
	if err != nil {
		return nil, err
	}
	return devicesListFromRequestBody(resp.Body)
}

func GetDeviceListByAppID(appID string) ([]Device, error) {
	resp, err := api.GetDeviceListByAppID(appID)
	if err != nil {
		return nil, err
	}
	return devicesListFromRequestBody(resp.Body)
}

type Location struct {
	Lattitude float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}
