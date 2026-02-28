//go:build linux

package pacman

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("pacman")
	return err == nil
}

// buildArgs constructs the command name and argument list from the base args
// and the provided options.
func buildArgs(baseArgs []string, opts snack.Options) (string, []string) {
	cmd := "pacman"
	args := make([]string, 0, len(baseArgs)+4)

	if opts.Root != "" {
		args = append(args, "-r", opts.Root)
	}
	args = append(args, baseArgs...)
	if opts.AssumeYes {
		args = append(args, "--noconfirm")
	}
	if opts.DryRun {
		args = append(args, "--print")
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
		if strings.Contains(se, "permission denied") || strings.Contains(se, "requires root") {
			return "", fmt.Errorf("pacman: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("pacman: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
}

// formatTargets converts targets to pacman CLI arguments.
// Pacman uses "pkg=version" for version pinning.
func formatTargets(targets []snack.Target) []string {
	args := make([]string, 0, len(targets))
	for _, t := range targets {
		if t.Version != "" {
			args = append(args, t.Name+"="+t.Version)
		} else {
			args = append(args, t.Name)
		}
	}
	return args
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
	if len(toInstall) > 0 {
		base := []string{"-S", "--noconfirm"}
		if o.Refresh {
			base = []string{"-Sy", "--noconfirm"}
		}
		args := append(base, formatTargets(toInstall)...)
		if _, err := run(ctx, args, o); err != nil {
			return snack.InstallResult{}, err
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
	if len(toRemove) > 0 {
		args := append([]string{"-R", "--noconfirm"}, snack.TargetNames(toRemove)...)
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

func purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	args := append([]string{"-Rns", "--noconfirm"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args, o)
	return err
}

func upgrade(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	_, err := run(ctx, []string{"-Syu", "--noconfirm"}, o)
	return err
}

func update(ctx context.Context) error {
	_, err := run(ctx, []string{"-Sy"}, snack.Options{})
	return err
}

func list(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"-Q"}, snack.Options{})
	if err != nil {
		return nil, fmt.Errorf("pacman list: %w", err)
	}
	return parseList(out), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	out, err := run(ctx, []string{"-Ss", query}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("pacman search: %w", err)
	}
	return parseSearch(out), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	out, err := run(ctx, []string{"-Si", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("pacman info %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("pacman info: %w", err)
	}
	p := parseInfo(out)
	if p == nil {
		return nil, fmt.Errorf("pacman info %s: %w", pkg, snack.ErrNotFound)
	}
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	c := exec.CommandContext(ctx, "pacman", "-Q", pkg)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("pacman isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"-Q", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("pacman version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("pacman version: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) < 2 {
		return "", fmt.Errorf("pacman version %s: unexpected output %q", pkg, out)
	}
	return parts[1], nil
}
