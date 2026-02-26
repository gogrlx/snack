//go:build linux

package snap

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	if _, err := exec.LookPath("snap"); err != nil {
		return false
	}
	// Verify snapd is running by checking snap version (requires daemon).
	cmd := exec.Command("snap", "version")
	return cmd.Run() == nil
}

func run(ctx context.Context, args []string) (string, error) {
	c := exec.CommandContext(ctx, "snap", args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") || strings.Contains(se, "requires root") {
			return "", fmt.Errorf("snap: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("snap: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
}

func install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	_ = snack.ApplyOptions(opts...)
	for _, t := range pkgs {
		args := []string{"install"}
		// Handle --classic or --channel via FromRepo
		if t.FromRepo != "" {
			if t.FromRepo == "classic" {
				args = append(args, "--classic")
			} else {
				args = append(args, "--channel="+t.FromRepo)
			}
		}
		args = append(args, t.Name)
		if _, err := run(ctx, args); err != nil {
			return err
		}
	}
	return nil
}

func remove(ctx context.Context, pkgs []snack.Target, _ ...snack.Option) error {
	args := append([]string{"remove"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args)
	return err
}

func purge(ctx context.Context, pkgs []snack.Target, _ ...snack.Option) error {
	args := append([]string{"remove", "--purge"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args)
	return err
}

func upgrade(ctx context.Context, _ ...snack.Option) error {
	_, err := run(ctx, []string{"refresh"})
	return err
}

func update(ctx context.Context) error {
	// snap auto-refreshes; just check for available updates
	_, _ = run(ctx, []string{"refresh", "--list"})
	return nil
}

func list(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"list"})
	if err != nil {
		return nil, fmt.Errorf("snap list: %w", err)
	}
	return parseSnapList(out), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	out, err := run(ctx, []string{"find", query})
	if err != nil {
		if strings.Contains(err.Error(), "No matching snaps") {
			return nil, nil
		}
		return nil, fmt.Errorf("snap search: %w", err)
	}
	return parseSnapFind(out), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	out, err := run(ctx, []string{"info", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("snap info %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("snap info: %w", err)
	}
	p := parseSnapInfo(out)
	if p == nil {
		return nil, fmt.Errorf("snap info %s: %w", pkg, snack.ErrNotFound)
	}
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	c := exec.CommandContext(ctx, "snap", "list", pkg)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("snap isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"list", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("snap version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("snap version: %w", err)
	}
	pkgs := parseSnapList(out)
	if len(pkgs) == 0 {
		return "", fmt.Errorf("snap version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return pkgs[0].Version, nil
}

func latestVersion(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"info", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("snap latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("snap latestVersion: %w", err)
	}
	ver := parseSnapInfoVersion(out)
	if ver == "" {
		return "", fmt.Errorf("snap latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return ver, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"refresh", "--list"})
	if err != nil {
		// "All snaps up to date" exits with code 0 on some versions, 1 on others
		if strings.Contains(err.Error(), "All snaps up to date") {
			return nil, nil
		}
		return nil, fmt.Errorf("snap listUpgrades: %w", err)
	}
	return parseSnapRefreshList(out), nil
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
	return semverCmp(ver1, ver2), nil
}
