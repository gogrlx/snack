package flatpak

import (
	"strconv"
	"strings"

	"github.com/gogrlx/snack"
)

// parseList parses `flatpak list --columns=name,application,version,origin`.
// Format: "Name\tApplication ID\tVersion\tOrigin"
func parseList(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		pkg := snack.Package{
			Name:      strings.TrimSpace(parts[0]),
			Installed: true,
		}
		if len(parts) >= 2 {
			pkg.Description = strings.TrimSpace(parts[1]) // application ID
		}
		if len(parts) >= 3 {
			pkg.Version = strings.TrimSpace(parts[2])
		}
		if len(parts) >= 4 {
			pkg.Repository = strings.TrimSpace(parts[3])
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// parseSearch parses `flatpak search <query> --columns=name,application,version,remotes`.
// Format: "Name\tApplication ID\tVersion\tRemotes"
func parseSearch(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "No matches found") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		pkg := snack.Package{
			Name: strings.TrimSpace(parts[0]),
		}
		if len(parts) >= 2 {
			pkg.Description = strings.TrimSpace(parts[1])
		}
		if len(parts) >= 3 {
			pkg.Version = strings.TrimSpace(parts[2])
		}
		if len(parts) >= 4 {
			pkg.Repository = strings.TrimSpace(parts[3])
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// parseInfo parses `flatpak info <pkg>` output (key: value format).
func parseInfo(output string) *snack.Package {
	pkg := &snack.Package{}
	for line := range strings.SplitSeq(output, "\n") {
		before, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key := strings.TrimSpace(before)
		val := strings.TrimSpace(after)
		switch key {
		case "Name":
			pkg.Name = val
		case "Version":
			pkg.Version = val
		case "Description":
			pkg.Description = val
		case "Arch":
			pkg.Arch = val
		case "Origin":
			pkg.Repository = val
		}
	}
	if pkg.Name == "" {
		return nil
	}
	return pkg
}

// parseRemotes parses `flatpak remotes --columns=name,url,options`.
// Format: "Name\tURL\tOptions"
func parseRemotes(output string) []snack.Repository {
	var repos []snack.Repository
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 1 {
			continue
		}
		repo := snack.Repository{
			ID:      strings.TrimSpace(parts[0]),
			Name:    strings.TrimSpace(parts[0]),
			Enabled: true,
		}
		if len(parts) >= 2 {
			repo.URL = strings.TrimSpace(parts[1])
		}
		if len(parts) >= 3 {
			opts := strings.TrimSpace(parts[2])
			if strings.Contains(opts, "disabled") {
				repo.Enabled = false
			}
		}
		repos = append(repos, repo)
	}
	return repos
}

// semverCmp does a basic semver-ish comparison.
// Returns -1 if a < b, 0 if equal, 1 if a > b.
func semverCmp(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := max(len(partsB), len(partsA))

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
