package flatpak

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseList parses `flatpak list --columns=name,application,version,origin`.
// Format: "Name\tApplication ID\tVersion\tOrigin"
func parseList(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
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
	for _, line := range strings.Split(output, "\n") {
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
	for _, line := range strings.Split(output, "\n") {
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
