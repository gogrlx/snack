// Package rpm provides Go bindings for the rpm low-level package manager.
package rpm

import (
	"context"

	"github.com/gogrlx/snack"
)

// RPM wraps the rpm package manager CLI.
type RPM struct {
	snack.Locker
}

// New returns a new RPM manager.
func New() *RPM {
	return &RPM{}
}

// Name returns "rpm".
func (r *RPM) Name() string { return "rpm" }

// Available reports whether rpm is present on the system.
func (r *RPM) Available() bool { return available() }

// Install one or more packages.
func (r *RPM) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	r.Lock()
	defer r.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (r *RPM) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	r.Lock()
	defer r.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages (same as Remove for rpm).
func (r *RPM) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	r.Lock()
	defer r.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Upgrade upgrades packages from files.
func (r *RPM) Upgrade(ctx context.Context, opts ...snack.Option) error {
	r.Lock()
	defer r.Unlock()
	return upgradeAll(ctx, opts...)
}

// Update is not supported by rpm (use dnf instead).
func (r *RPM) Update(_ context.Context) error {
	return snack.ErrUnsupportedPlatform
}

// List returns all installed packages.
func (r *RPM) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries installed packages matching the query.
func (r *RPM) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (r *RPM) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (r *RPM) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (r *RPM) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// Verify interface compliance at compile time.
var _ snack.Manager = (*RPM)(nil)
