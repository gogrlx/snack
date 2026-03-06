//go:build windows

package winget

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("winget")
	return err == nil
}

// commonArgs returns flags used by all winget commands for non-interactive operation.
func commonArgs() []string {
	return []string{
		"--accept-source-agreements",
		"--disable-interactivity",
	}
}

func run(ctx context.Context, args []string) (string, error) {
	c := exec.CommandContext(ctx, "winget", args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	out := stdout.String()
	// winget writes progress and VT sequences to stdout; strip them.
	out = stripProgress(out)
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "Access is denied") ||
			strings.Contains(out, "administrator") {
			return "", fmt.Errorf("winget: %w", snack.ErrPermissionDenied)
		}
		combined := se + out
		if strings.Contains(combined, "No package found") ||
			strings.Contains(combined, "No installed package found") {
			return "", fmt.Errorf("winget: %w", snack.ErrNotFound)
		}
		errMsg := strings.TrimSpace(se)
		if errMsg == "" {
			errMsg = strings.TrimSpace(out)
		}
		return "", fmt.Errorf("winget: %s: %w", errMsg, err)
	}
	return out, nil
}

// stripProgress removes VT100 escape sequences and progress lines from output.
func stripProgress(s string) string {
	var b strings.Builder
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		clean := stripVT(line)
		clean = strings.TrimRight(clean, "\r")
		// Skip pure progress lines (e.g. "██████████████  50%")
		if isProgressLine(clean) {
			continue
		}
		b.WriteString(clean)
		b.WriteByte('\n')
	}
	return b.String()
}

func install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	o := snack.ApplyOptions(opts...)
	var toInstall []snack.Target
	var unchanged []string
	for _, t := range pkgs {
		if o.Reinstall || t.Version != "" || o.DryRun {
			toInstall = append(toInstall, t)
			continue
		}
		ok, err := isInstalled(ctx, t.Name)
		if err != nil {
			return snack.InstallResult{}, err
		}
		if ok {
			unchanged = append(unchanged, t.Name)
		} else {
			toInstall = append(toInstall, t)
		}
	}
	for _, t := range toInstall {
		args := []string{"install", "--id", t.Name, "--exact", "--silent"}
		args = append(args, commonArgs()...)
		args = append(args, "--accept-package-agreements")
		if t.Version != "" {
			args = append(args, "--version", t.Version)
		}
		if t.FromRepo != "" {
			args = append(args, "--source", t.FromRepo)
		} else if o.FromRepo != "" {
			args = append(args, "--source", o.FromRepo)
		}
		if _, err := run(ctx, args); err != nil {
			return snack.InstallResult{}, fmt.Errorf("winget install %s: %w", t.Name, err)
		}
	}
	var installed []snack.Package
	for _, t := range toInstall {
		v, _ := version(ctx, t.Name)
		installed = append(installed, snack.Package{Name: t.Name, Version: v, Installed: true})
	}
	return snack.InstallResult{Installed: installed, Unchanged: unchanged}, nil
}

func remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	o := snack.ApplyOptions(opts...)
	var toRemove []snack.Target
	var unchanged []string
	for _, t := range pkgs {
		if o.DryRun {
			toRemove = append(toRemove, t)
			continue
		}
		ok, err := isInstalled(ctx, t.Name)
		if err != nil {
			return snack.RemoveResult{}, err
		}
		if !ok {
			unchanged = append(unchanged, t.Name)
		} else {
			toRemove = append(toRemove, t)
		}
	}
	for _, t := range toRemove {
		args := []string{"uninstall", "--id", t.Name, "--exact", "--silent"}
		args = append(args, commonArgs()...)
		if _, err := run(ctx, args); err != nil {
			return snack.RemoveResult{}, fmt.Errorf("winget uninstall %s: %w", t.Name, err)
		}
	}
	var removed []snack.Package
	for _, t := range toRemove {
		removed = append(removed, snack.Package{Name: t.Name})
	}
	return snack.RemoveResult{Removed: removed, Unchanged: unchanged}, nil
}

func purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	for _, t := range pkgs {
		args := []string{"uninstall", "--id", t.Name, "--exact", "--silent", "--purge"}
		args = append(args, commonArgs()...)
		if _, err := run(ctx, args); err != nil {
			return fmt.Errorf("winget purge %s: %w", t.Name, err)
		}
	}
	return nil
}

func upgrade(ctx context.Context, _ ...snack.Option) error {
	args := []string{"upgrade", "--all", "--silent"}
	args = append(args, commonArgs()...)
	args = append(args, "--accept-package-agreements")
	_, err := run(ctx, args)
	return err
}

func update(ctx context.Context) error {
	_, err := run(ctx, []string{"source", "update"})
	return err
}

