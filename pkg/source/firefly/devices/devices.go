package devices

import (
	"encoding/json"
	"io"

	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/api"
)

type Device struct {
	Address               string `json:"address"`
	ApplicationKey        string `json:"application_key"`
	ApplicationSessionKey string `json:"application_session_key"`
	ClassC                bool   `json:"class_c"`
	Deduplicate           bool   `json:"deduplicate"`
	Description           string `json:"description"`
	DeviceClassID         int    `json:"device_class_id"`
	EUI                   string `json:"eui"`
	FrameCounter          int    `json:"frame_counter"`
	InsertedAt            string `json:"inserted_at"`
	Name                  string `json:"name"`
	NetworkSessionKey     string `json:"network_session_key"`
	OTAA                  bool   `json:"otaa"`
	OrganizationID        int    `json:"organization_id"`
	Region                string `json:"region"`
	Rx2DataRate           int    `json:"rx2_data_rate"`
	UpdatedAt             string `json:"updated_at"`
}

func (d Device) DeviceClass() (*DeviceClass, error) {
	return nil, nil
}

type JSONDevice struct {
	Device Device
}

func deviceFromRequestBody(readCloser io.ReadCloser) (*Device, error) {
	var device JSONDevice
	decoder := json.NewDecoder(readCloser)
	if err := decoder.Decode(&device); err != nil {
		return nil, err
	}
	return &device.Device, nil
}

func GetDevice(eui string) (*Device, error) {
	req, err := api.RequestDeviceByEUI(eui)
	if err != nil {
		return nil, err
	}
	return deviceFromRequestBody(req.Body)
}

func devicesListFromRequestBody(readCloser io.ReadCloser) ([]*Device, error) {
	var devices JSONDevices
	decoder := json.NewDecoder(readCloser)
	if err := decoder.Decode(&devices); err != nil {
		return nil, err
	}
	return devices.Devices, nil
}

func GetAllDevices() ([]*Device, error) {
	req, err := api.RequestDevicesList()
	if err != nil {
		return nil, err
	}
	return devicesListFromRequestBody(req.Body)
}

func GetDeviceListByAppID(appID string) ([]*Device, error) {
	req, err := api.RequestDevicesListByAppID(appID)
	if err != nil {
		return nil, err
	}
	return devicesListFromRequestBody(req.Body)
}

type JSONDevices struct {
	Devices []*Device
}
