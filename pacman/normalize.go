package pacman

// normalizeName returns the canonical form of a package name.
// Pacman package names do not include architecture suffixes, so this
// is essentially a pass-through. The package name is returned as-is.
func normalizeName(name string) string {
	return name
}

// parseArch extracts the architecture from a package name if present.
// Pacman package names do not include architecture suffixes in the name itself
// (the arch is separate metadata), so this returns the name unchanged with an
// empty architecture string.
func parseArchNormalize(name string) (string, string) {
	return name, ""
}
