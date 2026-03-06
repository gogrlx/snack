package winget

import "strings"

// normalizeName returns the canonical form of a winget package ID.
// Winget IDs use dot-separated Publisher.Package format (e.g.
// "Microsoft.VisualStudioCode"). This trims whitespace but otherwise
// preserves the ID as-is since winget IDs are case-insensitive but
// conventionally PascalCase.
func normalizeName(name string) string {
	return strings.TrimSpace(name)
}

// parseArch extracts the architecture from a winget package name if present.
// Winget IDs do not include architecture suffixes, so this returns the
// name unchanged with an empty string.
func parseArch(name string) (string, string) {
	return name, ""
}
