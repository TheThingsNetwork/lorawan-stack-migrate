package api

import (
	"net/http"
	"strings"

	"github.com/apex/log"
)

var (
	logger *log.Logger

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

func urlWithAuth(content string) string {
	url := urlPrefix + apiURL + strings.TrimSuffix(content, "/") + "?auth=" + auth
	logger.WithField("url", url).Debug("Generate URL")
	return url
}

func RequestDeviceByEUI(eui string) (*http.Response, error) {
	return http.Get(urlWithAuth("devices/eui/" + eui))
}

func RequestDevicesList() (*http.Response, error) {
	return http.Get(urlWithAuth("devices"))
}

func RequestDevicesListByAppID(appID string) (*http.Response, error) {
	return http.Get(urlWithAuth("application/" + appID + "/euis"))
}
