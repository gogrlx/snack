// Package ports provides Go bindings for OpenBSD ports/packages.
package ports

import (
	"context"

	"github.com/gogrlx/snack"
)

// Ports wraps the OpenBSD pkg_add/pkg_delete/pkg_info CLI tools.
type Ports struct {
	snack.Locker
}

// New returns a new Ports manager.
func New() *Ports {
	return &Ports{}
}

// Name returns "ports".
func (p *Ports) Name() string { return "ports" }

// Available reports whether pkg_add is present on the system.
func (p *Ports) Available() bool { return available() }

// Install one or more packages.
func (p *Ports) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	p.Lock()
	defer p.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (p *Ports) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	p.Lock()
	defer p.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages and cleans up dependencies.
func (p *Ports) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages.
func (p *Ports) Upgrade(ctx context.Context, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return upgrade(ctx, opts...)
}

// Update is a no-op on OpenBSD (updates via fw_update or syspatch).
func (p *Ports) Update(ctx context.Context) error {
	return update(ctx)
}

// List returns all installed packages.
func (p *Ports) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries for packages matching the query.
func (p *Ports) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (p *Ports) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (p *Ports) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (p *Ports) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// NormalizeName returns the canonical form of a package name.
func (p *Ports) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
func (p *Ports) ParseArch(name string) (string, string) {
	return parseArchNormalize(name)
}

// Verify interface compliance at compile time.
var _ snack.Manager = (*Ports)(nil)
var _ snack.PackageUpgrader = (*Ports)(nil)

// UpgradePackages upgrades specific installed packages.
func (p *Ports) UpgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	p.Lock()
	defer p.Unlock()
	return upgradePackages(ctx, pkgs, opts...)
}
