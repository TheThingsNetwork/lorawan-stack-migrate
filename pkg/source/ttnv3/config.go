package ttnv3

import (
	"fmt"
	"os"

	"github.com/TheThingsNetwork/go-utils/handlers/cli"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	ttnapex "github.com/TheThingsNetwork/go-utils/log/apex"
	apex "github.com/apex/log"
	"github.com/spf13/pflag"
)

var logger *apex.Logger

type config struct {
	appAccessKey string
	appID        string

	identityServerGRPCAddress    string
	joinServerGRPCAddress        string
	applicationServerGRPCAddress string
	networkServerGRPCAddress     string
}

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.String(flagWithPrefix("app-id"), os.Getenv("TTNV3_APP_ID"), "TTS Application ID")
	flags.String(flagWithPrefix("app-access-key"), os.Getenv("TTNV3_APP_ACCESS_KEY"), "TTS Application Access Key (with 'devices' permissions)")
	flags.String(flagWithPrefix("identity-server-grpc-address"), os.Getenv("TTNV3_IDENTITY_SERVER_GRPC_ADDRESS"), "TTS Identity Server GRPC Address")
	flags.String(flagWithPrefix("join-server-grpc-address"), os.Getenv("TTNV3_JOIN_SERVER_GRPC_ADDRESS"), "TTS Join Server GRPC Address")
	flags.String(flagWithPrefix("application-server-grpc-address"), os.Getenv("TTNV3_APPLICATION_SERVER_GRPC_ADDRESS"), "TTS Application Server GRPC Address")
	flags.String(flagWithPrefix("network-server-grpc-address"), os.Getenv("TTNV3_NETWORK_SERVER_GRPC_ADDRESS"), "TTS Network Server GRPC Address")
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

	appAccessKey := stringFlag(flagWithPrefix("app-access-key"))
	if appAccessKey == "" {
		return nil, errNoAppAccessKey
	}
	appID := stringFlag(flagWithPrefix("app-id"))
	if appID == "" {
		return nil, errNoAppID
	}
	identityServerGRPCAddress := stringFlag(flagWithPrefix("identity-server-grpc-address"))
	if identityServerGRPCAddress == "" {
		return nil, errNoIdentityServerGRPCAddress
	}
	joinServerGRPCAddress := stringFlag(flagWithPrefix("join-server-grpc-address"))
	if joinServerGRPCAddress == "" {
		return nil, errNoJoinServerGRPCAddress
	}
	applicationServerGRPCAddress := stringFlag(flagWithPrefix("network-server-grpc-address"))
	if applicationServerGRPCAddress == "" {
		return nil, errNoApplicationServerGRPCAddress
	}
	networkServerGRPCAddress := stringFlag(flagWithPrefix("join-server-grpc-address"))
	if networkServerGRPCAddress == "" {
		return nil, errNoNetworkServerGRPCAddress
	}
	logLevel := ttnapex.InfoLevel
	if boolFlag("verbose") {
		logLevel = ttnapex.DebugLevel
	}
	logger = &apex.Logger{
		Level:   logLevel,
		Handler: cli.New(os.Stderr),
	}
	ttnlog.Set(ttnapex.Wrap(logger))
	return &config{
		appAccessKey: appAccessKey,
		appID:        appID,

		identityServerGRPCAddress:    identityServerGRPCAddress,
		joinServerGRPCAddress:        joinServerGRPCAddress,
		applicationServerGRPCAddress: applicationServerGRPCAddress,
		networkServerGRPCAddress:     networkServerGRPCAddress,
	}, nil
}

func flagWithPrefix(f string) string {
	return fmt.Sprintf("ttnv3.%s", f)
}
