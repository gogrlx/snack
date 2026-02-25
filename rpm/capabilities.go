package rpm

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.FileOwner      = (*RPM)(nil)
	_ snack.NameNormalizer = (*RPM)(nil)
)

// FileList returns all files installed by a package.
func (r *RPM) FileList(ctx context.Context, pkg string) ([]string, error) {
	return fileList(ctx, pkg)
}

// Owner returns the package that owns a given file path.
func (r *RPM) Owner(ctx context.Context, path string) (string, error) {
	return owner(ctx, path)
}

// NormalizeName returns the canonical form of a package name.
func (r *RPM) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
func (r *RPM) ParseArch(name string) (string, string) {
	return parseArchSuffix(name)
}
