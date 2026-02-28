package dnf

import (
	"strings"

	"github.com/gogrlx/snack"
)

// knownArchs is the set of RPM architecture suffixes.
var knownArchs = map[string]bool{
	"x86_64":  true,
	"i686":    true,
	"i386":    true,
	"aarch64": true,
	"armv7hl": true,
	"ppc64le": true,
	"s390x":   true,
	"noarch":  true,
	"src":     true,
}

// normalizeName strips architecture suffixes from a package name.
func normalizeName(name string) string {
	n, _ := parseArch(name)
	return n
}

// parseArch extracts the architecture from a package name if present.
// Returns the name without arch and the arch string.
func parseArch(name string) (string, string) {
	idx := strings.LastIndex(name, ".")
	if idx < 0 {
		return name, ""
	}
	arch := name[idx+1:]
	if knownArchs[arch] {
		return name[:idx], arch
	}
	return name, ""
}

// parseList parses the output of `dnf list installed` or `dnf list upgrades`.
// Format (after header line):
//
//	pkg-name.arch          version-release          repo
func parseList(output string) []snack.Package {
	var pkgs []snack.Package
	inBody := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip header lines until we see a line of dashes or the first package
		if !inBody {
			if strings.HasPrefix(line, "Installed Packages") ||
				strings.HasPrefix(line, "Available Upgrades") ||
				strings.HasPrefix(line, "Available Packages") ||
				strings.HasPrefix(line, "Upgraded Packages") {
				inBody = true
				continue
			}
			// Also skip metadata lines
			if strings.HasPrefix(line, "Last metadata") || strings.HasPrefix(line, "Updating") {
				continue
			}
			// If we see a line with fields that looks like a package, process it
			parts := strings.Fields(line)
			if len(parts) >= 2 && strings.Contains(parts[0], ".") {
				inBody = true
				// fall through to process
			} else {
				continue
			}
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		nameArch := parts[0]
		ver := parts[1]
		repo := ""
		if len(parts) >= 3 {
			repo = parts[2]
		}

		name, arch := parseArch(nameArch)
		pkgs = append(pkgs, snack.Package{
			Name:       name,
			Version:    ver,
			Arch:       arch,
			Repository: repo,
			Installed:  true,
		})
	}
	return pkgs
}

// parseSearch parses the output of `dnf search`.
// Format:
//
//	=== Name Exactly Matched: query ===
//	pkg-name.arch : Description text
//	=== Name & Summary Matched: query ===
//	pkg-name.arch : Description text
func parseSearch(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "===") || strings.HasPrefix(line, "Last metadata") {
			continue
		}
		// "pkg-name.arch : Description"
		idx := strings.Index(line, " : ")
		if idx < 0 {
			continue
		}
		nameArch := strings.TrimSpace(line[:idx])
		desc := strings.TrimSpace(line[idx+3:])
		name, arch := parseArch(nameArch)
		pkgs = append(pkgs, snack.Package{
			Name:        name,
			Arch:        arch,
			Description: desc,
		})
	}
	return pkgs
}

// parseInfo parses the output of `dnf info`.
// Format is "Key  : Value" lines.
func parseInfo(output string) *snack.Package {
	pkg := &snack.Package{}
	for _, line := range strings.Split(output, "\n") {
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		switch key {
		case "Name":
			pkg.Name = val
		case "Version":
			pkg.Version = val
		case "Release":
			if pkg.Version != "" {
				pkg.Version = pkg.Version + "-" + val
			}
		case "Architecture", "Arch":
			pkg.Arch = val
		case "Summary":
			pkg.Description = val
		case "From repo", "Repository":
			pkg.Repository = val
		}
	}
	if pkg.Name == "" {
		return nil
	}
	return pkg
}

// parseVersionLock parses `dnf versionlock list` output.
// Lines are typically package NEVRA patterns like "pkg-0:1.2.3-4.el9.*"
func parseVersionLock(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Last metadata") || strings.HasPrefix(line, "Adding") {
			continue
		}
		// Try to extract name from NEVRA pattern
		// Format: "name-epoch:version-release.arch" or simpler variants
		name := line
		// Strip trailing .* glob
		name = strings.TrimSuffix(name, ".*")
		// Strip arch
		name, _ = parseArch(name)
		// Strip version-release: find epoch pattern like "-0:" or last "-" before version
		// Try epoch pattern first: "name-epoch:version-release"
		for {
			idx := strings.LastIndex(name, "-")
			if idx <= 0 {
				break
			}
			rest := name[idx+1:]
			// Check if what follows looks like a version or epoch (digit or epoch:)
			if len(rest) > 0 && (rest[0] >= '0' && rest[0] <= '9') {
				name = name[:idx]
			} else {
				break
			}
		}
		if name != "" {
			pkgs = append(pkgs, snack.Package{Name: name, Installed: true})
		}
	}
	return pkgs
}

// parseRepoList parses `dnf repolist --all` output.
// Format:
//
//	repo id                    repo name                       status
//	appstream                  CentOS Stream 9 - AppStream     enabled
func parseRepoList(output string) []snack.Repository {
	var repos []snack.Repository
	inBody := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !inBody {
			if strings.HasPrefix(line, "repo id") {
				inBody = true
				continue
			}
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		id := parts[0]
		enabled := false
		status := parts[len(parts)-1]
		if status == "enabled" {
			enabled = true
		}
		// Name is everything between id and status
		name := strings.Join(parts[1:len(parts)-1], " ")
		repos = append(repos, snack.Repository{
			ID:      id,
			Name:    name,
			Enabled: enabled,
		})
	}
	return repos
}

// parseGroupList parses `dnf group list` output.
// Groups are listed under section headers like "Available Groups:" / "Installed Groups:".
func parseGroupList(output string) []string {
	var groups []string
	inSection := false
	for _, line := range strings.Split(output, "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "Groups:") ||
			strings.HasSuffix(strings.TrimSpace(line), "groups:") {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			inSection = false
			continue
		}
		groups = append(groups, trimmed)
	}
	return groups
}

// parseGroupIsInstalled checks whether a named group appears under "Installed Groups:"
// in `dnf group list` output.
func parseGroupIsInstalled(output, group string) bool {
	inInstalled := false
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, "Groups:") || strings.HasSuffix(trimmed, "groups:") {
			inInstalled = strings.HasPrefix(strings.ToLower(trimmed), "installed")
			continue
		}
		if !inInstalled || trimmed == "" {
			continue
		}
		if strings.EqualFold(trimmed, group) {
			return true
		}
	}
	return false
}

// parseGroupInfo parses `dnf group info <group>` output.
// Packages are listed under section headers like "Mandatory Packages:", "Default Packages:", "Optional Packages:".
func parseGroupInfo(output string) []snack.Package {
	var pkgs []snack.Package
	inPkgSection := false
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			inPkgSection = false
			continue
		}
		if strings.HasSuffix(trimmed, "Packages:") || strings.HasSuffix(trimmed, "packages:") {
			inPkgSection = true
			continue
		}
		if !inPkgSection {
			continue
		}
		name := trimmed
		// Strip leading marks like "=" or "-" or "+"
		if len(name) > 2 && (name[0] == '=' || name[0] == '-' || name[0] == '+') && name[1] == ' ' {
			name = name[2:]
		}
		name = strings.TrimSpace(name)
		if name != "" {
			pkgs = append(pkgs, snack.Package{Name: name})
		}
	}
	return pkgs
}
