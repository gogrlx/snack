package flatpak

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.VersionQuerier = (*Flatpak)(nil)
	_ snack.Cleaner        = (*Flatpak)(nil)
	_ snack.RepoManager    = (*Flatpak)(nil)
	_ snack.NameNormalizer = (*Flatpak)(nil)
)

// LatestVersion returns the latest available version of a flatpak.
func (f *Flatpak) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns flatpaks that have newer versions available.
func (f *Flatpak) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (f *Flatpak) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings.
func (f *Flatpak) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove removes unused runtimes and extensions.
func (f *Flatpak) Autoremove(ctx context.Context, opts ...snack.Option) error {
	f.Lock()
	defer f.Unlock()
	return autoremove(ctx, opts...)
}

// Clean is a no-op for flatpak (no direct cache clean equivalent).
func (f *Flatpak) Clean(ctx context.Context) error {
	return nil
}

// ListRepos returns all configured remotes.
func (f *Flatpak) ListRepos(ctx context.Context) ([]snack.Repository, error) {
	return listRepos(ctx)
}

// AddRepo adds a new remote.
func (f *Flatpak) AddRepo(ctx context.Context, repo snack.Repository) error {
	f.Lock()
	defer f.Unlock()
	return addRepo(ctx, repo)
}

// RemoveRepo removes a configured remote.
func (f *Flatpak) RemoveRepo(ctx context.Context, id string) error {
	f.Lock()
	defer f.Unlock()
	return removeRepo(ctx, id)
}
