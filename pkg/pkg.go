// Package pkg provides Go bindings for pkg(8) (FreeBSD package manager).
package pkg

import (
	"context"

	"github.com/gogrlx/snack"
)

// Pkg wraps the FreeBSD pkg(8) package manager CLI.
type Pkg struct {
	snack.Locker
}

// New returns a new Pkg manager.
func New() *Pkg {
	return &Pkg{}
}

// Name returns "pkg".
func (p *Pkg) Name() string { return "pkg" }

// Available reports whether pkg is present on the system.
func (p *Pkg) Available() bool { return available() }

// Install one or more packages.
func (p *Pkg) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (p *Pkg) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages with force (removing dependent packages).
func (p *Pkg) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages to their latest versions.
func (p *Pkg) Upgrade(ctx context.Context, opts ...snack.Option) error {
	p.Lock()
	defer p.Unlock()
	return upgrade(ctx, opts...)
}

// Update refreshes the package database.
func (p *Pkg) Update(ctx context.Context) error {
	p.Lock()
	defer p.Unlock()
	return update(ctx)
}

// List returns all installed packages.
func (p *Pkg) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the repositories for packages matching the query.
func (p *Pkg) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (p *Pkg) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (p *Pkg) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (p *Pkg) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// Verify interface compliance at compile time.
var _ snack.Manager = (*Pkg)(nil)
