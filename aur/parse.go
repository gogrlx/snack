package aur

import (
	"strings"

	"github.com/gogrlx/snack"
)

// parsePackageList parses "name version" lines from pacman -Q output.
func parsePackageList(output string) []snack.Package {
	var pkgs []snack.Package
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		pkgs = append(pkgs, snack.Package{
			Name:       parts[0],
			Version:    parts[1],
			Repository: "aur",
			Installed:  true,
		})
	}
	return pkgs
}
