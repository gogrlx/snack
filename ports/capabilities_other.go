//go:build !openbsd

package ports

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

func autoremove(_ context.Context, _ ...snack.Option) error {
	return snack.ErrUnsupportedPlatform
}

func clean(_ context.Context) error {
	return snack.ErrUnsupportedPlatform
}

func fileList(_ context.Context, _ string) ([]string, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func owner(_ context.Context, _ string) (string, error) {
	return "", snack.ErrUnsupportedPlatform
}

func upgradePackages(_ context.Context, _ []snack.Target, _ ...snack.Option) (snack.InstallResult, error) {
	return snack.InstallResult{}, snack.ErrUnsupportedPlatform
}
