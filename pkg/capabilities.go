package pkg

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.VersionQuerier = (*Pkg)(nil)
	_ snack.Cleaner        = (*Pkg)(nil)
	_ snack.FileOwner      = (*Pkg)(nil)
)

// LatestVersion returns the latest available version from configured repositories.
func (p *Pkg) LatestVersion(ctx context.Context, pkg string) (string, error) {
	return latestVersion(ctx, pkg)
}

// ListUpgrades returns packages that have newer versions available.
func (p *Pkg) ListUpgrades(ctx context.Context) ([]snack.Package, error) {
	return listUpgrades(ctx)
}

// UpgradeAvailable reports whether a newer version is available.
func (p *Pkg) UpgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	return upgradeAvailable(ctx, pkg)
}

// VersionCmp compares two version strings using pkg's version comparison.
func (p *Pkg) VersionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	return versionCmp(ctx, ver1, ver2)
}

// Autoremove removes packages no longer required as dependencies.
func (p *Pkg) Autoremove(ctx context.Context, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return autoremove(ctx, opts...)
}

// Clean removes cached package files.
func (p *Pkg) Clean(ctx context.Context) error {
	p.Lock()
	defer p.Unlock()
	return clean(ctx)
}

// FileList returns all files installed by a package.
func (p *Pkg) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (p *Pkg) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}
