package ttnv3

import (
	"github.com/spf13/pflag"
)

type config struct{}

func flagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	return flags
}

func getConfig(flags *pflag.FlagSet) (*config, error) {
	return &config{}, nil
}
