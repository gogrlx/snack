package pacman

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseList parses the output of `pacman -Q` into packages.
// Each line is "name version".
func parseList(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		pkgs = append(pkgs, snack.Package{
			Name:      parts[0],
			Version:   parts[1],
			Installed: true,
		})
	}
	return pkgs
}

// parseSearch parses the output of `pacman -Ss` into packages.
// Format:
//
//	repo/name version [installed]
//	    Description text
func parseSearch(output string) []snack.Package {
	var pkgs []snack.Package
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		// repo/name version [installed]
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		repoName := strings.SplitN(parts[0], "/", 2)
		pkg := snack.Package{
			Version: parts[1],
		}
		if len(repoName) == 2 {
			pkg.Repository = repoName[0]
			pkg.Name = repoName[1]
		} else {
			pkg.Name = repoName[0]
		}
		for _, p := range parts[2:] {
			if strings.Contains(p, "installed") {
				pkg.Installed = true
			}
		}
		// Next line is description
		if i+1 < len(lines) {
			desc := strings.TrimSpace(lines[i+1])
			if desc != "" {
				pkg.Description = desc
				i++
			}
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// parseInfo parses the output of `pacman -Si` into a Package.
// Format is "Key : Value" lines.
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
		case "Architecture":
			pkg.Arch = val
		case "Repository":
			pkg.Repository = val
		}
	}
	if pkg.Name == "" {
		return nil
	}
	return pkg
}
