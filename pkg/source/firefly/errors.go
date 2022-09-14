package firefly

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	// errNoAppID  = errors.DefineInvalidArgument("no_app_id", "no app id")
	errNoAPIURL  = errors.DefineInvalidArgument("no_api_url", "no api url")
	errNoAPIKEY  = errors.DefineInvalidArgument("no_api_key", "no api key")
	errNoJoinEUI = errors.DefineInvalidArgument("no_join_eui", "no join eui")
)
