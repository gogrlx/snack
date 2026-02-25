package dnf

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.VersionQuerier = (*DNF)(nil)
	_ snack.Holder         = (*DNF)(nil)
	_ snack.Cleaner        = (*DNF)(nil)
	_ snack.FileOwner      = (*DNF)(nil)
	_ snack.RepoManager    = (*DNF)(nil)
	_ snack.KeyManager     = (*DNF)(nil)
	_ snack.Grouper        = (*DNF)(nil)
	_ snack.NameNormalizer = (*DNF)(nil)
)

// LatestVersion returns the latest available version from configured repositories.
func (d *DNF) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns packages that have newer versions available.
func (d *DNF) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (d *DNF) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings using RPM version comparison.
func (d *DNF) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Hold pins packages at their current version.
func (d *DNF) Hold(ctx context.Context, pkgs []string) error {
	d.Lock()
	defer d.Unlock()
	return hold(ctx, pkgs)
}

// Unhold removes version pins.
func (d *DNF) Unhold(ctx context.Context, pkgs []string) error {
	d.Lock()
	defer d.Unlock()
	return unhold(ctx, pkgs)
}

// ListHeld returns all currently held packages.
func (d *DNF) ListHeld(ctx context.Context) ([]snack.Package, error) {
	return listHeld(ctx)
}

// Autoremove removes orphaned packages.
func (d *DNF) Autoremove(ctx context.Context, opts ...snack.Option) error {
	d.Lock()
	defer d.Unlock()
	return autoremove(ctx, opts...)
}

// Clean removes cached package files.
func (d *DNF) Clean(ctx context.Context) error {
	d.Lock()
	defer d.Unlock()
	return clean(ctx)
}

// FileList returns all files installed by a package.
func (d *DNF) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (d *DNF) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}

// ListRepos returns all configured package repositories.
func (d *DNF) ListRepos(ctx context.Context) ([]snack.Repository, error) {
	return listRepos(ctx)
}

// AddRepo adds a new package repository.
func (d *DNF) AddRepo(ctx context.Context, repo snack.Repository) error {
	d.Lock()
	defer d.Unlock()
	return addRepo(ctx, repo)
}

// RemoveRepo removes a configured repository.
func (d *DNF) RemoveRepo(ctx context.Context, id string) error {
	d.Lock()
	defer d.Unlock()
	return removeRepo(ctx, id)
}

// AddKey imports a GPG key for package verification.
func (d *DNF) AddKey(ctx context.Context, key string) error {
	d.Lock()
	defer d.Unlock()
	return addKey(ctx, key)
}

// RemoveKey removes a GPG key.
func (d *DNF) RemoveKey(ctx context.Context, keyID string) error {
	d.Lock()
	defer d.Unlock()
	return removeKey(ctx, keyID)
}

// ListKeys returns all trusted package signing keys.
func (d *DNF) ListKeys(ctx context.Context) ([]string, error) {
	return listKeys(ctx)
}

// GroupList returns all available package groups.
func (d *DNF) GroupList(ctx context.Context) ([]string, error) {
	return groupList(ctx)
}

// GroupInfo returns the packages in a group.
func (d *DNF) GroupInfo(ctx context.Context, group string) ([]snack.Package, error) {
	return groupInfo(ctx, group)
}

// GroupInstall installs all packages in a group.
func (d *DNF) GroupInstall(ctx context.Context, group string, opts ...snack.Option) error {
	d.Lock()
	defer d.Unlock()
	return groupInstall(ctx, group, opts...)
}

// NormalizeName returns the canonical form of a package name.
func (d *DNF) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
func (d *DNF) ParseArch(name string) (string, string) {
	return parseArch(name)
}
