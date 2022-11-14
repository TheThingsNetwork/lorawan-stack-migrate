package ttnv3

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"go.thethings.network/lorawan-stack-migrate/pkg/source/ttnv3/api"
)

var logger *zap.SugaredLogger

type config struct {
	appID string

	identityServerGRPCAddress,
	joinServerGRPCAddress,
	applicationServerGRPCAddress,
	networkServerGRPCAddress string

	deleteSourceDevice,
	dryRun,
	noSession bool
}

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.String(flagWithPrefix("app-id"), os.Getenv("TTNV3_APP_ID"), "TTS Application ID")
	flags.String(flagWithPrefix("app-api-key"), os.Getenv("TTNV3_APP_API_KEY"), "TTS Application Access Key (with 'devices' permissions)")
	flags.String(flagWithPrefix("ca-file"), os.Getenv("TTNV3_CA_FILE"), "TTS Path to a CA file (optional)")
	flags.String(flagWithPrefix("identity-server-grpc-address"), os.Getenv("TTNV3_IDENTITY_SERVER_GRPC_ADDRESS"), "TTS Identity Server GRPC Address")
	flags.String(flagWithPrefix("join-server-grpc-address"), os.Getenv("TTNV3_JOIN_SERVER_GRPC_ADDRESS"), "TTS Join Server GRPC Address")
	flags.String(flagWithPrefix("application-server-grpc-address"), os.Getenv("TTNV3_APPLICATION_SERVER_GRPC_ADDRESS"), "TTS Application Server GRPC Address")
	flags.String(flagWithPrefix("network-server-grpc-address"), os.Getenv("TTNV3_NETWORK_SERVER_GRPC_ADDRESS"), "TTS Network Server GRPC Address")
	flags.Bool(flagWithPrefix("insecure"), false, "TTS allow TCP connection")
	flags.Bool(flagWithPrefix("no-session"), false, "TTS export devices without session")
	flags.Bool(flagWithPrefix("delete-source-device"), false, "TTS delete exported devices")
	return flags
}

func getConfig(flags *pflag.FlagSet) (*config, error) {
	stringFlag := func(f string) string {
		s, _ := flags.GetString(f)
		return s
	}
	boolFlag := func(f string) bool {
		b, _ := flags.GetBool(f)
		return b
	}

	apiKey := stringFlag(flagWithPrefix("app-api-key"))
	if apiKey == "" {
		return nil, errNoAppAPIKey.New()
	}
	api.SetAuth("bearer", apiKey)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{}
	if insecure := boolFlag(flagWithPrefix("insecure")); insecure {
		api.SetInsecure(true)
		logger.Warn("Using insecure connection to API")
	} else if caPath := stringFlag(flagWithPrefix("ca-file")); caPath != "" {
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
		if err = api.AddCA(pemBytes); err != nil {
			return nil, err
		}
	}

	identityServerGRPCAddress := stringFlag(flagWithPrefix("identity-server-grpc-address"))
	if identityServerGRPCAddress == "" {
		return nil, errNoIdentityServerGRPCAddress.New()
	}
	joinServerGRPCAddress := stringFlag(flagWithPrefix("join-server-grpc-address"))
	if joinServerGRPCAddress == "" {
		return nil, errNoJoinServerGRPCAddress.New()
	}
	applicationServerGRPCAddress := stringFlag(flagWithPrefix("application-server-grpc-address"))
	if applicationServerGRPCAddress == "" {
		return nil, errNoApplicationServerGRPCAddress.New()
	}
	networkServerGRPCAddress := stringFlag(flagWithPrefix("network-server-grpc-address"))
	if networkServerGRPCAddress == "" {
		return nil, errNoNetworkServerGRPCAddress.New()
	}
	cfg := zap.NewProductionConfig()
	if boolFlag("verbose") {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	logger = zapLogger.Sugar()

	dryRun := boolFlag("dry-run")
	deleteSourceDevice := boolFlag(flagWithPrefix("delete-source-device"))
	// deleteSourceDevice not allowed when dryRun
	if dryRun && deleteSourceDevice {
		logger.Warn("Cannot delete source device during a dry run.")
		deleteSourceDevice = false
	}
	noSession := boolFlag(flagWithPrefix("no-session"))

	return &config{
		appID: stringFlag(flagWithPrefix("app-id")),

		identityServerGRPCAddress:    identityServerGRPCAddress,
		joinServerGRPCAddress:        joinServerGRPCAddress,
		applicationServerGRPCAddress: applicationServerGRPCAddress,
		networkServerGRPCAddress:     networkServerGRPCAddress,

		deleteSourceDevice: deleteSourceDevice,
		dryRun:             dryRun,
		noSession:          noSession,
	}, nil
}

func flagWithPrefix(f string) string {
	return fmt.Sprintf("ttnv3.%s", f)
}
