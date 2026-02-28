//go:build openbsd

package ports

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("pkg_add")
	return err == nil
}

func runCmd(ctx context.Context, name string, args []string, opts snack.Options) (string, error) {
	cmdName := name
	cmdArgs := make([]string, 0, len(args)+2)
	cmdArgs = append(cmdArgs, args...)

	if opts.Sudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmdName = "sudo"
	}

	c := exec.CommandContext(ctx, cmdName, cmdArgs...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") || strings.Contains(se, "need root") {
			return "", fmt.Errorf("ports: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("ports: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
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
	o := snack.ApplyOptions(opts...)
	if len(toInstall) > 0 {
		if _, err := runCmd(ctx, "pkg_add", snack.TargetNames(toInstall), o); err != nil {
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
	o := snack.ApplyOptions(opts...)
	if len(toRemove) > 0 {
		if _, err := runCmd(ctx, "pkg_delete", snack.TargetNames(toRemove), o); err != nil {
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
	args := append([]string{"-c"}, snack.TargetNames(pkgs)...)
	_, err := runCmd(ctx, "pkg_delete", args, o)
	return err
}

func upgrade(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	_, err := runCmd(ctx, "pkg_add", []string{"-u"}, o)
	return err
}

func update(_ context.Context) error {
	// No-op on OpenBSD; updates handled via fw_update or syspatch.
	return nil
}

func list(ctx context.Context) ([]snack.Package, error) {
	out, err := runCmd(ctx, "pkg_info", nil, snack.Options{})
	if err != nil {
		return nil, fmt.Errorf("ports list: %w", err)
	}
	return parseList(out), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	out, err := runCmd(ctx, "pkg_info", []string{"-Q", query}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("ports search: %w", err)
	}
	return parseSearchResults(out), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	out, err := runCmd(ctx, "pkg_info", []string{pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("ports info %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("ports info: %w", err)
	}
	p := parseInfoOutput(out, pkg)
	if p == nil {
		return nil, fmt.Errorf("ports info %s: %w", pkg, snack.ErrNotFound)
	}
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	c := exec.CommandContext(ctx, "pkg_info", pkg)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return false, nil
		}
		return false, fmt.Errorf("ports isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	out, err := runCmd(ctx, "pkg_info", []string{pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("ports version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("ports version: %w", err)
	}
	p := parseInfoOutput(out, pkg)
	if p == nil || p.Version == "" {
		return "", fmt.Errorf("ports version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return p.Version, nil
}
