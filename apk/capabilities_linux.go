//go:build linux

package apk

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func latestVersion(ctx context.Context, pkg string) (string, error) {
	// Use `apk search -e` for exact match, output is "pkg-version"
	c := exec.CommandContext(ctx, "apk", "search", "-e", pkg)
	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("apk latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", fmt.Errorf("apk latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	// Output: "pkg-1.2.3-r0"
	_, ver := splitNameVersion(lines[0])
	if ver == "" {
		return "", fmt.Errorf("apk latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return ver, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	c := exec.CommandContext(ctx, "apk", "upgrade", "--simulate")
	out, err := c.CombinedOutput()
	if err != nil {
		// If simulate fails, it could mean no upgrades or an error
		outStr := strings.TrimSpace(string(out))
		if outStr == "" || strings.Contains(outStr, "OK") {
			return nil, nil
		}
		return nil, fmt.Errorf("apk listUpgrades: %w", err)
	}
	return parseUpgradeSimulation(string(out)), nil
}

// parseUpgradeSimulation parses `apk upgrade --simulate` output.
// Lines look like: "(1/3) Upgrading pkg (oldver -> newver)"
func parseUpgradeSimulation(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "Upgrading") {
			continue
		}
		// "(1/3) Upgrading pkg (oldver -> newver)"
		idx := strings.Index(line, "Upgrading ")
		if idx < 0 {
			continue
		}
		rest := line[idx+len("Upgrading "):]
		// "pkg (oldver -> newver)"
		parts := strings.SplitN(rest, " (", 2)
		if len(parts) < 1 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		var ver string
		if len(parts) == 2 {
			// "oldver -> newver)"
			verPart := strings.TrimSuffix(parts[1], ")")
			arrow := strings.Split(verPart, " -> ")
			if len(arrow) == 2 {
				ver = strings.TrimSpace(arrow[1])
			}
		}
		pkgs = append(pkgs, snack.Package{
			Name:      name,
			Version:   ver,
			Installed: true,
		})
	}
	return pkgs
}

func upgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	upgrades, err := listUpgrades(ctx)
	if err != nil {
		return false, err
	}
	for _, u := range upgrades {
		if u.Name == pkg {
			return true, nil
		}
	}
	return false, nil
}

func versionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	c := exec.CommandContext(ctx, "apk", "version", "-t", ver1, ver2)
	out, err := c.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("apk versionCmp: %w", err)
	}
	result := strings.TrimSpace(string(out))
	switch result {
	case "<":
		return -1, nil
	case ">":
		return 1, nil
	case "=":
		return 0, nil
	default:
		return 0, fmt.Errorf("apk versionCmp: unexpected output %q", result)
	}
}

// autoremove is a no-op for apk. Alpine's apk does not have a direct
// autoremove equivalent. `apk fix` repairs packages but does not remove
// unused dependencies.
func autoremove(_ context.Context, _ ...snack.Option) error {
	return nil
}

func clean(ctx context.Context) error {
	c := exec.CommandContext(ctx, "apk", "cache", "clean")
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apk clean: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func fileList(ctx context.Context, pkg string) ([]string, error) {
	c := exec.CommandContext(ctx, "apk", "info", "-L", pkg)
	out, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("apk fileList %s: %w", pkg, snack.ErrNotInstalled)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// First line is "pkg-version contains:" — skip it
		if strings.Contains(line, " contains:") {
			continue
		}
		files = append(files, line)
	}
	return files, nil
}

func owner(ctx context.Context, path string) (string, error) {
	c := exec.CommandContext(ctx, "apk", "info", "--who-owns", path)
	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("apk owner %s: %w", path, snack.ErrNotFound)
	}
	// Output: "/path is owned by pkg-version"
	outStr := strings.TrimSpace(string(out))
	if idx := strings.Index(outStr, "is owned by "); idx != -1 {
		nameVer := strings.TrimSpace(outStr[idx+len("is owned by "):])
		name, _ := splitNameVersion(nameVer)
		if name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("apk owner %s: unexpected output %q", path, outStr)
}
