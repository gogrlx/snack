//go:build linux

package dnf

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("dnf")
	return err == nil
}

// detectVersion checks whether the system dnf is dnf5 by inspecting
// `dnf --version` output. dnf5 prints "dnf5 version 5.x.x".
func (d *DNF) detectVersion() {
	if d.v5Set {
		return
	}
	d.v5Set = true
	out, err := exec.Command("dnf", "--version").CombinedOutput()
	if err != nil {
		return
	}
	d.v5 = strings.Contains(string(out), "dnf5")
}

// buildArgs constructs the command name and argument list from the base args
// and the provided options.
func buildArgs(baseArgs []string, opts snack.Options) (string, []string) {
	cmd := "dnf"
	args := make([]string, 0, len(baseArgs)+4)

	if opts.Root != "" {
		args = append(args, "--installroot="+opts.Root)
	}
	args = append(args, baseArgs...)
	if opts.AssumeYes {
		args = append(args, "-y")
	}
	if opts.DryRun {
		args = append(args, "--setopt=tsflags=test")
	}

	if opts.Sudo {
		args = append([]string{cmd}, args...)
		cmd = "sudo"
	}
	return cmd, args
}

func run(ctx context.Context, baseArgs []string, opts snack.Options) (string, error) {
	cmd, args := buildArgs(baseArgs, opts)
	c := exec.CommandContext(ctx, cmd, args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") || strings.Contains(se, "requires root") ||
			strings.Contains(se, "This command has to be run with superuser privileges") {
			return "", fmt.Errorf("dnf: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("dnf: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
}

// formatTargets converts targets to dnf CLI arguments.
// DNF uses "pkg-version" for version pinning.
func formatTargets(targets []snack.Target) []string {
	args := make([]string, 0, len(targets))
	for _, t := range targets {
		if t.Source != "" {
			args = append(args, t.Source)
		} else if t.Version != "" {
			args = append(args, t.Name+"-"+t.Version)
		} else {
			args = append(args, t.Name)
		}
	}
	return args
}

func install(ctx context.Context, v5 bool, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	o := snack.ApplyOptions(opts...)
	var toInstall []snack.Target
	var unchanged []string
	for _, t := range pkgs {
		if o.Reinstall || t.Version != "" || o.DryRun {
			toInstall = append(toInstall, t)
			continue
		}
		ok, err := isInstalled(ctx, t.Name, v5)
		if err != nil {
			return snack.InstallResult{}, err
		}
		if ok {
			unchanged = append(unchanged, t.Name)
		} else {
			toInstall = append(toInstall, t)
		}
	}
	if len(toInstall) > 0 {
		base := []string{"install", "-y"}
		if o.Refresh {
			base = append(base, "--refresh")
		}
		if o.FromRepo != "" {
			base = append(base, "--repo="+o.FromRepo)
		}
		if o.Reinstall {
			base[0] = "reinstall"
		}
		for _, t := range toInstall {
			if t.FromRepo != "" {
				base = append(base, "--repo="+t.FromRepo)
				break
			}
		}
		args := append(base, formatTargets(toInstall)...)
		if _, err := run(ctx, args, o); err != nil {
			return snack.InstallResult{}, err
		}
	}
	var installed []snack.Package
	for _, t := range toInstall {
		v, _ := version(ctx, t.Name, v5)
		installed = append(installed, snack.Package{Name: t.Name, Version: v, Installed: true})
	}
	return snack.InstallResult{Installed: installed, Unchanged: unchanged}, nil
}

func remove(ctx context.Context, v5 bool, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	o := snack.ApplyOptions(opts...)
	var toRemove []snack.Target
	var unchanged []string
	for _, t := range pkgs {
		if o.DryRun {
			toRemove = append(toRemove, t)
			continue
		}
		ok, err := isInstalled(ctx, t.Name, v5)
		if err != nil {
			return snack.RemoveResult{}, err
		}
		if !ok {
			unchanged = append(unchanged, t.Name)
		} else {
			toRemove = append(toRemove, t)
		}
	}
	if len(toRemove) > 0 {
		args := append([]string{"remove", "-y"}, snack.TargetNames(toRemove)...)
		if _, err := run(ctx, args, o); err != nil {
			return snack.RemoveResult{}, err
		}
	}
	var removed []snack.Package
	for _, t := range toRemove {
		removed = append(removed, snack.Package{Name: t.Name})
	}
	return snack.RemoveResult{Removed: removed, Unchanged: unchanged}, nil
}

func upgrade(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	_, err := run(ctx, []string{"upgrade", "-y"}, o)
	return err
}

func update(ctx context.Context) error {
	_, err := run(ctx, []string{"makecache"}, snack.Options{})
	return err
}

func listArgs(v5 bool) []string {
	if v5 {
		return []string{"list", "--installed"}
	}
	return []string{"list", "installed"}
}

func list(ctx context.Context, v5 bool) ([]snack.Package, error) {
	out, err := run(ctx, listArgs(v5), snack.Options{})
	if err != nil {
		return nil, fmt.Errorf("dnf list: %w", err)
	}
	if v5 {
		return parseListDNF5(out), nil
	}
	return parseList(out), nil
}

func search(ctx context.Context, query string, v5 bool) ([]snack.Package, error) {
	out, err := run(ctx, []string{"search", query}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("dnf search: %w", err)
	}
	if v5 {
		return parseSearchDNF5(out), nil
	}
	return parseSearch(out), nil
}

func info(ctx context.Context, pkg string, v5 bool) (*snack.Package, error) {
	out, err := run(ctx, []string{"info", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("dnf info %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("dnf info: %w", err)
	}
	var p *snack.Package
	if v5 {
		p = parseInfoDNF5(out)
	} else {
		p = parseInfo(out)
	}
	if p == nil {
		return nil, fmt.Errorf("dnf info %s: %w", pkg, snack.ErrNotFound)
	}
	return p, nil
}

func isInstalled(ctx context.Context, pkg string, v5 bool) (bool, error) {
	args := []string{"list", "installed", pkg}
	if v5 {
		args = []string{"list", "--installed", pkg}
	}
	c := exec.CommandContext(ctx, "dnf", args...)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("dnf isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string, v5 bool) (string, error) {
	args := append(listArgs(v5), pkg)
	out, err := run(ctx, args, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("dnf version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("dnf version: %w", err)
	}
	var pkgs []snack.Package
	if v5 {
		pkgs = parseListDNF5(out)
	} else {
		pkgs = parseList(out)
	}
	if len(pkgs) == 0 {
		return "", fmt.Errorf("dnf version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return pkgs[0].Version, nil
}

func upgradePackages(ctx context.Context, v5 bool, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	o := snack.ApplyOptions(opts...)
	var toUpgrade []snack.Target
	var unchanged []string
	for _, t := range pkgs {
		if o.DryRun {
			toUpgrade = append(toUpgrade, t)
			continue
		}
		ok, err := isInstalled(ctx, t.Name, v5)
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
		base := []string{"upgrade"}
		args := append(base, formatTargets(toUpgrade)...)
		if _, err := run(ctx, args, o); err != nil {
			return snack.InstallResult{}, err
		}
	}
	var upgraded []snack.Package
	for _, t := range toUpgrade {
		v, _ := version(ctx, t.Name, v5)
		upgraded = append(upgraded, snack.Package{Name: t.Name, Version: v, Installed: true})
	}
	return snack.InstallResult{Installed: upgraded, Unchanged: unchanged}, nil
}
