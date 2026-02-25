//go:build !linux

package dpkg

import (
	"context"

	"github.com/gogrlx/snack"
)

func available() bool { return false }

func install(_ context.Context, _ []snack.Target, _ ...snack.Option) error {
	return snack.ErrUnsupportedPlatform
}

func remove(_ context.Context, _ []snack.Target, _ ...snack.Option) error {
	return snack.ErrUnsupportedPlatform
}

func purge(_ context.Context, _ []snack.Target, _ ...snack.Option) error {
	return snack.ErrUnsupportedPlatform
}

func list(_ context.Context) ([]snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func search(_ context.Context, _ string) ([]snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func info(_ context.Context, _ string) (*snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func isInstalled(_ context.Context, _ string) (bool, error) {
	return false, snack.ErrUnsupportedPlatform
}

func version(_ context.Context, _ string) (string, error) {
	return "", snack.ErrUnsupportedPlatform
}
