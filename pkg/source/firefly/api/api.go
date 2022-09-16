package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/apex/log"
)

var (
	logger = &log.Logger{}

	urlPrefix = "http://"
	apiURL    string
	auth      string
)

func SetLogger(l *log.Logger) {
	logger = l
}

func SetApiURL(s string) {
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	s = strings.TrimPrefix(strings.TrimPrefix(s, "http://"), "https://")
	apiURL = s
	logger.WithField("apiURL", apiURL).Debug("Set API URL")
}

func SetAuth(s string) {
	auth = s
	logger.WithField("auth", auth).Debug("Set auth")
}

func UseTLS(b bool) {
	if b {
		urlPrefix = "https://"
		return
	}
	urlPrefix = "http://"
}

func urlWithAuth(content string, fields ...string) string {
	url := urlPrefix + apiURL + strings.TrimSuffix(content, "/") + "?auth=" + auth
	for _, f := range fields {
		url += "&" + f
	}
	logger.WithField("url", url).Debug("Generate URL")
	return url
}

func GetDeviceByEUI(eui string) (*http.Response, error) {
	return http.Get(urlWithAuth("devices/eui/" + eui))
}

func GetDeviceList() (*http.Response, error) {
	return http.Get(urlWithAuth("devices"))
}

func GetDeviceListByAppID(appID string) (*http.Response, error) {
	return http.Get(urlWithAuth("application/" + appID + "/euis"))
}

func GetPacketList() (*http.Response, error) {
	return http.Get(urlWithAuth("packets"))
}

func GetLastPacket() (*http.Response, error) {
	return http.Get(urlWithAuth("packets", "limit_to_last=1"))
}

func PutDeviceUpdate(eui string, fields map[string]string) (*http.Response, error) {
	d := struct {
		Device map[string]string `json:"device"`
	}{Device: fields}
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	logger.WithField("json", b).Debug("Update fields of device")
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("devices/eui/%s", eui), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return http.DefaultClient.Do(req)
}
