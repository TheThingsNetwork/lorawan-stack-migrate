package firefly

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/api"
)

type Config struct {
	source.Config

	apiKey string
	apiURL string
	caPath string

	appID           string
	frequencyPlanID string
	joinEUI         string
	macVersion      string

	verbose     bool
	dryRun      bool
	withSession bool

	flags *pflag.FlagSet
}

// New
func New() *Config {
	config := &Config{
		flags: &pflag.FlagSet{},
	}
	config.flags.StringVar(&config.appID,
		"app-id",
		os.Getenv("FIREFLY_APP_ID"),
		"Firefly application ID")
	config.flags.StringVar(&config.apiURL,
		"api-url",
		os.Getenv("FIREFLY_API_URL"),
		"URL of the Firefly API")
	config.flags.StringVar(&config.apiKey,
		"api-key",
		os.Getenv("FIREFLY_API_KEY"),
		"Key to access the Firefly API")
	config.flags.StringVar(&config.joinEUI,
		"join-eui",
		os.Getenv("JOIN_EUI"),
		"JoinEUI for the exported devices")
	config.flags.StringVar(&config.joinEUI,
		"frequency-plan-id",
		os.Getenv("FREQUENCY_PLAN_ID"),
		"Frequency Plan ID for the exported devices")
	config.flags.StringVar(&config.joinEUI,
		"mac-version",
		os.Getenv("MAC_VERSION"),
		"MAC version for the exported devices")

	config.verbose, _ = config.flags.GetBool("verbose")
	config.dryRun, _ = config.flags.GetBool("dry-run")
	config.withSession, _ = config.flags.GetBool("with-session")

	return config
}

var logger *zap.SugaredLogger

// Initialize the configuration.
func (c *Config) Initialize() error {
	cfg := zap.NewProductionConfig()
	if c.verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapLogger, err := cfg.Build()
	if err != nil {
		return err
	}
	logger = zapLogger.Sugar()
	api.SetLogger(logger)

	if c.appID == "" {
		return errNoAppID.New()
	}
	if c.apiURL == "" {
		return errNoAPIURL.New()
	}
	api.SetApiURL(c.apiURL)
	if c.apiKey == "" {
		return errNoAPIKEY.New()
	}
	api.SetAuth(c.apiKey)
	if c.joinEUI == "" {
		return errNoJoinEUI.New()
	}
	if c.frequencyPlanID == "" {
		return errNoFrequencyPlanID.New()
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{}
	if c.caPath != "" {
		pemBytes, err := os.ReadFile(c.caPath)
		if err != nil {
			return err
		}
		rootCAs := http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs
		if rootCAs == nil {
			if rootCAs, err = x509.SystemCertPool(); err != nil {
				rootCAs = x509.NewCertPool()
			}
		}
		rootCAs.AppendCertsFromPEM(pemBytes)
		http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs = rootCAs
		api.UseTLS(true)
	}

	return nil
}

// Flags returns the flags for the configuration.
func (c *Config) Flags() *pflag.FlagSet {
	return c.flags
}
