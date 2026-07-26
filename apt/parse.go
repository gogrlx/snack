package apt

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parseList parses dpkg-query -W output into packages.
func parseList(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
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
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		// Format: "package - description"
		parts := strings.SplitN(line, " - ", 2)
		p := snack.Package{Name: strings.TrimSpace(parts[0])}
		if len(parts) == 2 {
			p.Description = strings.TrimSpace(parts[1])
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parsePolicyCandidate extracts the Candidate version from apt-cache policy output.
// Returns empty string if no candidate is found or candidate is "(none)".
func parsePolicyCandidate(output string) string {
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "Candidate:"); ok {
			candidate := strings.TrimSpace(after)
			if candidate == "(none)" {
				return ""
			}
			return candidate
		}
	}
	return ""
}

// parsePolicyInstalled extracts the Installed version from apt-cache policy output.
// Returns empty string if not installed or "(none)".
func parsePolicyInstalled(output string) string {
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "Installed:"); ok {
			installed := strings.TrimSpace(after)
			if installed == "(none)" {
				return ""
			}
			return installed
		}
	}
	return ""
}

// parseUpgradeSimulation parses apt-get --just-print upgrade output.
// Lines starting with "Inst " indicate upgradable packages.
// Format: "Inst pkg [old-ver] (new-ver repo [arch])"
func parseUpgradeSimulation(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Inst ") {
			continue
		}
		line = strings.TrimPrefix(line, "Inst ")
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		parenStart := strings.Index(line, "(")
		parenEnd := strings.Index(line, ")")
		if parenStart < 0 || parenEnd < 0 {
			continue
		}
		verFields := strings.Fields(line[parenStart+1 : parenEnd])
		if len(verFields) < 1 {
			continue
		}
		pkgs = append(pkgs, snack.Package{
			Name:      name,
			Version:   verFields[0],
			Installed: true,
		})
	}
	return pkgs
}

// parseHoldList parses apt-mark showhold output (one package name per line).
func parseHoldList(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pkgs = append(pkgs, snack.Package{Name: line, Installed: true})
	}
	return pkgs
}

// parseFileList parses dpkg-query -L output (one file path per line).
func parseFileList(output string) []string {
	var files []string
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

// parseOwner parses dpkg -S output to extract the owning package name.
// Output format: "package: /path/to/file" or "pkg1, pkg2: /path".
// Returns the first package name.
func parseOwner(output string) string {
	line := strings.TrimSpace(strings.Split(output, "\n")[0])
	before, _, ok := strings.Cut(line, ":")
	if !ok {
		return ""
	}
	pkgPart := before
	if commaIdx := strings.Index(pkgPart, ","); commaIdx >= 0 {
		pkgPart = strings.TrimSpace(pkgPart[:commaIdx])
	}
	return strings.TrimSpace(pkgPart)
}

// parseSourcesLine parses a single deb/deb-src line from sources.list.
// Returns a Repository if the line is valid, or nil if it's a comment/blank.
func parseSourcesLine(line string) *snack.Repository {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}
	if !strings.HasPrefix(line, "deb ") && !strings.HasPrefix(line, "deb-src ") {
		return nil
	}
	return &snack.Repository{
		ID:      line,
		URL:     extractURL(line),
		Enabled: true,
		Type:    strings.Fields(line)[0],
	}
}

// extractURL pulls the URL from a deb/deb-src line.
func extractURL(line string) string {
	fields := strings.Fields(line)
	inOptions := false
	for i, f := range fields {
		if i == 0 {
			continue // skip deb/deb-src
		}
		if inOptions {
			if strings.HasSuffix(f, "]") {
				inOptions = false
			}
			continue
		}
		if strings.HasPrefix(f, "[") {
			if strings.HasSuffix(f, "]") {
				// Single-token options like [arch=amd64]
				continue
			}
			inOptions = true
			continue
		}
		return f
	}
	return ""
}

// parseInfo parses apt-cache show output into a Package.
func parseInfo(output string) (*snack.Package, error) {
	p := &snack.Package{}
	for line := range strings.SplitSeq(output, "\n") {
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
