package aur

// normalizeName returns the canonical form of an AUR package name.
// AUR package names are simple identifiers without architecture or version
// suffixes, so this is essentially a pass-through.
func normalizeName(name string) string {
	return name
}

// parseArch extracts the architecture from a package name if present.
// AUR package names do not include architecture suffixes,
// so this returns the name unchanged with an empty string.
func parseArch(name string) (string, string) {
	return name, ""
}
