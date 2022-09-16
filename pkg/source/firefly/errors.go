package firefly

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errNoAPIKEY  = errors.DefineInvalidArgument("no_api_key", "no api key")
	errNoAPIURL  = errors.DefineInvalidArgument("no_api_url", "no api url")
	errNoAppID   = errors.DefineInvalidArgument("no_app_id", "no app id")
	errNoJoinEUI = errors.DefineInvalidArgument("no_join_eui", "no join eui")
)
