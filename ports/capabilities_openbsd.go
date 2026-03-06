//go:build openbsd

package ports

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogrlx/snack"
)

func latestVersion(ctx context.Context, pkg string) (string, error) {
	// pkg_info -Q returns available packages matching the query.
	out, err := runCmd(ctx, "pkg_info", []string{"-Q", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("ports latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("ports latestVersion: %w", err)
	}
	// Find the best match: exact name match with highest version.
	var best string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, ver := splitNameVersion(line)
		if name == pkg && ver != "" {
			if best == "" || ver > best {
				best = ver
			}
		}
	}
	if best == "" {
		return "", fmt.Errorf("ports latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return best, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	// pkg_add -u -n simulates upgrading all packages.
	out, err := runCmd(ctx, "pkg_add", []string{"-u", "-n"}, snack.Options{})
	if err != nil {
		// Exit status 1 with no output means nothing to upgrade.
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("ports listUpgrades: %w", err)
	}
	return parseUpgradeOutput(out), nil
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

func versionCmp(_ context.Context, ver1, ver2 string) (int, error) {
	// OpenBSD has no native version comparison tool.
	// Use simple string comparison.
	switch {
	case ver1 < ver2:
		return -1, nil
	case ver1 > ver2:
		return 1, nil
	default:
		return 0, nil
	}
}

func upgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	o := snack.ApplyOptions(opts...)
	var toUpgrade []snack.Target
	var unchanged []string
	for _, t := range pkgs {
		if o.DryRun {
			toUpgrade = append(toUpgrade, t)
			continue
		}
		ok, err := isInstalled(ctx, t.Name)
		if err != nil {
			return snack.InstallResult{}, err
		}
		if !ok {
			unchanged = append(unchanged, t.Name)
		} else {
			toUpgrade = append(toUpgrade, t)
		}
	}
	if len(toUpgrade) > 0 {
		args := append([]string{"-u"}, snack.TargetNames(toUpgrade)...)
		if _, err := runCmd(ctx, "pkg_add", args, o); err != nil {
			return snack.InstallResult{}, err
		}
	}
	var upgraded []snack.Package
	for _, t := range toUpgrade {
		v, _ := version(ctx, t.Name)
		upgraded = append(upgraded, snack.Package{Name: t.Name, Version: v, Installed: true})
	}
	return snack.InstallResult{Installed: upgraded, Unchanged: unchanged}, nil
}
