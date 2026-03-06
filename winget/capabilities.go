package winget

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.VersionQuerier = (*Winget)(nil)
	_ snack.RepoManager    = (*Winget)(nil)
	_ snack.NameNormalizer = (*Winget)(nil)
)

// LatestVersion returns the latest available version of a package.
func (w *Winget) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns packages that have newer versions available.
func (w *Winget) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (w *Winget) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings.
func (w *Winget) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// ListRepos returns configured winget sources.
func (w *Winget) ListRepos(ctx context.Context) ([]snack.Repository, error) {
	return sourceList(ctx)
}

// AddRepo adds a new winget source.
func (w *Winget) AddRepo(ctx context.Context, repo snack.Repository) error {
	return sourceAdd(ctx, repo)
}

// RemoveRepo removes a configured winget source.
func (w *Winget) RemoveRepo(ctx context.Context, id string) error {
	return sourceRemove(ctx, id)
}
