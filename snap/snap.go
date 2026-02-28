// Package snap provides Go bindings for the snap package manager.
package snap

import (
	"context"

	"github.com/gogrlx/snack"
)

// Snap wraps the snap CLI.
type Snap struct {
	snack.Locker
}

// New returns a new Snap manager.
func New() *Snap {
	return &Snap{}
}

// Name returns "snap".
func (s *Snap) Name() string { return "snap" }

// Available reports whether snap is present on the system.
func (s *Snap) Available() bool { return available() }

// Install one or more packages.
func (s *Snap) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	s.Lock()
	defer s.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (s *Snap) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	s.Lock()
	defer s.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages including all data.
func (s *Snap) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	s.Lock()
	defer s.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed snaps.
func (s *Snap) Upgrade(ctx context.Context, opts ...snack.Option) error {
	s.Lock()
	defer s.Unlock()
	return upgrade(ctx, opts...)
}

// Update checks for available updates (snap auto-refreshes).
func (s *Snap) Update(ctx context.Context) error {
	return update(ctx)
}

// List returns all installed snaps.
func (s *Snap) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the snap store.
func (s *Snap) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific snap.
func (s *Snap) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a snap is currently installed.
func (s *Snap) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a snap.
func (s *Snap) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// Verify interface compliance at compile time.
var _ snack.Manager = (*Snap)(nil)
var _ snack.PackageUpgrader = (*Snap)(nil)

// UpgradePackages upgrades specific installed packages.
func (s *Snap) UpgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	s.Lock()
	defer s.Unlock()
	return upgradePackages(ctx, pkgs, opts...)
}
