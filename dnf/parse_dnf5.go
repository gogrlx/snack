package dnf

import (
	"strings"

	"github.com/gogrlx/snack"
)

// stripPreamble removes dnf5 repository loading output that appears on stdout.
// It strips everything from "Updating and loading repositories:" through
// "Repositories loaded." inclusive.
func stripPreamble(output string) string {
	lines := strings.Split(output, "\n")
	var result []string
	inPreamble := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Updating and loading repositories:") {
			inPreamble = true
			continue
		}
		if inPreamble {
			if strings.HasPrefix(trimmed, "Repositories loaded.") {
				inPreamble = false
				continue
			}
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// parseListDNF5 parses `dnf5 list --installed` / `dnf5 list --upgrades` output.
// Format:
//
//	Installed packages
//	name.arch   version-release   repo-hash
func parseListDNF5(output string) []snack.Package {
	output = stripPreamble(output)
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Skip section headers
		lower := strings.ToLower(trimmed)
		if lower == "installed packages" || lower == "available packages" ||
			lower == "upgraded packages" || lower == "available upgrades" {
			continue
		}
		parts := strings.Fields(trimmed)
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

// parseSearchDNF5 parses `dnf5 search` output.
// Format:
//
//	Matched fields: name
//	 name.arch  Description text
//	Matched fields: summary
//	 name.arch  Description text
func parseSearchDNF5(output string) []snack.Package {
	output = stripPreamble(output)
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Matched fields:") {
			continue
		}
		// Lines are: "name.arch  Description"
		// Split on first double-space or use Fields
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			continue
		}
		nameArch := parts[0]
		if !strings.Contains(nameArch, ".") {
			continue
		}
		desc := strings.Join(parts[1:], " ")
		name, arch := parseArch(nameArch)
		pkgs = append(pkgs, snack.Package{
			Name:        name,
			Arch:        arch,
			Description: desc,
		})
	}
	return pkgs
}

// parseInfoDNF5 parses `dnf5 info` output.
// Format: "Key  : Value" lines with possible section headers like "Available packages".
func parseInfoDNF5(output string) *snack.Package {
	output = stripPreamble(output)
	pkg := &snack.Package{}
	for _, line := range strings.Split(output, "\n") {
		idx := strings.Index(line, " : ")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+3:])
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
		case "Repository", "From repo":
			pkg.Repository = val
		}
	}
	if pkg.Name == "" {
		return nil
	}
	return pkg
}

// parseRepoListDNF5 parses `dnf5 repolist --all` output.
// Same tabular format as dnf4, reuses the same logic.
func parseRepoListDNF5(output string) []snack.Repository {
	output = stripPreamble(output)
	return parseRepoList(output)
}

// parseGroupListDNF5 parses `dnf5 group list` tabular output.
// Format:
//
//	ID                          Name                      Installed
//	neuron-modelling-simulators Neuron Modelling Simulators     no
func parseGroupListDNF5(output string) []string {
	output = stripPreamble(output)
	var groups []string
	inBody := false
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !inBody {
			if strings.HasPrefix(trimmed, "ID") && strings.Contains(trimmed, "Name") {
				inBody = true
				continue
			}
			continue
		}
		// Last field is "yes"/"no", second-to-last through first space group is name.
		// Parse: first token is ID, last token is yes/no, middle is name.
		parts := strings.Fields(trimmed)
		if len(parts) < 3 {
			continue
		}
		// Last field is installed status
		status := parts[len(parts)-1]
		if status != "yes" && status != "no" {
			continue
		}
		name := strings.Join(parts[1:len(parts)-1], " ")
		groups = append(groups, name)
	}
	return groups
}

// parseGroupInfoDNF5 parses `dnf5 group info` output.
// Format:
//
//	Id                   : kde-desktop
//	Mandatory packages   : plasma-desktop
//	                     : plasma-workspace
//	Default packages     : NetworkManager-config-connectivity-fedora
func parseGroupInfoDNF5(output string) []snack.Package {
	output = stripPreamble(output)
	var pkgs []snack.Package
	inPkgSection := false
	for _, line := range strings.Split(output, "\n") {
		idx := strings.Index(line, " : ")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+3:])

		if key == "Mandatory packages" || key == "Default packages" ||
			key == "Optional packages" || key == "Conditional packages" {
			inPkgSection = true
			if val != "" {
				pkgs = append(pkgs, snack.Package{Name: val})
			}
			continue
		}

		// Continuation line: key is empty
		if key == "" && inPkgSection && val != "" {
			pkgs = append(pkgs, snack.Package{Name: val})
			continue
		}

		// Any other key ends the package section
		if key != "" {
			inPkgSection = false
		}
	}
	return pkgs
}
