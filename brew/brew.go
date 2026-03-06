// Package brew provides Go bindings for Homebrew package manager.
package brew

import (
	"context"

	"github.com/gogrlx/snack"
)

// Brew wraps the brew CLI.
type Brew struct {
	snack.Locker
}

// New returns a new Brew manager.
func New() *Brew {
	return &Brew{}
}

// Compile-time interface checks.
var (
	_ snack.Manager         = (*Brew)(nil)
	_ snack.VersionQuerier  = (*Brew)(nil)
	_ snack.Cleaner         = (*Brew)(nil)
	_ snack.FileOwner       = (*Brew)(nil)
	_ snack.NameNormalizer  = (*Brew)(nil)
	_ snack.PackageUpgrader = (*Brew)(nil)
)

// Name returns "brew".
func (b *Brew) Name() string { return "brew" }

// Available reports whether brew is present on the system.
func (b *Brew) Available() bool { return available() }

// Install one or more packages.
func (b *Brew) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	b.Lock()
	defer b.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (b *Brew) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	b.Lock()
	defer b.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages including all dependencies.
func (b *Brew) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	b.Lock()
	defer b.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages.
func (b *Brew) Upgrade(ctx context.Context, opts ...snack.Option) error {
	b.Lock()
	defer b.Unlock()
	return upgrade(ctx, opts...)
}

// Update refreshes the package index.
func (b *Brew) Update(ctx context.Context) error {
	b.Lock()
	defer b.Unlock()
	return update(ctx)
}

// List returns all installed packages.
func (b *Brew) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries Homebrew for packages.
func (b *Brew) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (b *Brew) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (b *Brew) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (b *Brew) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// NormalizeName returns the canonical form of a package name.
func (b *Brew) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
// Homebrew does not embed architecture in package names.
func (b *Brew) ParseArch(name string) (string, string) {
	return name, ""
}

// LatestVersion returns the latest available version of a package.
func (b *Brew) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns packages that have newer versions available.
func (b *Brew) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (b *Brew) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings.
func (b *Brew) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove removes packages that were installed as dependencies but are no longer needed.
func (b *Brew) Autoremove(ctx context.Context, opts ...snack.Option) error {
	b.Lock()
	defer b.Unlock()
	return autoremove(ctx, opts...)
}

// Clean removes cached package files.
func (b *Brew) Clean(ctx context.Context) error {
	b.Lock()
	defer b.Unlock()
	return clean(ctx)
}

// FileList returns all files installed by a package.
func (b *Brew) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (b *Brew) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}

// UpgradePackages upgrades specific installed packages.
func (b *Brew) UpgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	b.Lock()
	defer b.Unlock()
	return upgradePackages(ctx, pkgs, opts...)
}
