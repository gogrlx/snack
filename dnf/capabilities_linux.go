//go:build linux

package dnf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gogrlx/snack"
)

func latestVersion(ctx context.Context, pkg string, v5 bool) (string, error) {
	// Try "dnf info <pkg>" which shows both installed and available
	out, err := run(ctx, []string{"info", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("dnf latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("dnf latestVersion: %w", err)
	}
	var p *snack.Package
	if v5 {
		p = parseInfoDNF5(out)
	} else {
		p = parseInfo(out)
	}
	if p == nil || p.Version == "" {
		return "", fmt.Errorf("dnf latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return p.Version, nil
}

func listUpgrades(ctx context.Context, v5 bool) ([]snack.Package, error) {
	args := []string{"list", "upgrades"}
	if v5 {
		args = []string{"list", "--upgrades"}
	}
	out, err := run(ctx, args, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("dnf listUpgrades: %w", err)
	}
	if v5 {
		return parseListDNF5(out), nil
	}
	return parseList(out), nil
}

func upgradeAvailable(ctx context.Context, pkg string, v5 bool) (bool, error) {
	args := []string{"list", "upgrades", pkg}
	if v5 {
		args = []string{"list", "--upgrades", pkg}
	}
	c := exec.CommandContext(ctx, "dnf", args...)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("dnf upgradeAvailable: %w", err)
	}
	return true, nil
}

func versionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	c := exec.CommandContext(ctx, "rpmdev-vercmp", ver1, ver2)
	var stdout bytes.Buffer
	c.Stdout = &stdout
	err := c.Run()
	out := strings.TrimSpace(stdout.String())
	// rpmdev-vercmp exits 0 for equal, 11 for ver1 > ver2, 12 for ver1 < ver2
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			switch exitErr.ExitCode() {
			case 11:
				return 1, nil
			case 12:
				return -1, nil
			}
		}
		return 0, fmt.Errorf("rpmdev-vercmp: %s: %w", out, err)
	}
	return 0, nil
}

func hold(ctx context.Context, pkgs []string) error {
	args := append([]string{"versionlock", "add"}, pkgs...)
	_, err := run(ctx, args, snack.Options{})
	return err
}

func unhold(ctx context.Context, pkgs []string) error {
	args := append([]string{"versionlock", "delete"}, pkgs...)
	_, err := run(ctx, args, snack.Options{})
	return err
}

func listHeld(ctx context.Context, v5 bool) ([]snack.Package, error) {
	out, err := run(ctx, []string{"versionlock", "list"}, snack.Options{})
	if err != nil {
		return nil, fmt.Errorf("dnf listHeld: %w", err)
	}
	if v5 {
		return parseVersionLockDNF5(out), nil
	}
	return parseVersionLock(out), nil
}

func autoremove(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	_, err := run(ctx, []string{"autoremove", "-y"}, o)
	return err
}

func clean(ctx context.Context) error {
	_, err := run(ctx, []string{"clean", "all"}, snack.Options{})
	return err
}

func fileList(ctx context.Context, pkg string) ([]string, error) {
	c := exec.CommandContext(ctx, "rpm", "-ql", pkg)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "is not installed") {
			return nil, fmt.Errorf("dnf fileList %s: %w", pkg, snack.ErrNotInstalled)
		}
		return nil, fmt.Errorf("rpm -ql: %s: %w", strings.TrimSpace(se), err)
	}
	var files []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "(contains no files)") {
			files = append(files, line)
		}
	}
	return files, nil
}

func owner(ctx context.Context, path string) (string, error) {
	c := exec.CommandContext(ctx, "rpm", "-qf", path)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "is not owned by any package") {
			return "", fmt.Errorf("dnf owner %s: %w", path, snack.ErrNotFound)
		}
		return "", fmt.Errorf("rpm -qf: %s: %w", strings.TrimSpace(se), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func listRepos(ctx context.Context, v5 bool) ([]snack.Repository, error) {
	out, err := run(ctx, []string{"repolist", "--all"}, snack.Options{})
	if err != nil {
		return nil, fmt.Errorf("dnf listRepos: %w", err)
	}
	if v5 {
		return parseRepoListDNF5(out), nil
	}
	return parseRepoList(out), nil
}

func addRepo(ctx context.Context, repo snack.Repository, v5 bool) error {
	var args []string
	if v5 {
		args = []string{"config-manager", "addrepo", "--from-repofile=" + repo.URL}
	} else {
		args = []string{"config-manager", "--add-repo", repo.URL}
	}
	_, err := run(ctx, args, snack.Options{})
	return err
}

func removeRepo(_ context.Context, id string) error {
	// Remove the repo file from /etc/yum.repos.d/
	repoFile := filepath.Join("/etc/yum.repos.d", id+".repo")
	if err := os.Remove(repoFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("dnf removeRepo %s: %w", id, snack.ErrNotFound)
		}
		return fmt.Errorf("dnf removeRepo %s: %w", id, err)
	}
	return nil
}

func addKey(ctx context.Context, key string) error {
	c := exec.CommandContext(ctx, "rpm", "--import", key)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("rpm --import: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	return nil
}

func removeKey(ctx context.Context, keyID string) error {
	c := exec.CommandContext(ctx, "rpm", "-e", keyID)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("rpm -e: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	return nil
}

func listKeys(ctx context.Context) ([]string, error) {
	c := exec.CommandContext(ctx, "rpm", "-qa", "gpg-pubkey*")
	var stdout bytes.Buffer
	c.Stdout = &stdout
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("rpm -qa gpg-pubkey: %w", err)
	}
	var keys []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			keys = append(keys, line)
		}
	}
	return keys, nil
}

func groupList(ctx context.Context, v5 bool) ([]string, error) {
	out, err := run(ctx, []string{"group", "list"}, snack.Options{})
	if err != nil {
		return nil, fmt.Errorf("dnf groupList: %w", err)
	}
	if v5 {
		return parseGroupListDNF5(out), nil
	}
	return parseGroupList(out), nil
}

func groupInfo(ctx context.Context, group string, v5 bool) ([]snack.Package, error) {
	out, err := run(ctx, []string{"group", "info", group}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("dnf groupInfo %s: %w", group, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("dnf groupInfo: %w", err)
	}
	if v5 {
		return parseGroupInfoDNF5(out), nil
	}
	return parseGroupInfo(out), nil
}

func groupInstall(ctx context.Context, group string, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	_, err := run(ctx, []string{"group", "install", "-y", group}, o)
	return err
}
