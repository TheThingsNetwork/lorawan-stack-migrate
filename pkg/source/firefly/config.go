package firefly

import (
	"fmt"
	"os"

	"github.com/TheThingsNetwork/go-utils/handlers/cli"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	ttnapex "github.com/TheThingsNetwork/go-utils/log/apex"
	apex "github.com/apex/log"
	"github.com/spf13/pflag"

	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/api"
)

type config struct {
	// appID  string
	apiURL string
	apiKey string

	joinEUI string

	frequencyPlanID string
	macVersion      string
}

var logger *apex.Logger

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	// flags.String(flagWithPrefix("app-id"), os.Getenv("FIREFLY_APP_ID"), "Firefly app ID")
	flags.String(flagWithPrefix("api-url"), os.Getenv("FIREFLY_API_URL"), "Firefly API URL")
	flags.String(flagWithPrefix("api-key"), os.Getenv("FIREFLY_API_KEY"), "Firefly API key")
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
	// appID := stringFlag(flagWithPrefix("app-id"))
	// if appID == "" {
	// 	return nil, errNoAppID
	// }
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
	return &config{
		// appID:  appID,
		apiURL: apiURL,
		apiKey: apiKey,

		joinEUI: joinEUI,
	}, nil
}

func flagWithPrefix(flag string) string {
	return fmt.Sprint("firefly.", flag)
}
