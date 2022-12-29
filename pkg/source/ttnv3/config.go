package ttnv3

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"go.thethings.network/lorawan-stack-migrate/pkg/source/ttnv3/api"
)

var logger *zap.SugaredLogger

type serverConfig struct {
	applicationServerGRPCAddress,
	identityServerGRPCAddress,
	joinServerGRPCAddress,
	networkServerGRPCAddress string
}

func (c *serverConfig) SetEmptyFields(address string) {
	if a := &c.applicationServerGRPCAddress; *a == "" {
		a = &address
	}
	if a := &c.identityServerGRPCAddress; *a == "" {
		a = &address
	}
	if a := &c.joinServerGRPCAddress; *a == "" {
		a = &address
	}
	if a := &c.networkServerGRPCAddress; *a == "" {
		a = &address
	}
}

func (c *serverConfig) AnyFieldEmpty() error {
	if c.applicationServerGRPCAddress == "" {
		return errNoApplicationServerGRPCAddress.New()
	}
	if c.identityServerGRPCAddress == "" {
		return errNoIdentityServerGRPCAddress.New()
	}
	if c.joinServerGRPCAddress == "" {
		return errNoJoinServerGRPCAddress.New()
	}
	if c.networkServerGRPCAddress == "" {
		return errNoNetworkServerGRPCAddress.New()
	}
	return nil
}

type Config struct {
	appID string

	serverConfig serverConfig

	deleteSourceDevice,
	dryRun,
	noSession bool
}

var config Config

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	flags.StringVar(&config.appID,
		"app-id",
		os.Getenv("TTNV3_APP_ID"),
		"TTS Application ID")
	flags.String(
		"app-api-key",
		os.Getenv("TTNV3_APP_API_KEY"),
		"TTS Application Access Key (with 'devices' permissions)")

	flags.String(
		"ca-file",
		os.Getenv("TTNV3_CA_FILE"),
		"TTS Path to a CA file (optional)")
	flags.Bool(
		"insecure",
		false,
		"TTS allow TCP connection")

	flags.String(
		"default-grpc-address",
		os.Getenv("TTNV3_DEFAULT_GRPC_ADDRESS"),
		"TTS default GRPC Address (optional)")
	flags.StringVar(&config.serverConfig.applicationServerGRPCAddress,
		"appplication-server-grpc-address",
		os.Getenv("TTNV3_APPLICATION_SERVER_GRPC_ADDRESS"),
		"TTS Application Server GRPC Address")
	flags.StringVar(&config.serverConfig.identityServerGRPCAddress,
		"identity-server-grpc-address",
		os.Getenv("TTNV3_IDENTITY_SERVER_GRPC_ADDRESS"),
		"TTS Identity Server GRPC Address")
	flags.StringVar(&config.serverConfig.joinServerGRPCAddress,
		"join-server-grpc-address",
		os.Getenv("TTNV3_JOIN_SERVER_GRPC_ADDRESS"),
		"TTS Join Server GRPC Address")
	flags.StringVar(&config.serverConfig.networkServerGRPCAddress,
		"network-server-grpc-address",
		os.Getenv("TTNV3_NETWORK_SERVER_GRPC_ADDRESS"),
		"TTS Network Server GRPC Address")

	flags.BoolVar(&config.noSession,
		"no-session",
		false,
		"TTS export devices without session")
	flags.BoolVar(&config.deleteSourceDevice,
		"delete-source-device",
		false,
		"TTS delete exported devices")

	return flags
}

func getConfig(flags *pflag.FlagSet) (*Config, error) {
	cfg := zap.NewProductionConfig()
	verbose, err := flags.GetBool("verbose")
	if err != nil {
		return nil, err
	}
	if verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	logger = zapLogger.Sugar()

	switch apiKey, err := flags.GetString("app-api-key"); {
	case err != nil:
		return nil, err

	case apiKey == "":
		return nil, errNoAppAPIKey.New()

	default:
		api.SetAuth("bearer", apiKey)
	}

	switch insecure, err := flags.GetBool("insecure"); {
	case err != nil:
		return nil, err

	case insecure:
		api.SetInsecure(true)
		logger.Warn("Using insecure connection to API")

	default:
		caPath, err := flags.GetString("ca-path")
		if err != nil {
			return nil, err
		}
		if caPath != "" {
			setCustomCA(caPath)
		}
	}

	switch defaultGRPCAddress, err := flags.GetString("default-grpc-address"); {
	case err != nil:
		return nil, err

	case defaultGRPCAddress != "":
		config.serverConfig.SetEmptyFields(defaultGRPCAddress)

	default:
		if err := config.serverConfig.AnyFieldEmpty(); err != nil {
			return nil, err
		}
	}

	dryRun, err := flags.GetBool("dry-run")
	if err != nil {
		return nil, err
	}
	config.dryRun = dryRun

	// deleteSourceDevice not allowed when dryRun
	if config.dryRun && config.deleteSourceDevice {
		logger.Warn("Cannot delete source device during a dry run.")
		config.deleteSourceDevice = false
	}

	return &config, nil
}

func setCustomCA(path string) error {
	pemBytes, err := os.ReadFile(path)
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
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: rootCAs}
	if err = api.AddCA(pemBytes); err != nil {
		return err
	}
	return nil
}
