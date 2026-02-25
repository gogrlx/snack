package apt

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseList parses dpkg-query -W output into packages.
func parseList(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		p := snack.Package{
			Name:      parts[0],
			Version:   parts[1],
			Installed: true,
		}
		if len(parts) == 3 {
			p.Description = parts[2]
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseSearch parses apt-cache search output.
func parseSearch(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		// Format: "package - description"
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) < 1 {
			continue
		}
		p := snack.Package{Name: strings.TrimSpace(parts[0])}
		if len(parts) == 2 {
			p.Description = strings.TrimSpace(parts[1])
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseInfo parses apt-cache show output into a Package.
func parseInfo(output string) (*snack.Package, error) {
	p := &snack.Package{}
	for _, line := range strings.Split(output, "\n") {
		key, val, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}
		switch key {
		case "Package":
			p.Name = val
		case "Version":
			p.Version = val
		case "Description":
			p.Description = val
		case "Architecture":
			p.Arch = val
		}
	}
	if p.Name == "" {
		return nil, snack.ErrNotFound
	}
	return p, nil
}
