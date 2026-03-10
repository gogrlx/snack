package aur

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.Manager         = (*AUR)(nil)
	_ snack.VersionQuerier  = (*AUR)(nil)
	_ snack.Cleaner         = (*AUR)(nil)
	_ snack.PackageUpgrader = (*AUR)(nil)
)

// LatestVersion returns the latest version available in the AUR.
func (a *AUR) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns installed foreign packages that have newer versions in the AUR.
func (a *AUR) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available in the AUR.
func (a *AUR) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings using pacman's vercmp.
func (a *AUR) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove removes orphaned packages via pacman.
func (a *AUR) Autoremove(ctx context.Context, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return autoremove(ctx, opts...)
}

// Clean removes cached build artifacts from the build directory.
func (a *AUR) Clean(_ context.Context) error {
	return a.cleanBuildDir()
}

// UpgradePackages rebuilds and reinstalls specific AUR packages.
func (a *AUR) UpgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	a.Lock()
	defer a.Unlock()
	return a.install(ctx, pkgs, opts...)
}
