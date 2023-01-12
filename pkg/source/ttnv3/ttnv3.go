package ttnv3

import (
	"go.uber.org/zap"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack-migrate/pkg/source/ttnv3/config"
)

var logger *zap.SugaredLogger

func init() {
	cfg, flags := config.New()

	logger, _ = config.NewLogger(cfg.Verbose)

	source.RegisterSource(source.Registration{
		Name:        "ttnv3",
		Description: "Migrate from The Things Stack",
		FlagSet:     flags,
		Create:      createNewSource(cfg),
	})
}
