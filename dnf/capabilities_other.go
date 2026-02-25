//go:build !linux

package dnf

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

func hold(_ context.Context, _ []string) error {
	return snack.ErrUnsupportedPlatform
}

func unhold(_ context.Context, _ []string) error {
	return snack.ErrUnsupportedPlatform
}

func listHeld(_ context.Context) ([]snack.Package, error) {
	return nil, snack.ErrUnsupportedPlatform
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

func listRepos(_ context.Context) ([]snack.Repository, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func addRepo(_ context.Context, _ snack.Repository) error {
	return snack.ErrUnsupportedPlatform
}

func removeRepo(_ context.Context, _ string) error {
	return snack.ErrUnsupportedPlatform
}

func addKey(_ context.Context, _ string) error {
	return snack.ErrUnsupportedPlatform
}

func removeKey(_ context.Context, _ string) error {
	return snack.ErrUnsupportedPlatform
}

func listKeys(_ context.Context) ([]string, error) {
	return nil, snack.ErrUnsupportedPlatform
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
