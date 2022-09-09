package firefly

import (
	"context"

	"github.com/spf13/pflag"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

func init() {
	source.RegisterSource(source.Registration{
		Name:        "firefly",
		Description: "Migrate from Digimondo's Firefly",
		FlagSet:     &pflag.FlagSet{},
		Create:      func(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) { return nil, nil },
	})
}
