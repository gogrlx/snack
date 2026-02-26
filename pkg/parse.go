package pkg

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseQuery parses the output of `pkg query '%n\t%v\t%c'`.
func parseQuery(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
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

// parseSearch parses the output of `pkg search <query>`.
// Format: "name-version                Comment text"
func parseSearch(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Split on whitespace; first field is name-version
		parts := strings.SplitN(line, " ", 2)
		nameVer := parts[0]
		name, ver := splitNameVersion(nameVer)
		p := snack.Package{
			Name:    name,
			Version: ver,
		}
		if len(parts) == 2 {
			p.Description = strings.TrimSpace(parts[1])
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseInfo parses the output of `pkg info <pkg>`.
// Format is "Key: Value" lines.
func parseInfo(output string) *snack.Package {
	pkg := &snack.Package{Installed: true}
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
		case "Comment":
			pkg.Description = val
		case "Arch":
			pkg.Arch = val
		}
	}
	if pkg.Name == "" {
		return nil
	}
	return pkg
}

// parseUpgrades parses the output of `pkg upgrade -n`.
// Looks for lines like:
//
//	Installing pkg-name: oldver -> newver
//	Upgrading pkg-name: oldver -> newver
func parseUpgrades(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// Look for "Upgrading" or "Reinstalling" lines with ->
		if !strings.Contains(line, "->") {
			continue
		}
		var nameVer string
		if strings.HasPrefix(line, "Upgrading ") {
			nameVer = strings.TrimPrefix(line, "Upgrading ")
		} else if strings.HasPrefix(line, "Installing ") {
			nameVer = strings.TrimPrefix(line, "Installing ")
		} else if strings.HasPrefix(line, "Reinstalling ") {
			nameVer = strings.TrimPrefix(line, "Reinstalling ")
		} else {
			continue
		}
		// "name: oldver -> newver"
		colonIdx := strings.Index(nameVer, ":")
		if colonIdx < 0 {
			continue
		}
		name := strings.TrimSpace(nameVer[:colonIdx])
		rest := strings.TrimSpace(nameVer[colonIdx+1:])
		parts := strings.Fields(rest)
		if len(parts) >= 3 && parts[1] == "->" {
			pkgs = append(pkgs, snack.Package{
				Name:      name,
				Version:   parts[2],
				Installed: true,
			})
		}
	}
	return pkgs
}

// parseFileList parses `pkg info -l <pkg>` output.
// Lines starting with "/" after the header are file paths.
func parseFileList(output string) []string {
	var files []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/") {
			files = append(files, line)
		}
	}
	return files
}

// parseOwner parses `pkg which <path>` output.
// Format: "/path was installed by package name-version"
func parseOwner(output string) string {
	output = strings.TrimSpace(output)
	if idx := strings.Index(output, "was installed by package "); idx != -1 {
		remainder := output[idx+len("was installed by package "):]
		name, _ := splitNameVersion(strings.TrimSpace(remainder))
		return name
	}
	return output
}

// splitNameVersion splits "name-version" into name and version.
// FreeBSD pkg uses the last hyphen as separator (name can contain hyphens).
func splitNameVersion(s string) (string, string) {
	idx := strings.LastIndex(s, "-")
	if idx <= 0 {
		return s, ""
	}
	return s[:idx], s[idx+1:]
}
