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

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.uber.org/zap"
)

const defaultTimeout = 10 * time.Second

// Config is the Firefly client configuration.
type Config struct {
	APIKey     string
	Host       string
	CACertPath string
	UseHTTP    bool
}

// Client is a Firefly client.
type Client struct {
	*Config
	*http.Client
	logger *zap.SugaredLogger
}

// New creates a new Firefly client.
func (cfg *Config) NewClient(logger *zap.SugaredLogger) (*Client, error) {
	httpTransport := &http.Transport{}
	if cfg.CACertPath != "" && !cfg.UseHTTP {
		pemBytes, err := os.ReadFile(cfg.CACertPath)
		if err != nil {
			return nil, err
		}
		rootCAs := http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs
		if rootCAs == nil {
			if rootCAs, err = x509.SystemCertPool(); err != nil {
				rootCAs = x509.NewCertPool()
			}
		}
		rootCAs.AppendCertsFromPEM(pemBytes)
		httpTransport.TLSClientConfig = &tls.Config{
			RootCAs: rootCAs,
		}
	}
	return &Client{
		Config: cfg,
		Client: &http.Client{
			Transport: httpTransport,
			Timeout:   defaultTimeout,
		},
		logger: logger,
	}, nil
}

var (
	errResourceNotFound     = errors.DefineNotFound("resource_not_found", "resource `{resource}` not found")
	errServer               = errors.Define("server", "server error with code `{code}`")
	errUnexpectedStatusCode = errors.Define("unexpected_status_code", "unexpected status code `{code}`")
)

// do executes an HTTP request.
func (c *Client) do(resource, method string, body []byte, params string) ([]byte, error) {
	scheme := "https"
	if c.UseHTTP {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s/api/v1/%s?auth=%s", scheme, c.Host, resource, c.APIKey)
	if params != "" {
		url += "&" + params
	}
	logger := c.logger.With("url", url)
	logger.Debug("Request resource")
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusOK:
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	case res.StatusCode == http.StatusNotFound:
		return nil, errResourceNotFound.WithAttributes("resource", resource)
	case res.StatusCode >= 500:
		return nil, errServer.WithAttributes("code", res.StatusCode)
	default:
		return nil, errUnexpectedStatusCode.WithAttributes("code", res.StatusCode)
	}
}

// GetDeviceByEUI gets a device by the EUI.
func (c *Client) GetDeviceByEUI(eui string) (*Device, error) {
	body, err := c.do(fmt.Sprintf("devices/eui/%s", eui), http.MethodGet, nil, "")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Device Device `json:"device"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}
	return &wrapper.Device, nil
}

// UpdateDevice updates the device.
func (c *Client) UpdateDeviceByEUI(eui string, dev Device) error {
	wrapper := struct {
		Device Device `json:"device"`
	}{
		Device: dev,
	}
	body, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}
	_, err = c.do(fmt.Sprintf("devices/eui/%s", eui), http.MethodPut, body, "")
	return err
}

// GetLastPacket gets the last packet for a device.
func (c *Client) GetLastPacket(eui string) (*Packet, error) {
	body, err := c.do(fmt.Sprintf("devices/eui/%s/packets", eui), http.MethodGet, nil, "limit_to_last=1")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Packets []Packet `json:"packets"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}
	if len(wrapper.Packets) == 0 {
		return &Packet{}, nil
	}

	if len(wrapper.Packets) > 1 {
		c.logger.Warn("More than one packet found for device. Returning the last one.")
	}
	return &wrapper.Packets[0], nil
}

// GetAllDevices gets all devices that the API key has access to.
func (c *Client) GetAllDevices() ([]Device, error) {
	body, err := c.do("devices", http.MethodGet, nil, "")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Device []Device `json:"devices"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Device, nil
}
