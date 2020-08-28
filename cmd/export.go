package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack-migrate/pkg/source"
)

func exportCommand(cmd *cobra.Command, args []string, f func(s source.Source, item string) error) error {
	var iter Iterator
	switch len(args) {
	case 0:
		iter = NewReaderIterator(os.Stdin, byte('\n'))
	default:
		iter = NewListIterator(args)
	}

	s, err := source.NewSource(ctx, cmd.Flags())
	if err != nil {
		return err
	}

	for {
		item, err := iter.Next()
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}
		if item == "" {
			continue
		}

		if err := f(s, item); err != nil {
			return err
		}
	}
}
