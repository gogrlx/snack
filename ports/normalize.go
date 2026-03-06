package ports

// normalizeName returns the canonical form of a package name.
// OpenBSD packages use "name-version" format. This strips the version
// portion if present, returning just the name.
func normalizeName(name string) string {
	n, _ := splitNameVersion(name)
	return n
}

// parseArchNormalize extracts the architecture from a package name if present.
// OpenBSD package names do not embed architecture in the name itself
// (the arch is separate), so this returns the name unchanged.
func parseArchNormalize(name string) (string, string) {
	return name, ""
}
