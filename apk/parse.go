package apk

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseListInstalled parses output from `apk list --installed`.
// Each line looks like: "name-1.2.3-r0 x86_64 {origin} (license) [installed]"
func parseListInstalled(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(output, "\n") {
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
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// `apk search -v` output: "name-version - description"
		if before, after, ok := strings.Cut(line, " - "); ok {
			nameVer := before
			desc := after
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
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "description:"); ok {
			pkg.Description = strings.TrimSpace(after)
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
	if len(lines) == 0 || lines[0] == "" {
		return "", ""
	}
	fields := strings.Fields(lines[0])
	if len(fields) == 0 {
		return "", ""
	}
	return splitNameVersion(fields[0])
}

// parseUpgradeSimulation parses `apk upgrade --simulate` output.
// Lines look like: "(1/3) Upgrading pkg (oldver -> newver)"
func parseUpgradeSimulation(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "Upgrading") {
			continue
		}
		// "(1/3) Upgrading pkg (oldver -> newver)"
		_, after, ok := strings.Cut(line, "Upgrading ")
		if !ok {
			continue
		}
		rest := after
		// "pkg (oldver -> newver)"
		parts := strings.SplitN(rest, " (", 2)
		if len(parts) < 1 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		var ver string
		if len(parts) == 2 {
			// "oldver -> newver)"
			verPart := strings.TrimSuffix(parts[1], ")")
			arrow := strings.Split(verPart, " -> ")
			if len(arrow) == 2 {
				ver = strings.TrimSpace(arrow[1])
			}
		}
		pkgs = append(pkgs, snack.Package{
			Name:      name,
			Version:   ver,
			Installed: true,
		})
	}
	return pkgs
}
