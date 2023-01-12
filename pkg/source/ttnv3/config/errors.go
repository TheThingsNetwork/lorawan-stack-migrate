package config

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errNoAppAPIKey                    = errors.DefineInvalidArgument("no_app_api_key", "no app api key")
	errNoIdentityServerGRPCAddress    = errors.DefineInvalidArgument("no_identity_server_grpc_address", "no identity server grpc address")
	errNoJoinServerGRPCAddress        = errors.DefineInvalidArgument("no_join_server_grpc_address", "no join server grpc address")
	errNoApplicationServerGRPCAddress = errors.DefineInvalidArgument("no_application_server_grpc_address", "no application server grpc address")
	errNoNetworkServerGRPCAddress     = errors.DefineInvalidArgument("no_network_server_grpc_address", "no network server grpc address")
)
