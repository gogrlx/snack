package apk

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.VersionQuerier = (*Apk)(nil)
	_ snack.Cleaner        = (*Apk)(nil)
	_ snack.FileOwner      = (*Apk)(nil)
	_ snack.DryRunner      = (*Apk)(nil)
)

// SupportsDryRun reports that apk honors [snack.WithDryRun] via --simulate.
func (a *Apk) SupportsDryRun() bool { return true }

// LatestVersion returns the latest available version from configured repositories.
func (a *Apk) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns packages that have newer versions available.
func (a *Apk) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (a *Apk) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings using apk's native comparison.
func (a *Apk) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove is not directly supported by apk. apk does not track
// "auto-installed" dependencies in a way that allows safe autoremoval.
// This is a no-op that returns nil.
func (a *Apk) Autoremove(ctx context.Context, opts ...snack.Option) error {
	return autoremove(ctx, opts...)
}

// Clean removes cached package files.
func (a *Apk) Clean(ctx context.Context) error {
	a.Lock()
	defer a.Unlock()
	return clean(ctx)
}

// FileList returns all files installed by a package.
func (a *Apk) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (a *Apk) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}
