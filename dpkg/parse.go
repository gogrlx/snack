package dpkg

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseList parses dpkg-query -W -f='${Package}\t${Version}\t${Status}\n' output.
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
			Name:    parts[0],
			Version: parts[1],
		}
		if len(parts) == 3 && strings.Contains(parts[2], "install ok installed") {
			p.Installed = true
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseDpkgList parses dpkg-query -l output (table format with header).
func parseDpkgList(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		// Data lines start with status flags like "ii", "rc", etc.
		if len(line) < 4 || line[2] != ' ' {
			continue
		}
		// Skip lines where flags aren't letters (e.g. "+++-..." separator)
		if line[0] < 'a' || line[0] > 'z' {
			continue
		}
		status := line[:2]
		fields := strings.Fields(line[3:])
		if len(fields) < 2 {
			continue
		}
		p := snack.Package{
			Name:      fields[0],
			Version:   fields[1],
			Installed: status == "ii",
		}
		if len(fields) > 3 {
			p.Description = strings.Join(fields[3:], " ")
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseInfo parses dpkg-query -s output into a Package.
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
		case "Status":
			p.Installed = val == "install ok installed"
		}
	}
	if p.Name == "" {
		return nil, snack.ErrNotFound
	}
	return p, nil
}
