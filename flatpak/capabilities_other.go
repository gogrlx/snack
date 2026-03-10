//go:build !linux

package flatpak

import (
	"context"

	"github.com/gogrlx/snack"
)

func latestVersion(_ context.Context, _ string) (string, error) {
	return "", snack.ErrUnsupportedPlatform
}

func listUpgrades(_ context.Context) ([]snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func upgradeAvailable(_ context.Context, _ string) (bool, error) {
	return false, snack.ErrUnsupportedPlatform
}

func versionCmp(_ context.Context, _, _ string) (int, error) {
	return 0, snack.ErrUnsupportedPlatform
}
