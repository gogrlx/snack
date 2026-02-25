package apk

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseListInstalled parses output from `apk list --installed`.
// Each line looks like: "name-1.2.3-r0 x86_64 {origin} (license) [installed]"
func parseListInstalled(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pkg := parseListLine(line)
		if pkg.Name != "" {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

// parseListLine parses a single line from `apk list`.
// Format: "name-1.2.3-r0 x86_64 {origin} (license) [installed]"
func parseListLine(line string) snack.Package {
	var pkg snack.Package
	pkg.Installed = strings.Contains(line, "[installed]")

	fields := strings.Fields(line)
	if len(fields) < 1 {
		return pkg
	}

	// First field is name-version
	nameVer := fields[0]
	pkg.Name, pkg.Version = splitNameVersion(nameVer)

	if len(fields) >= 2 {
		pkg.Arch = fields[1]
	}

	return pkg
}

// splitNameVersion splits "name-1.2.3-r0" into ("name", "1.2.3-r0").
// apk versions start with a digit, so we find the last hyphen before a digit.
func splitNameVersion(s string) (string, string) {
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == '-' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

// parseSearch parses output from `apk search` or `apk search -v`.
func parseSearch(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// `apk search -v` output: "name-version - description"
		if idx := strings.Index(line, " - "); idx != -1 {
			nameVer := line[:idx]
			desc := line[idx+3:]
			name, ver := splitNameVersion(nameVer)
			pkgs = append(pkgs, snack.Package{
				Name:        name,
				Version:     ver,
				Description: desc,
			})
		} else {
			// plain `apk search` just returns name-version
			name, ver := splitNameVersion(line)
			pkgs = append(pkgs, snack.Package{
				Name:    name,
				Version: ver,
			})
		}
	}
	return pkgs
}

// parseInfo parses output from `apk info -a <pkg>`.
func parseInfo(output string) *snack.Package {
	pkg := &snack.Package{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description:") {
			pkg.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}

	// First line is typically "pkgname-version description"
	// But `apk info -a` starts with "pkgname-version installed size:"
	// Let's parse key-value style
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if k, v, ok := strings.Cut(line, ":"); ok {
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			switch strings.ToLower(k) {
			case "description":
				pkg.Description = v
			case "arch":
				pkg.Arch = v
			case "url", "webpage":
				// skip
			}
		}
	}

	return pkg
}

// parseInfoNameVersion extracts name and version from `apk info <pkg>` output.
// The first line is typically "pkgname-version description".
func parseInfoNameVersion(output string) (string, string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "", ""
	}
	// first line: name-version
	first := strings.Fields(lines[0])[0]
	return splitNameVersion(first)
}
