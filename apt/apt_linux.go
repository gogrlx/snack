//go:build linux

package apt

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("apt-get")
	return err == nil
}

// formatTargets converts targets to apt CLI arguments.
// apt uses "pkg=version" for version pinning.
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

func buildArgs(command string, pkgs []snack.Target, opts ...snack.Option) []string {
	o := snack.ApplyOptions(opts...)
	var args []string
	if o.Sudo {
		args = append(args, "sudo")
	}
	args = append(args, "apt-get", command)
	if o.AssumeYes {
		args = append(args, "-y")
	}
	if o.DryRun {
		args = append(args, "--dry-run")
	}
	if o.FromRepo != "" {
		args = append(args, "-t", o.FromRepo)
	}
	if o.Reinstall && command == "install" {
		args = append(args, "--reinstall")
	}
	args = append(args, formatTargets(pkgs)...)
	return args
}

func runAptGet(ctx context.Context, command string, pkgs []snack.Target, opts ...snack.Option) error {
	args := buildArgs(command, pkgs, opts...)
	var cmd *exec.Cmd
	if args[0] == "sudo" {
		cmd = exec.CommandContext(ctx, args[0], args[1:]...)
	} else {
		cmd = exec.CommandContext(ctx, args[0], args[1:]...)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if strings.Contains(errMsg, "Permission denied") || strings.Contains(errMsg, "are you root") {
			return fmt.Errorf("apt-get %s: %w", command, snack.ErrPermissionDenied)
		}
		if strings.Contains(errMsg, "Unable to locate package") {
			return fmt.Errorf("apt-get %s: %w", command, snack.ErrNotFound)
		}
		return fmt.Errorf("apt-get %s: %w: %s", command, err, errMsg)
	}
	return nil
}

func install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	var toInstall []snack.Target
	var unchanged []string
	for _, t := range pkgs {
		ok, _ := isInstalled(ctx, t.Name)
		if ok {
			unchanged = append(unchanged, t.Name)
		} else {
			toInstall = append(toInstall, t)
		}
	}
	if len(toInstall) > 0 {
		if err := runAptGet(ctx, "install", toInstall, opts...); err != nil {
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
	var toRemove []snack.Target
	var unchanged []string
	for _, t := range pkgs {
		ok, _ := isInstalled(ctx, t.Name)
		if !ok {
			unchanged = append(unchanged, t.Name)
		} else {
			toRemove = append(toRemove, t)
		}
	}
	if len(toRemove) > 0 {
		if err := runAptGet(ctx, "remove", toRemove, opts...); err != nil {
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
	return runAptGet(ctx, "purge", pkgs, opts...)
}

func upgrade(ctx context.Context, opts ...snack.Option) error {
	return runAptGet(ctx, "upgrade", nil, opts...)
}

func update(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "apt-get", "update")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt-get update: %w: %s", err, stderr.String())
	}
	return nil
}

func list(ctx context.Context) ([]snack.Package, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Package}\\t${Version}\\t${Description}\\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dpkg-query list: %w", err)
	}
	return parseList(string(out)), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	cmd := exec.CommandContext(ctx, "apt-cache", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("apt-cache search: %w", err)
	}
	return parseSearch(string(out)), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	cmd := exec.CommandContext(ctx, "apt-cache", "show", pkg)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(stderr.String(), "No packages found") {
			return nil, fmt.Errorf("apt-cache show %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("apt-cache show %s: %w", pkg, err)
	}
	return parseInfo(string(out))
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Status}", pkg)
	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "install ok installed", nil
}

func version(ctx context.Context, pkg string) (string, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Version}", pkg)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("version %s: %w", pkg, snack.ErrNotInstalled)
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return "", fmt.Errorf("version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return v, nil
}
