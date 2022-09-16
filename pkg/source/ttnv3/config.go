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

	identityServerGRPCAddress    string
	joinServerGRPCAddress        string
	applicationServerGRPCAddress string
	networkServerGRPCAddress     string

	dryRun bool
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

	appID := stringFlag(flagWithPrefix("app-id"))
	if appID == "" {
		return nil, errNoAppID
	}

	apiKey := stringFlag(flagWithPrefix("app-api-key"))
	if apiKey == "" {
		return nil, errNoAppAPIKey
	}
	api.SetAuth("bearer", apiKey)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{}
	if insecure := boolFlag(flagWithPrefix("insecure")); insecure {
		api.SetInsecure(true)
		logger.Warn("Using insecure connection to API")
	} else {
		caPath := stringFlag(flagWithPrefix("ca-file"))
		if caPath == "" {
			return nil, errNoCA
		}
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
		return nil, errNoIdentityServerGRPCAddress
	}
	joinServerGRPCAddress := stringFlag(flagWithPrefix("join-server-grpc-address"))
	if joinServerGRPCAddress == "" {
		return nil, errNoJoinServerGRPCAddress
	}
	applicationServerGRPCAddress := stringFlag(flagWithPrefix("application-server-grpc-address"))
	if applicationServerGRPCAddress == "" {
		return nil, errNoApplicationServerGRPCAddress
	}
	networkServerGRPCAddress := stringFlag(flagWithPrefix("network-server-grpc-address"))
	if networkServerGRPCAddress == "" {
		return nil, errNoNetworkServerGRPCAddress
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
	return &config{
		appID: appID,

		identityServerGRPCAddress:    identityServerGRPCAddress,
		joinServerGRPCAddress:        joinServerGRPCAddress,
		applicationServerGRPCAddress: applicationServerGRPCAddress,
		networkServerGRPCAddress:     networkServerGRPCAddress,

		dryRun: boolFlag("dry-run"),
	}, nil
}

func flagWithPrefix(f string) string {
	return fmt.Sprintf("ttnv3.%s", f)
}
