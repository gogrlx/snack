package ports

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseList parses the output of `pkg_info`.
// Format: "name-version description text"
func parseList(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// First field is name-version, rest is description
		parts := strings.SplitN(line, " ", 2)
		nameVer := parts[0]
		name, ver := splitNameVersion(nameVer)
		p := snack.Package{
			Name:      name,
			Version:   ver,
			Installed: true,
		}
		if len(parts) == 2 {
			p.Description = strings.TrimSpace(parts[1])
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseSearchResults parses the output of `pkg_info -Q <query>`.
// Each line is a package name-version.
func parseSearchResults(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, ver := splitNameVersion(line)
		pkgs = append(pkgs, snack.Package{
			Name:    name,
			Version: ver,
		})
	}
	return pkgs
}

// parseInfoOutput parses `pkg_info <pkg>` output.
// The first line is typically "Information for name-version" or
// the package description block. We extract name/version from
// the stem or the provided pkg name.
func parseInfoOutput(output string, pkg string) *snack.Package {
	lines := strings.Split(output, "\n")
	p := &snack.Package{Installed: true}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Information for ") {
			nameVer := strings.TrimPrefix(line, "Information for ")
			nameVer = strings.TrimSuffix(nameVer, ":")
			p.Name, p.Version = splitNameVersion(nameVer)
			continue
		}
	}

	// If we got description lines (after the header), join them
	if p.Name != "" {
		var desc []string
		inDesc := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "Information for ") {
				inDesc = false
				continue
			}
			if strings.HasPrefix(trimmed, "Comment:") {
				p.Description = strings.TrimSpace(strings.TrimPrefix(trimmed, "Comment:"))
				continue
			}
			if strings.HasPrefix(trimmed, "Description:") {
				inDesc = true
				continue
			}
			if inDesc && trimmed != "" {
				desc = append(desc, trimmed)
			}
		}
		if p.Description == "" && len(desc) > 0 {
			p.Description = strings.Join(desc, " ")
		}
	}

	if p.Name == "" {
		// Fallback: try to parse from pkg argument
		p.Name, p.Version = splitNameVersion(pkg)
		if p.Name == "" {
			return nil
		}
	}
	return p
}

// splitNameVersion splits "name-version" at the last hyphen.
// OpenBSD packages use the last hyphen before a version number as separator.
func splitNameVersion(s string) (string, string) {
	idx := strings.LastIndex(s, "-")
	if idx <= 0 {
		return s, ""
	}
	return s[:idx], s[idx+1:]
}
