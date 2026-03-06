package brew

import "strings"

// normalizeName returns the canonical form of a package name.
// Homebrew formulae can have version suffixes like `python@3.12`.
// This strips the version suffix to get the base formula name.
func normalizeName(name string) string {
	n, _ := parseVersionSuffix(name)
	return n
}

// parseVersionSuffix extracts the version suffix from a formula name.
// Homebrew uses @ to denote versioned formulae (e.g., "python@3.12").
// Returns the name without version and the version string.
func parseVersionSuffix(name string) (string, string) {
	if idx := strings.LastIndex(name, "@"); idx > 0 {
		return name[:idx], name[idx+1:]
	}
	return name, ""
}
