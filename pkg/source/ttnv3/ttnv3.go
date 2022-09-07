package ttnv3

import (
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

func init() {
	source.RegisterSource(source.Registration{
		Name:        "ttnv3",
		Description: "Migrate from The Things Stack",
		FlagSet:     flagSet(),
		Create:      NewSource,
	})
}
