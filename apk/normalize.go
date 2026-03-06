package apk

import "strings"

// normalizeName returns the canonical form of a package name.
// Alpine package names sometimes include version suffixes in queries.
// This strips common version patterns.
func normalizeName(name string) string {
	n, _ := parseArchNormalize(name)
	return n
}

// parseArchNormalize extracts the architecture from a package name if present.
// Alpine package names typically don't embed architecture in the package name,
// but some query outputs may include it. Common patterns:
//   - package-x86_64
//   - package-aarch64
func parseArchNormalize(name string) (string, string) {
	knownArchs := map[string]bool{
		"x86_64":   true,
		"x86":      true,
		"aarch64":  true,
		"armhf":    true,
		"armv7":    true,
		"ppc64le":  true,
		"s390x":    true,
		"riscv64":  true,
		"loongarch64": true,
	}

	if idx := strings.LastIndex(name, "-"); idx >= 0 {
		suffix := name[idx+1:]
		if knownArchs[suffix] {
			return name[:idx], suffix
		}
	}
	return name, ""
}
