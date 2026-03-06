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
		remote := t.FromRepo
		if remote == "" {
			remote = "flathub"
		}
		args := []string{"install", "-y", remote, t.Name}
		if _, err := run(ctx, args); err != nil {
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
		args := append([]string{"uninstall", "-y"}, snack.TargetNames(toRemove)...)
		if _, err := run(ctx, args); err != nil {
			return snack.RemoveResult{}, err
		}
	}
	var removed []snack.Package
	for _, t := range toRemove {
		removed = append(removed, snack.Package{Name: t.Name})
	}
	return snack.RemoveResult{Removed: removed, Unchanged: unchanged}, nil
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

func latestVersion(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"remote-info", "flathub", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("flatpak latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("flatpak latestVersion: %w", err)
	}
	p := parseInfo(out)
	if p == nil || p.Version == "" {
		return "", fmt.Errorf("flatpak latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return p.Version, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"remote-ls", "--updates", "--columns=name,application,version,origin"})
	if err != nil {
		if strings.Contains(err.Error(), "No updates") {
			return nil, nil
		}
		return nil, fmt.Errorf("flatpak listUpgrades: %w", err)
	}
	return parseList(out), nil
}

func upgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	upgrades, err := listUpgrades(ctx)
	if err != nil {
		return false, err
	}
	for _, u := range upgrades {
		if u.Name == pkg || u.Description == pkg {
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
	if len(toUpgrade) > 0 {
		for _, t := range toUpgrade {
			args := []string{"update", "-y", t.Name}
			if _, err := run(ctx, args); err != nil {
				return snack.InstallResult{}, fmt.Errorf("flatpak update %s: %w", t.Name, err)
			}
		}
	}
	var upgraded []snack.Package
	for _, t := range toUpgrade {
		v, _ := version(ctx, t.Name)
		upgraded = append(upgraded, snack.Package{Name: t.Name, Version: v, Installed: true})
	}
	return snack.InstallResult{Installed: upgraded, Unchanged: unchanged}, nil
}
