package firefly

import (
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

func init() {
	cfg := New()

	source.RegisterSource(source.Registration{
		Name:        "firefly",
		Description: "Migrate from Digimondo's Firefly",
		FlagSet:     cfg.Flags(),
		Create:      createNewSource(cfg),
	})
}