func list(ctx context.Context) ([]snack.Package, error) {
	args := []string{"list"}
	args = append(args, commonArgs()...)
	out, err := run(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("winget list: %w", err)
	}
	return parseTable(out, true), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	args := []string{"search", query}
	args = append(args, commonArgs()...)
	out, err := run(ctx, args)
	if err != nil {
		if strings.Contains(err.Error(), string(snack.ErrNotFound.Error())) {
			return nil, nil
		}
		return nil, fmt.Errorf("winget search: %w", err)
	}
	return parseTable(out, false), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	args := []string{"show", "--id", pkg, "--exact"}
	args = append(args, commonArgs()...)
	out, err := run(ctx, args)
	if err != nil {
		if strings.Contains(err.Error(), string(snack.ErrNotFound.Error())) {
			return nil, fmt.Errorf("winget info %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("winget info: %w", err)
	}
	p := parseShow(out)
	if p == nil {
		return nil, fmt.Errorf("winget info %s: %w", pkg, snack.ErrNotFound)
	}
	// Check if installed
	ok, _ := isInstalled(ctx, pkg)
	p.Installed = ok
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	args := []string{"list", "--id", pkg, "--exact"}
	args = append(args, commonArgs()...)
	out, err := run(ctx, args)
	if err != nil {
		// "No installed package found" is returned as an error by run()
		if strings.Contains(err.Error(), string(snack.ErrNotFound.Error())) {
			return false, nil
		}
		return false, fmt.Errorf("winget isInstalled: %w", err)
	}
	pkgs := parseTable(out, true)
	return len(pkgs) > 0, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	args := []string{"list", "--id", pkg, "--exact"}
	args = append(args, commonArgs()...)
	out, err := run(ctx, args)
	if err != nil {
		if strings.Contains(err.Error(), string(snack.ErrNotFound.Error())) {
			return "", fmt.Errorf("winget version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("winget version: %w", err)
	}
	pkgs := parseTable(out, true)
	if len(pkgs) == 0 {
		return "", fmt.Errorf("winget version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return pkgs[0].Version, nil
}

func latestVersion(ctx context.Context, pkg string) (string, error) {
	args := []string{"show", "--id", pkg, "--exact"}
	args = append(args, commonArgs()...)
	out, err := run(ctx, args)
	if err != nil {
		if strings.Contains(err.Error(), string(snack.ErrNotFound.Error())) {
			return "", fmt.Errorf("winget latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("winget latestVersion: %w", err)
	}
	p := parseShow(out)
	if p == nil || p.Version == "" {
		return "", fmt.Errorf("winget latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return p.Version, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	args := []string{"upgrade"}
	args = append(args, commonArgs()...)
	out, err := run(ctx, args)
	if err != nil {
		// No upgrades available may exit non-zero on some versions
		return nil, nil
	}
	return parseTable(out, false), nil
}

func upgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	upgrades, err := listUpgrades(ctx)
	if err != nil {
		return false, err
	}
	for _, u := range upgrades {
		if strings.EqualFold(u.Name, pkg) {
			return true, nil
		}
	}
	return false, nil
}

func versionCmp(_ context.Context, ver1, ver2 string) (int, error) {
	return semverCmp(ver1, ver2), nil
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
	for _, t := range toUpgrade {
		args := []string{"upgrade", "--id", t.Name, "--exact", "--silent"}
		args = append(args, commonArgs()...)
		args = append(args, "--accept-package-agreements")
		if t.Version != "" {
			args = append(args, "--version", t.Version)
		}
		if _, err := run(ctx, args); err != nil {
			return snack.InstallResult{}, fmt.Errorf("winget upgrade %s: %w", t.Name, err)
		}
	}
	var upgraded []snack.Package
	for _, t := range toUpgrade {
		v, _ := version(ctx, t.Name)
		upgraded = append(upgraded, snack.Package{Name: t.Name, Version: v, Installed: true})
	}
	return snack.InstallResult{Installed: upgraded, Unchanged: unchanged}, nil
}

// sourceList returns configured winget sources.
func sourceList(ctx context.Context) ([]snack.Repository, error) {
	args := []string{"source", "list"}
	args = append(args, commonArgs()...)
	out, err := run(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("winget source list: %w", err)
	}
	return parseSourceList(out), nil
}

// sourceAdd adds a new winget source.
func sourceAdd(ctx context.Context, repo snack.Repository) error {
	args := []string{"source", "add", "--name", repo.Name, "--arg", repo.URL}
	args = append(args, commonArgs()...)
	if repo.Type != "" {
		args = append(args, "--type", repo.Type)
	}
	_, err := run(ctx, args)
	return err
}

// sourceRemove removes a winget source.
func sourceRemove(ctx context.Context, name string) error {
	args := []string{"source", "remove", "--name", name}
	args = append(args, commonArgs()...)
	_, err := run(ctx, args)
	return err
}
