package ttnv3

import (
	"context"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

// Source implements the Source interface.
type Source struct {
	ctx context.Context

	config *config
}

// NewSource creates a new TTNv3 cource
func NewSource(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) {
	config, err := getConfig(flags)
	if err != nil {
		return Source{}, err
	}
	s := Source{
		ctx:    ctx,
		config: config,
	}
	return s, nil
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) { return nil, nil }

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(appID string, f func(source.Source, string) error) error { return nil }

// Close implements the Source interface.
func (s Source) Close() error { return nil }
