// Package dpkg provides Go bindings for dpkg (low-level Debian package tool).
package dpkg

import (
	"context"

	"github.com/gogrlx/snack"
)

// Dpkg implements the snack.Manager interface using dpkg and dpkg-query.
type Dpkg struct {
	snack.Locker
}

// New returns a new Dpkg manager.
func New() *Dpkg {
	return &Dpkg{}
}

// Name returns "dpkg".
func (d *Dpkg) Name() string { return "dpkg" }

// Install one or more .deb files.
func (d *Dpkg) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	d.Lock()
	defer d.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (d *Dpkg) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	d.Lock()
	defer d.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge one or more packages including config files.
func (d *Dpkg) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	d.Lock()
	defer d.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade is not supported by dpkg.
func (d *Dpkg) Upgrade(_ context.Context, _ ...snack.Option) error {
	return snack.ErrUnsupportedPlatform
}

// Update is not supported by dpkg.
func (d *Dpkg) Update(_ context.Context) error {
	return snack.ErrUnsupportedPlatform
}

// List returns all installed packages.
func (d *Dpkg) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries installed packages matching the pattern.
func (d *Dpkg) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (d *Dpkg) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (d *Dpkg) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (d *Dpkg) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// Available reports whether dpkg is present on the system.
func (d *Dpkg) Available() bool {
	return available()
}

// FileList returns all files installed by a package.
func (d *Dpkg) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (d *Dpkg) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}

// NormalizeName returns the canonical form of a package name.
func (d *Dpkg) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
func (d *Dpkg) ParseArch(name string) (string, string) {
	return parseArch(name)
}

// Compile-time interface checks.
var (
	_ snack.Manager        = (*Dpkg)(nil)
	_ snack.FileOwner      = (*Dpkg)(nil)
	_ snack.NameNormalizer = (*Dpkg)(nil)
)
