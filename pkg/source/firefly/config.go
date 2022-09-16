package firefly

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/TheThingsNetwork/go-utils/handlers/cli"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	ttnapex "github.com/TheThingsNetwork/go-utils/log/apex"
	apex "github.com/apex/log"
	"github.com/spf13/pflag"

	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/api"
)

type config struct {
	apiKey string
	apiURL string

	appID           string
	frequencyPlanID string
	joinEUI         string
	macVersion      string

	dryRun      bool
	withSession bool
}

var logger *apex.Logger

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.String(flagWithPrefix("app-id"), os.Getenv("FIREFLY_APP_ID"), "Firefly app ID")
	flags.String(flagWithPrefix("api-url"), os.Getenv("FIREFLY_API_URL"), "Firefly API URL")
	flags.String(flagWithPrefix("api-key"), os.Getenv("FIREFLY_API_KEY"), "Firefly API key")
	flags.String(flagWithPrefix("ca-file"), os.Getenv("FIREFLY_CA_FILE"), "Firefly CA file for TLS (optional)")
	flags.String(flagWithPrefix("join-eui"), os.Getenv("JOIN_EUI"), "JoinEUI of exported devices")
	return flags
}

func getConfig(flags *pflag.FlagSet) (*config, error) {
	stringFlag := func(n string) string {
		f, _ := flags.GetString(n)
		return f
	}
	boolFlag := func(n string) bool {
		f, _ := flags.GetBool(n)
		return f
	}

	logLevel := ttnapex.InfoLevel
	if boolFlag("verbose") {
		logLevel = ttnapex.DebugLevel
	}
	logger = &apex.Logger{
		Level:   logLevel,
		Handler: cli.New(os.Stderr),
	}
	api.SetLogger(logger)
	ttnlog.Set(ttnapex.Wrap(logger))
	appID := stringFlag(flagWithPrefix("app-id"))
	if appID == "" {
		return nil, errNoAppID
	}
	apiURL := stringFlag(flagWithPrefix("api-url"))
	if apiURL == "" {
		return nil, errNoAPIURL
	}
	api.SetApiURL(apiURL)
	apiKey := stringFlag(flagWithPrefix("api-key"))
	if apiKey == "" {
		return nil, errNoAPIKEY
	}
	api.SetAuth(apiKey)
	joinEUI := stringFlag(flagWithPrefix("join-eui"))
	if joinEUI == "" {
		return nil, errNoJoinEUI
	}
	if caPath := stringFlag(flagWithPrefix("ca-file")); caPath != "" {
		pemBytes, err := os.ReadFile(caPath)
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
		http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs = rootCAs
	}
	return &config{
		apiKey: apiKey,
		apiURL: apiURL,

		appID:   appID,
		joinEUI: joinEUI,

		dryRun: boolFlag("dry-run"),
	}, nil
}

func flagWithPrefix(flag string) string {
	return fmt.Sprint("firefly.", flag)
}
