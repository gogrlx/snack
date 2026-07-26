package rpm

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
	n, _ := parseArchSuffix(name)
	return n
}

// parseArchSuffix extracts the architecture from a package name if present.
func parseArchSuffix(name string) (string, string) {
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

// parseList parses rpm -qa --queryformat output.
// Format: "NAME\tVERSION-RELEASE\tSUMMARY\n"
func parseList(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		pkg := snack.Package{
			Name:      parts[0],
			Version:   parts[1],
			Installed: true,
		}
		if len(parts) >= 3 {
			pkg.Description = parts[2]
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// parseInfo parses rpm -qi output.
// Format is "Key  : Value" lines.
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
		case "Release":
			if pkg.Version != "" {
				pkg.Version = pkg.Version + "-" + val
			}
		case "Architecture", "Arch":
			pkg.Arch = val
		case "Summary":
			pkg.Description = val
		}
	}
	if pkg.Name == "" {
		return nil
	}
	return pkg
}
