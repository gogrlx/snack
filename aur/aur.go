// Package aur provides a native Go client for the Arch User Repository.
//
// Unlike other snack backends that wrap CLI tools, aur uses the AUR RPC API
// directly for queries and git+makepkg for building. Packages are built in
// a temporary directory and installed via pacman -U.
//
// Requirements: git, makepkg, pacman (all present on any Arch Linux system).
package aur

import (
	"context"

	"github.com/gogrlx/snack"
)

// AUR wraps the Arch User Repository using its RPC API and makepkg.
type AUR struct {
	snack.Locker

	// BuildDir is the base directory for cloning and building packages.
	// If empty, a temporary directory is created per build.
	BuildDir string

	// MakepkgFlags are extra flags passed to makepkg (e.g. "--skippgpcheck").
	MakepkgFlags []string
}

// New returns a new AUR manager with default settings.
func New() *AUR {
	return &AUR{}
}

// Option configures an AUR manager.
type AUROption func(*AUR)

// WithBuildDir sets a persistent build directory.
func WithBuildDir(dir string) AUROption {
	return func(a *AUR) { a.BuildDir = dir }
}

// WithMakepkgFlags sets extra flags for makepkg.
func WithMakepkgFlags(flags ...string) AUROption {
	return func(a *AUR) { a.MakepkgFlags = flags }
}

// NewWithOptions returns a new AUR manager with the given options.
func NewWithOptions(opts ...AUROption) *AUR {
	a := New()
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Name returns "aur".
func (a *AUR) Name() string { return "aur" }

// Available reports whether the AUR toolchain (git, makepkg, pacman) is present.
func (a *AUR) Available() bool { return available() }

// Install clones, builds, and installs AUR packages.
func (a *AUR) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	a.Lock()
	defer a.Unlock()
	return a.install(ctx, pkgs, opts...)
}

// Remove removes packages via pacman (AUR packages are regular pacman packages once installed).
func (a *AUR) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	a.Lock()
	defer a.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages including config files via pacman.
func (a *AUR) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade rebuilds and reinstalls all foreign (AUR) packages.
func (a *AUR) Upgrade(ctx context.Context, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return a.upgradeAll(ctx, opts...)
}

// Update is a no-op for AUR (there is no local package index to refresh).
func (a *AUR) Update(_ context.Context) error {
	return nil
}

// List returns all installed foreign (non-repo) packages, which are typically AUR packages.
func (a *AUR) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the AUR RPC API for packages matching the query.
func (a *AUR) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return rpcSearch(ctx, query)
}

// Info returns details about a specific AUR package from the RPC API.
func (a *AUR) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return rpcInfo(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (a *AUR) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (a *AUR) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}
