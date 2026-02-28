//go:build linux

package pacman

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gogrlx/snack"
)

func latestVersion(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"-Si", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("pacman latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("pacman latestVersion: %w", err)
	}
	p := parseInfo(out)
	if p == nil || p.Version == "" {
		return "", fmt.Errorf("pacman latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return p.Version, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"-Qu"}, snack.Options{})
	if err != nil {
		// exit status 1 means no upgrades available
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("pacman listUpgrades: %w", err)
	}
	return parseUpgrades(out), nil
}

// parseUpgrades parses `pacman -Qu` output.
// Format: "pkg oldver -> newver"
func parseUpgrades(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// "pkg oldver -> newver"
		parts := strings.Fields(line)
		if len(parts) >= 4 && parts[2] == "->" {
			pkgs = append(pkgs, snack.Package{
				Name:      parts[0],
				Version:   parts[3],
				Installed: true,
			})
		} else if len(parts) >= 2 {
			// fallback: "pkg newver"
			pkgs = append(pkgs, snack.Package{
				Name:      parts[0],
				Version:   parts[1],
				Installed: true,
			})
		}
	}
	return pkgs
}

func upgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	_, err := run(ctx, []string{"-Qu", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return false, nil
		}
		return false, fmt.Errorf("pacman upgradeAvailable: %w", err)
	}
	return true, nil
}

func versionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	c := exec.CommandContext(ctx, "vercmp", ver1, ver2)
	out, err := c.Output()
	if err != nil {
		return 0, fmt.Errorf("vercmp: %w", err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("vercmp: unexpected output %q: %w", string(out), err)
	}
	// Normalize to -1, 0, 1
	switch {
	case n < 0:
		return -1, nil
	case n > 0:
		return 1, nil
	default:
		return 0, nil
	}
}

func autoremove(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)

	// First get list of orphans
	orphans, err := run(ctx, []string{"-Qdtq"}, snack.Options{})
	if err != nil {
		// exit status 1 means no orphans
		if strings.Contains(err.Error(), "exit status 1") {
			return nil
		}
		return fmt.Errorf("pacman autoremove: %w", err)
	}
	orphans = strings.TrimSpace(orphans)
	if orphans == "" {
		return nil
	}

	pkgs := strings.Fields(orphans)
	args := append([]string{"-Rns", "--noconfirm"}, pkgs...)
	_, err = run(ctx, args, o)
	return err
}

func clean(ctx context.Context) error {
	_, err := run(ctx, []string{"-Sc", "--noconfirm"}, snack.Options{})
	return err
}

func fileList(ctx context.Context, pkg string) ([]string, error) {
	out, err := run(ctx, []string{"-Ql", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("pacman fileList %s: %w", pkg, snack.ErrNotInstalled)
		}
		return nil, fmt.Errorf("pacman fileList: %w", err)
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// "pkg /path/to/file"
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			files = append(files, strings.TrimSpace(parts[1]))
		}
	}
	return files, nil
}

func owner(ctx context.Context, path string) (string, error) {
	out, err := run(ctx, []string{"-Qo", path}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") || strings.Contains(err.Error(), "No package owns") {
			return "", fmt.Errorf("pacman owner %s: %w", path, snack.ErrNotFound)
		}
		return "", fmt.Errorf("pacman owner: %w", err)
	}
	// Output: "/path is owned by pkg ver"
	out = strings.TrimSpace(out)
	if idx := strings.Index(out, "is owned by "); idx != -1 {
		remainder := out[idx+len("is owned by "):]
		parts := strings.Fields(remainder)
		if len(parts) >= 1 {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("pacman owner %s: unexpected output %q", path, out)
}

func groupList(ctx context.Context) ([]string, error) {
	out, err := run(ctx, []string{"-Sg"}, snack.Options{})
	if err != nil {
		return nil, fmt.Errorf("pacman groupList: %w", err)
	}
	seen := make(map[string]struct{})
	var groups []string
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			g := parts[0]
			if _, ok := seen[g]; !ok {
				seen[g] = struct{}{}
				groups = append(groups, g)
			}
		}
	}
	return groups, nil
}

func groupInfo(ctx context.Context, group string) ([]snack.Package, error) {
	out, err := run(ctx, []string{"-Sg", group}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("pacman groupInfo %s: %w", group, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("pacman groupInfo: %w", err)
	}
	var pkgs []snack.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// "group pkg"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			pkgs = append(pkgs, snack.Package{Name: parts[1]})
		}
	}
	return pkgs, nil
}

func groupInstall(ctx context.Context, group string, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	_, err := run(ctx, []string{"-S", "--noconfirm", group}, o)
	return err
}

// parseGroupPkgSet parses "group pkg" lines and returns the set of package names.
func parseGroupPkgSet(output string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			set[parts[1]] = struct{}{}
		}
	}
	return set
}

func groupIsInstalled(ctx context.Context, group string) (bool, error) {
	// Get all packages in the group from the sync database.
	availOut, err := run(ctx, []string{"-Sg", group}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return false, fmt.Errorf("pacman groupIsInstalled %s: %w", group, snack.ErrNotFound)
		}
		return false, fmt.Errorf("pacman groupIsInstalled: %w", err)
	}
	avail := parseGroupPkgSet(availOut)
	if len(avail) == 0 {
		return false, fmt.Errorf("pacman groupIsInstalled %s: %w", group, snack.ErrNotFound)
	}
	// Get packages from the group that are installed (local database).
	instOut, err := run(ctx, []string{"-Qg", group}, snack.Options{})
	if err != nil {
		// exit status 1 means nothing from this group is installed.
		if strings.Contains(err.Error(), "exit status 1") {
			return false, nil
		}
		return false, fmt.Errorf("pacman groupIsInstalled: %w", err)
	}
	inst := parseGroupPkgSet(instOut)
	for pkg := range avail {
		if _, ok := inst[pkg]; !ok {
			return false, nil
		}
	}
	return true, nil
}
