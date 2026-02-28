//go:build !linux

package dnf

import (
	"context"

	"github.com/gogrlx/snack"
)

func available() bool { return false }

func (d *DNF) detectVersion() {}

func install(_ context.Context, _ bool, _ []snack.Target, _ ...snack.Option) (snack.InstallResult, error) {
	return snack.InstallResult{}, snack.ErrUnsupportedPlatform
}

func remove(_ context.Context, _ bool, _ []snack.Target, _ ...snack.Option) (snack.RemoveResult, error) {
	return snack.RemoveResult{}, snack.ErrUnsupportedPlatform
}

func upgrade(_ context.Context, _ ...snack.Option) error {
	return snack.ErrUnsupportedPlatform
}

func update(_ context.Context) error {
	return snack.ErrUnsupportedPlatform
}

func list(_ context.Context, _ bool) ([]snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func search(_ context.Context, _ string, _ bool) ([]snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func info(_ context.Context, _ string, _ bool) (*snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func isInstalled(_ context.Context, _ string, _ bool) (bool, error) {
	return false, snack.ErrUnsupportedPlatform
}

func version(_ context.Context, _ string, _ bool) (string, error) {
	return "", snack.ErrUnsupportedPlatform
}

func upgradePackages(_ context.Context, _ bool, _ []snack.Target, _ ...snack.Option) (snack.InstallResult, error) {
	return snack.InstallResult{}, snack.ErrUnsupportedPlatform
}
