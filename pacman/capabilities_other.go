//go:build !linux

package pacman

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

func groupList(_ context.Context) ([]string, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func groupInfo(_ context.Context, _ string) ([]snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func groupInstall(_ context.Context, _ string, _ ...snack.Option) error {
	return snack.ErrUnsupportedPlatform
}

func groupIsInstalled(_ context.Context, _ string) (bool, error) {
	return false, snack.ErrUnsupportedPlatform
}
