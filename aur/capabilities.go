package aur

import (
	"context"

	"github.com/gogrlx/snack"
)

// LatestVersion returns the latest version of an AUR package.
func (a *AUR) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns AUR packages that have newer versions available.
func (a *AUR) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (a *AUR) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings using vercmp.
func (a *AUR) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove removes orphaned packages via pacman.
func (a *AUR) Autoremove(ctx context.Context, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return autoremove(ctx, opts...)
}

// Clean is a no-op for AUR (builds use temp directories).
func (a *AUR) Clean(ctx context.Context) error {
	return clean(ctx)
}

// UpgradePackages rebuilds and reinstalls specific AUR packages.
func (a *AUR) UpgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	a.Lock()
	defer a.Unlock()
	return upgradePackages(ctx, pkgs, opts...)
}
