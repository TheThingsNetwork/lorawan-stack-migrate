package firefly

import (
	"context"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

type Source struct{}

// NewSource creates a ner Firefly source
func NewSource(ctx context.Context, flags *pflag.FlagSet) (source.Source, error) {
	return Source{}, nil
}

// ExportDevice implements the source.Source interface.
func (s Source) ExportDevice(devID string) (*ttnpb.EndDevice, error) { return nil, nil }

// RangeDevices implements the source.Source interface.
func (s Source) RangeDevices(_ string, f func(source.Source, string) error) error { return nil }

// Close implements the Source interface.
func (s Source) Close() error { return nil }
