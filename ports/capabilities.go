package ports

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.VersionQuerier = (*Ports)(nil)
	_ snack.Cleaner        = (*Ports)(nil)
	_ snack.FileOwner      = (*Ports)(nil)
)

// LatestVersion returns the latest available version from configured repositories.
func (p *Ports) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns packages that have newer versions available.
func (p *Ports) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (p *Ports) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings.
// OpenBSD has no native version comparison tool, so this uses a simple
// lexicographic comparison of the version strings.
func (p *Ports) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove removes packages no longer required as dependencies.
func (p *Ports) Autoremove(ctx context.Context, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return autoremove(ctx, opts...)
}

// Clean removes cached package files.
func (p *Ports) Clean(ctx context.Context) error {
	p.Lock()
	defer p.Unlock()
	return clean(ctx)
}

// FileList returns all files installed by a package.
func (p *Ports) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (p *Ports) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}
