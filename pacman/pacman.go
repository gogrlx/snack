// Package pacman provides Go bindings for the pacman package manager (Arch Linux).
package pacman

import (
	"context"

	"github.com/gogrlx/snack"
)

// Pacman wraps the pacman package manager CLI.
type Pacman struct {
	snack.Locker
}

// New returns a new Pacman manager.
func New() *Pacman {
	return &Pacman{}
}

// Name returns "pacman".
func (p *Pacman) Name() string { return "pacman" }

// Available reports whether pacman is present on the system.
func (p *Pacman) Available() bool { return available() }

// Install one or more packages.
func (p *Pacman) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	p.Lock()
	defer p.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (p *Pacman) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	p.Lock()
	defer p.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages including configuration files.
func (p *Pacman) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages to their latest versions.
func (p *Pacman) Upgrade(ctx context.Context, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return upgrade(ctx, opts...)
}

// Update refreshes the package database.
func (p *Pacman) Update(ctx context.Context) error {
	p.Lock()
	defer p.Unlock()
	return update(ctx)
}

// List returns all installed packages.
func (p *Pacman) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the repositories for packages matching the query.
func (p *Pacman) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (p *Pacman) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (p *Pacman) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (p *Pacman) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// NormalizeName returns the canonical form of a package name.
func (p *Pacman) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
func (p *Pacman) ParseArch(name string) (string, string) {
	return parseArchNormalize(name)
}

// Verify interface compliance at compile time.
var _ snack.Manager = (*Pacman)(nil)
var _ snack.PackageUpgrader = (*Pacman)(nil)

// UpgradePackages upgrades specific installed packages.
func (p *Pacman) UpgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	p.Lock()
	defer p.Unlock()
	return upgradePackages(ctx, pkgs, opts...)
}
