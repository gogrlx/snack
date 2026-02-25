//go:build linux

package flatpak

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("flatpak")
	return err == nil
}

func run(ctx context.Context, args []string) (string, error) {
	c := exec.CommandContext(ctx, "flatpak", args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") || strings.Contains(se, "requires root") {
			return "", fmt.Errorf("flatpak: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("flatpak: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
}

func install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	_ = snack.ApplyOptions(opts...)
	for _, t := range pkgs {
		remote := t.FromRepo
		if remote == "" {
			remote = "flathub"
		}
		args := []string{"install", "-y", remote, t.Name}
		if _, err := run(ctx, args); err != nil {
			return err
		}
	}
	return nil
}

func remove(ctx context.Context, pkgs []snack.Target, _ ...snack.Option) error {
	args := append([]string{"uninstall", "-y"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args)
	return err
}

func purge(ctx context.Context, pkgs []snack.Target, _ ...snack.Option) error {
	args := append([]string{"uninstall", "-y", "--delete-data"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args)
	return err
}

func upgrade(ctx context.Context, _ ...snack.Option) error {
	_, err := run(ctx, []string{"update", "-y"})
	return err
}

func list(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"list", "--columns=name,application,version,origin"})
	if err != nil {
		return nil, fmt.Errorf("flatpak list: %w", err)
	}
	return parseList(out), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	out, err := run(ctx, []string{"search", query, "--columns=name,application,version,remotes"})
	if err != nil {
		if strings.Contains(err.Error(), "No matches found") {
			return nil, nil
		}
		return nil, fmt.Errorf("flatpak search: %w", err)
	}
	return parseSearch(out), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	out, err := run(ctx, []string{"info", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("flatpak info %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("flatpak info: %w", err)
	}
	p := parseInfo(out)
	if p == nil {
		return nil, fmt.Errorf("flatpak info %s: %w", pkg, snack.ErrNotFound)
	}
	p.Installed = true
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	c := exec.CommandContext(ctx, "flatpak", "info", pkg)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("flatpak isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"info", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("flatpak version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("flatpak version: %w", err)
	}
	p := parseInfo(out)
	if p == nil || p.Version == "" {
		return "", fmt.Errorf("flatpak version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return p.Version, nil
}

func autoremove(ctx context.Context, _ ...snack.Option) error {
	_, err := run(ctx, []string{"uninstall", "--unused", "-y"})
	return err
}

func listRepos(ctx context.Context) ([]snack.Repository, error) {
	out, err := run(ctx, []string{"remotes", "--columns=name,url,options"})
	if err != nil {
		return nil, fmt.Errorf("flatpak listRepos: %w", err)
	}
	return parseRemotes(out), nil
}

func addRepo(ctx context.Context, repo snack.Repository) error {
	_, err := run(ctx, []string{"remote-add", "--if-not-exists", repo.Name, repo.URL})
	return err
}

func removeRepo(ctx context.Context, id string) error {
	_, err := run(ctx, []string{"remote-delete", id})
	return err
}
