package pkg

// normalizeName returns the canonical form of a package name.
// FreeBSD pkg package names use "name-version" format. This function
// strips the version portion if present, returning just the name.
func normalizeName(name string) string {
	n, _ := splitNameVersion(name)
	return n
}

// parseArchNormalize extracts the architecture from a package name if present.
// FreeBSD pkg package names do not embed architecture in the name itself
// (the arch is separate metadata), so this returns the name unchanged with
// an empty architecture string.
func parseArchNormalize(name string) (string, string) {
	return name, ""
}
