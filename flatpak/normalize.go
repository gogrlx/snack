package flatpak

import "strings"

// normalizeName returns the canonical form of a flatpak application ID.
// Flatpak references can include branch/arch suffixes like:
//   - org.gnome.Calculator/x86_64/stable
//   - org.gnome.Calculator//stable (default arch)
//
// This strips branch and arch to return just the app ID.
func normalizeName(name string) string {
	n, _ := parseRef(name)
	return n
}

// parseRef extracts the architecture from a flatpak reference if present.
// Flatpak references can be in the form:
//   - app-id
//   - app-id/arch/branch
//   - app-id//branch (default arch)
//
// Returns the app-id and architecture (or empty string).
func parseRef(name string) (string, string) {
	parts := strings.SplitN(name, "/", 3)
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return name, ""
}
