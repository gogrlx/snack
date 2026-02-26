//go:build linux

package apk

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("apk")
	return err == nil
}

func buildArgs(base []string, opts snack.Options) (string, []string) {
	cmd := "apk"
	var args []string

	if opts.Root != "" {
		args = append(args, "--root", opts.Root)
	}
	if opts.DryRun {
		args = append(args, "--simulate")
	}

	args = append(args, base...)

	if opts.Sudo {
		args = append([]string{cmd}, args...)
		cmd = "sudo"
	}

	return cmd, args
}

func run(ctx context.Context, base []string, opts ...snack.Option) (string, error) {
	o := snack.ApplyOptions(opts...)
	cmd, args := buildArgs(base, o)
	c := exec.CommandContext(ctx, cmd, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(out))
		if strings.Contains(outStr, "permission denied") || strings.Contains(outStr, "Permission denied") {
			return outStr, fmt.Errorf("%s: %w", outStr, snack.ErrPermissionDenied)
		}
		return outStr, fmt.Errorf("apk: %s: %w", outStr, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// formatTargets converts targets to apk CLI arguments.
// apk uses "pkg=version" for version pinning.
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

func install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	args := append([]string{"add"}, formatTargets(pkgs)...)
	_, err := run(ctx, args, opts...)
	return err
}

func remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	args := append([]string{"del"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args, opts...)
	return err
}

func purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	args := append([]string{"del", "--purge"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args, opts...)
	return err
}

func upgrade(ctx context.Context, opts ...snack.Option) error {
	_, err := run(ctx, []string{"upgrade"}, opts...)
	return err
}

func update(ctx context.Context) error {
	_, err := run(ctx, []string{"update"})
	return err
}

func list(ctx context.Context) ([]snack.Package, error) {
	cmd := exec.CommandContext(ctx, "apk", "list", "--installed")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("apk list: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	return parseListInstalled(string(out)), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	cmd := exec.CommandContext(ctx, "apk", "search", "-v", query)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("apk search: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	results := parseSearch(string(out))
	if len(results) == 0 {
		return nil, fmt.Errorf("apk search %q: %w", query, snack.ErrNotFound)
	}
	return results, nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	cmd := exec.CommandContext(ctx, "apk", "info", "-a", pkg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("apk info %q: %w", pkg, snack.ErrNotFound)
	}
	p := parseInfo(string(out))
	if p == nil {
		return nil, fmt.Errorf("apk info %q: %w", pkg, snack.ErrNotFound)
	}
	name, ver := parseInfoNameVersion(string(out))
	if name == "" {
		name = pkg
	}
	p.Name = name
	p.Version = ver
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	cmd := exec.CommandContext(ctx, "apk", "info", "-e", pkg)
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("apk info -e %q: %w", pkg, err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	cmd := exec.CommandContext(ctx, "apk", "info", pkg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("apk info %q: %w", pkg, snack.ErrNotInstalled)
	}
	_, ver := parseInfoNameVersion(string(out))
	if ver == "" {
		return "", fmt.Errorf("apk version %q: %w", pkg, snack.ErrNotInstalled)
	}
	return ver, nil
}
