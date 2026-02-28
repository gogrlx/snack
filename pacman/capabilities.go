package pacman

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.VersionQuerier = (*Pacman)(nil)
	_ snack.Cleaner        = (*Pacman)(nil)
	_ snack.FileOwner      = (*Pacman)(nil)
	_ snack.Grouper        = (*Pacman)(nil)
	_ snack.GroupQuerier   = (*Pacman)(nil)
)

// NOTE: snack.Holder is not implemented for pacman. While pacman supports
// IgnorePkg in /etc/pacman.conf, there is no clean CLI command to hold/unhold
// packages. pacman-contrib provides some tooling but it's not standard.

// LatestVersion returns the latest available version from configured repositories.
func (p *Pacman) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns packages that have newer versions available.
func (p *Pacman) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (p *Pacman) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings using pacman's vercmp.
func (p *Pacman) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove removes orphaned packages.
func (p *Pacman) Autoremove(ctx context.Context, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return autoremove(ctx, opts...)
}

// Clean removes cached package files.
func (p *Pacman) Clean(ctx context.Context) error {
	p.Lock()
	defer p.Unlock()
	return clean(ctx)
}

// FileList returns all files installed by a package.
func (p *Pacman) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (p *Pacman) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}

// GroupList returns all available package groups.
func (p *Pacman) GroupList(ctx context.Context) ([]string, error) {
	return groupList(ctx)
}

// GroupInfo returns the packages in a group.
func (p *Pacman) GroupInfo(ctx context.Context, group string) ([]snack.Package, error) {
	return groupInfo(ctx, group)
}

// GroupInstall installs all packages in a group.
func (p *Pacman) GroupInstall(ctx context.Context, group string, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return groupInstall(ctx, group, opts...)
}

// GroupIsInstalled reports whether all packages in the group are installed.
func (p *Pacman) GroupIsInstalled(ctx context.Context, group string) (bool, error) {
	return groupIsInstalled(ctx, group)
}
