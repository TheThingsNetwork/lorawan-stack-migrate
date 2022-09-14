package firefly

import (
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

func init() {
	source.RegisterSource(source.Registration{
		Name:        "firefly",
		Description: "Migrate from Digimondo's Firefly",
		FlagSet:     flagSet(),
		Create:      NewSource,
	})
}
