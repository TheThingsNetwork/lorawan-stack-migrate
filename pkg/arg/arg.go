package arg

import (
	"os"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"golang.org/x/exp/slices"
)

func init() {
	if len(os.Args) <= 1 {
		return
	}
	if s := os.Args[1]; slices.Contains(source.Names(), s) {
		source.ActiveSource = s
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}
}
