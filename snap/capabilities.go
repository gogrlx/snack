package snap

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.VersionQuerier = (*Snap)(nil)
	_ snack.Cleaner        = (*Snap)(nil)
	_ snack.NameNormalizer = (*Snap)(nil)
)

// LatestVersion returns the latest stable version of a snap.
func (s *Snap) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns snaps that have newer versions available.
func (s *Snap) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (s *Snap) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings.
func (s *Snap) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove is a no-op for snap (snap doesn't have orphan packages).
func (s *Snap) Autoremove(_ context.Context, _ ...snack.Option) error {
	return nil
}

// Clean removes old snap revisions to free up space.
func (s *Snap) Clean(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()
	return clean(ctx)
}
