package ttnv3

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errRead = errors.DefinePermissionDenied("read", "failed to read `{file}`")

	errNoAppID                        = errors.DefineInvalidArgument("no_app_id", "no app id")
	errNoAppAPIKey                    = errors.DefineInvalidArgument("no_app_api_key", "no app api key")
	errNoCA                           = errors.DefineInvalidArgument("no_ca", "no CA")
	errNoIdentityServerGRPCAddress    = errors.DefineInvalidArgument("no_identity_server_grpc_address", "no identity server grpc address")
	errNoJoinServerGRPCAddress        = errors.DefineInvalidArgument("no_join_server_grpc_address", "no join server grpc address")
	errNoApplicationServerGRPCAddress = errors.DefineInvalidArgument("no_application_server_grpc_address", "no application server grpc address")
	errNoNetworkServerGRPCAddress     = errors.DefineInvalidArgument("no_network_server_grpc_address", "no network server grpc address")
)
