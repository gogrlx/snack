// Package apt provides Go bindings for APT (Advanced Packaging Tool) on Debian/Ubuntu.
package apt

import (
	"context"

	"github.com/gogrlx/snack"
)

// Apt implements the snack.Manager interface using apt-get and apt-cache.
type Apt struct {
	snack.Locker
}

// New returns a new Apt manager.
func New() *Apt {
	return &Apt{}
}

// Name returns "apt".
func (a *Apt) Name() string { return "apt" }

// Install one or more packages.
func (a *Apt) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (a *Apt) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge one or more packages including config files.
func (a *Apt) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages.
func (a *Apt) Upgrade(ctx context.Context, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return upgrade(ctx, opts...)
}

// Update refreshes the package index.
func (a *Apt) Update(ctx context.Context) error {
	a.Lock()
	defer a.Unlock()
	return update(ctx)
}

// List returns all installed packages.
func (a *Apt) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the package index.
func (a *Apt) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (a *Apt) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (a *Apt) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (a *Apt) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// Available reports whether apt-get is present on the system.
func (a *Apt) Available() bool {
	return available()
}

// LatestVersion returns the latest available version of a package.
func (a *Apt) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns packages that have newer versions available.
func (a *Apt) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version of a package is available.
func (a *Apt) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings using dpkg's native comparison.
func (a *Apt) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Hold pins packages at their current version.
func (a *Apt) Hold(ctx context.Context, pkgs []string) error {
	a.Lock()
	defer a.Unlock()
	return hold(ctx, pkgs)
}

// Unhold removes version pins from packages.
func (a *Apt) Unhold(ctx context.Context, pkgs []string) error {
	a.Lock()
	defer a.Unlock()
	return unhold(ctx, pkgs)
}

// ListHeld returns all currently held packages.
func (a *Apt) ListHeld(ctx context.Context) ([]snack.Package, error) {
	return listHeld(ctx)
}

// IsHeld reports whether a specific package is currently held.
func (a *Apt) IsHeld(ctx context.Context, pkg string) (bool, error) {
	return isHeld(ctx, pkg)
}

// Autoremove removes packages no longer required as dependencies.
func (a *Apt) Autoremove(ctx context.Context, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return autoremove(ctx, opts...)
}

// Clean removes cached package files.
func (a *Apt) Clean(ctx context.Context) error {
	a.Lock()
	defer a.Unlock()
	return clean(ctx)
}

// FileList returns all files installed by a package.
func (a *Apt) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (a *Apt) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}

// ListRepos returns all configured package repositories.
func (a *Apt) ListRepos(ctx context.Context) ([]snack.Repository, error) {
	return listRepos(ctx)
}

// AddRepo adds a new package repository.
func (a *Apt) AddRepo(ctx context.Context, repo snack.Repository) error {
	a.Lock()
	defer a.Unlock()
	return addRepo(ctx, repo)
}

// RemoveRepo removes a configured repository.
func (a *Apt) RemoveRepo(ctx context.Context, id string) error {
	a.Lock()
	defer a.Unlock()
	return removeRepo(ctx, id)
}

// AddKey imports a GPG key for package verification.
func (a *Apt) AddKey(ctx context.Context, key string) error {
	a.Lock()
	defer a.Unlock()
	return addKey(ctx, key)
}

// RemoveKey removes a GPG key.
func (a *Apt) RemoveKey(ctx context.Context, keyID string) error {
	a.Lock()
	defer a.Unlock()
	return removeKey(ctx, keyID)
}

// ListKeys returns all trusted package signing keys.
func (a *Apt) ListKeys(ctx context.Context) ([]string, error) {
	return listKeys(ctx)
}

// NormalizeName returns the canonical form of a package name.
func (a *Apt) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
func (a *Apt) ParseArch(name string) (string, string) {
	return parseArch(name)
}

// Compile-time interface checks.
var (
	_ snack.Manager        = (*Apt)(nil)
	_ snack.VersionQuerier = (*Apt)(nil)
	_ snack.Holder         = (*Apt)(nil)
	_ snack.Cleaner        = (*Apt)(nil)
	_ snack.FileOwner      = (*Apt)(nil)
	_ snack.RepoManager    = (*Apt)(nil)
	_ snack.KeyManager     = (*Apt)(nil)
	_ snack.NameNormalizer = (*Apt)(nil)
)
