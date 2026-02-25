package snap

import (
	"strconv"
	"strings"

	"github.com/gogrlx/snack"
)

// parseSnapList parses `snap list` tabular output.
// Header: Name  Version  Rev  Tracking  Publisher  Notes
func parseSnapList(output string) []snack.Package {
	var pkgs []snack.Package
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i == 0 { // skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pkg := snack.Package{
			Name:      fields[0],
			Version:   fields[1],
			Installed: true,
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// parseSnapFind parses `snap find <query>` tabular output.
// Header: Name  Version  Publisher  Notes  Summary
func parseSnapFind(output string) []snack.Package {
	var pkgs []snack.Package
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i == 0 { // skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pkg := snack.Package{
			Name:    fields[0],
			Version: fields[1],
		}
		// Summary is everything after the 4th field
		if len(fields) > 4 {
			pkg.Description = strings.Join(fields[4:], " ")
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// parseSnapInfo parses `snap info <pkg>` key:value output.
func parseSnapInfo(output string) *snack.Package {
	pkg := &snack.Package{}
	for _, line := range strings.Split(output, "\n") {
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		switch key {
		case "name":
			pkg.Name = val
		case "summary":
			pkg.Description = val
		case "installed":
			// "installed: 1.2.3 (rev) 100MB ..."
			parts := strings.Fields(val)
			if len(parts) >= 1 {
				pkg.Version = parts[0]
				pkg.Installed = true
			}
		case "snap-id":
			// presence indicates it exists
		}
	}
	if pkg.Name == "" {
		return nil
	}
	return pkg
}

// parseSnapInfoVersion extracts the latest/stable version from `snap info` output.
func parseSnapInfoVersion(output string) string {
	// Look for "latest/stable:" line
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "latest/stable:") {
			val := strings.TrimPrefix(line, "latest/stable:")
			val = strings.TrimSpace(val)
			fields := strings.Fields(val)
			if len(fields) >= 1 && fields[0] != "--" && fields[0] != "^" {
				return fields[0]
			}
		}
	}
	return ""
}

// parseSnapRefreshList parses `snap refresh --list` tabular output.
// Header: Name  Version  Rev  Publisher  Notes
func parseSnapRefreshList(output string) []snack.Package {
	var pkgs []snack.Package
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i == 0 { // skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// "All snaps up to date." means no upgrades
		if strings.Contains(line, "All snaps up to date") {
			return nil
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pkgs = append(pkgs, snack.Package{
			Name:      fields[0],
			Version:   fields[1],
			Installed: true,
		})
	}
	return pkgs
}

// semverCmp does a basic semver-ish comparison.
// Returns -1 if a < b, 0 if equal, 1 if a > b.
func semverCmp(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var numA, numB int
		if i < len(partsA) {
			numA, _ = strconv.Atoi(stripNonNumeric(partsA[i]))
		}
		if i < len(partsB) {
			numB, _ = strconv.Atoi(stripNonNumeric(partsB[i]))
		}
		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
	}
	return 0
}

// stripNonNumeric keeps only leading digits from a string.
func stripNonNumeric(s string) string {
	for i, c := range s {
		if c < '0' || c > '9' {
			return s[:i]
		}
	}
	return s
}
