package awsiot

import (
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/awsiot/config"
)

func init() {
	cfg := config.New()

	source.RegisterSource(source.Registration{
		Name:        "awsiot",
		Description: "Migrate from AWS IoT",
		FlagSet:     cfg.Flags(),
		Create:      createNewSource(cfg),
	})
}
