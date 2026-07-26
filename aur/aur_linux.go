//go:build linux

package aur

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/gogrlx/snack"
)

const aurGitBase = "https://aur.archlinux.org"

func available() bool {
	for _, tool := range []string{"makepkg", "pacman"} {
		if _, err := exec.LookPath(tool); err != nil {
			return false
		}
	}
	return true
}

// runPacman executes a pacman command and returns stdout.
func runPacman(ctx context.Context, args []string, sudo bool) (string, error) {
	cmd := "pacman"
	if sudo {
		args = append([]string{cmd}, args...)
		cmd = "sudo"
	}
	c := exec.CommandContext(ctx, cmd, args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") || strings.Contains(se, "requires root") {
			return "", fmt.Errorf("aur: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("aur: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
}

// install clones PKGBUILDs from the AUR, builds with makepkg, and installs with pacman.
func (a *AUR) install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	o := snack.ApplyOptions(opts...)

	var installed []snack.Package
	var unchanged []string

	for _, t := range pkgs {
		if !o.Reinstall && !o.DryRun {
			ok, err := isInstalled(ctx, t.Name)
			if err != nil {
				return snack.InstallResult{}, err
			}
			if ok && t.Version == "" {
				unchanged = append(unchanged, t.Name)
				continue
			}
		}

		pkgFile, cleanupDir, err := a.buildPackage(ctx, t)
		if err != nil {
			return snack.InstallResult{}, fmt.Errorf("aur install %s: %w", t.Name, err)
		}

		if o.DryRun {
			if cleanupDir != "" {
				os.RemoveAll(cleanupDir)
			}
			installed = append(installed, snack.Package{Name: t.Name, Repository: "aur"})
			continue
		}

		args := []string{"-U", "--noconfirm", pkgFile}
		installErr := func() error {
			if cleanupDir != "" {
				defer os.RemoveAll(cleanupDir)
			}
			_, err := runPacman(ctx, args, o.Sudo)
			return err
		}()
		if installErr != nil {
			return snack.InstallResult{}, fmt.Errorf("aur install %s: %w", t.Name, installErr)
		}

		v, _ := version(ctx, t.Name)
		installed = append(installed, snack.Package{
			Name:       t.Name,
			Version:    v,
			Repository: "aur",
			Installed:  true,
		})
	}

	return snack.InstallResult{Installed: installed, Unchanged: unchanged}, nil
}

// buildPackage clones the AUR git repo for a package and runs makepkg.
// Returns the path to the built .pkg.tar.zst file and an optional cleanup
// directory (non-empty only when a temp dir was created; caller must remove it).
func (a *AUR) buildPackage(ctx context.Context, t snack.Target) (pkgPath string, cleanupDir string, err error) {
	// Determine build directory
	buildDir := a.BuildDir
	if buildDir == "" {
		tmp, err := os.MkdirTemp("", "snack-aur-*")
		if err != nil {
			return "", "", fmt.Errorf("creating temp dir: %w", err)
		}
		buildDir = tmp
		cleanupDir = tmp
	}

	pkgDir := filepath.Join(buildDir, t.Name)

	// Clone or update the PKGBUILD repo
	if err := cloneOrPull(ctx, t.Name, pkgDir); err != nil {
		return "", cleanupDir, err
	}

	// Run makepkg
	args := []string{"-s", "-f", "--noconfirm"}
	args = append(args, a.MakepkgFlags...)
	c := exec.CommandContext(ctx, "makepkg", args...)
	c.Dir = pkgDir
	var stderr bytes.Buffer
	c.Stderr = &stderr
	c.Stdout = &stderr // makepkg output goes to stderr anyway
	if err := c.Run(); err != nil {
		return "", cleanupDir, fmt.Errorf("makepkg %s: %s: %w", t.Name, strings.TrimSpace(stderr.String()), err)
	}

	// Find the built package file
	matches, err := filepath.Glob(filepath.Join(pkgDir, "*.pkg.tar*"))
	if err != nil || len(matches) == 0 {
		return "", cleanupDir, fmt.Errorf("makepkg %s: no package file produced", t.Name)
	}
	return matches[len(matches)-1], cleanupDir, nil
}

// cloneOrPull clones the AUR git repo if it doesn't exist, or pulls if it does.
func cloneOrPull(ctx context.Context, pkg, dir string) error {
	repoURL := aurGitBase + "/" + pkg + ".git"

	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		// Repo exists, pull latest
		r, err := git.PlainOpen(dir)
		if err != nil {
			return fmt.Errorf("aur open %s: %w", pkg, err)
		}
		w, err := r.Worktree()
		if err != nil {
			return fmt.Errorf("aur worktree %s: %w", pkg, err)
		}
		if err := w.Pull(&git.PullOptions{}); err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("aur pull %s: %w", pkg, err)
		}
		return nil
	}

	// Clone fresh (depth 1)
	_, err := git.PlainCloneContext(ctx, dir, false, &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
	})
	if err != nil {
		if err == transport.ErrRepositoryNotFound {
			return fmt.Errorf("aur clone %s: %w", pkg, snack.ErrNotFound)
		}
		return fmt.Errorf("aur clone %s: %w", pkg, err)
	}
	return nil
}

func remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	o := snack.ApplyOptions(opts...)

	var toRemove []snack.Target
	var unchanged []string
	for _, t := range pkgs {
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
		if _, err := runPacman(ctx, args, o.Sudo); err != nil {
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
	_, err := runPacman(ctx, args, o.Sudo)
	return err
}

// upgradeAll rebuilds all installed foreign packages that have newer versions in the AUR.
func (a *AUR) upgradeAll(ctx context.Context, opts ...snack.Option) error {
	upgrades, err := listUpgrades(ctx)
	if err != nil {
		return err
	}
	if len(upgrades) == 0 {
		return nil
	}

	targets := make([]snack.Target, len(upgrades))
	for i, p := range upgrades {
		targets[i] = snack.Target{Name: p.Name}
	}

	// Force reinstall since we're upgrading
	allOpts := append([]snack.Option{snack.WithReinstall()}, opts...)
	_, err = a.install(ctx, targets, allOpts...)
	return err
}

func list(ctx context.Context) ([]snack.Package, error) {
	// pacman -Qm lists foreign (non-repo) packages, which are typically AUR
	out, err := runPacman(ctx, []string{"-Qm"}, false)
	if err != nil {
		// exit status 1 means no foreign packages
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("aur list: %w", err)
	}
	return parsePackageList(out), nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	c := exec.CommandContext(ctx, "pacman", "-Q", pkg)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("aur isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	out, err := runPacman(ctx, []string{"-Q", pkg}, false)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("aur version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("aur version: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) < 2 {
		return "", fmt.Errorf("aur version %s: unexpected output %q", pkg, out)
	}
	return parts[1], nil
}

func latestVersion(ctx context.Context, pkg string) (string, error) {
	p, err := rpcInfo(ctx, pkg)
	if err != nil {
		return "", err
	}
	return p.Version, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	// Get all installed foreign packages
	installed, err := list(ctx)
	if err != nil {
		return nil, err
	}
	if len(installed) == 0 {
		return nil, nil
	}

	// Batch-query the AUR for all of them
	names := make([]string, len(installed))
	for i, p := range installed {
		names[i] = p.Name
	}
	aurInfo, err := rpcInfoMulti(ctx, names)
	if err != nil {
		return nil, err
	}

	// Compare versions
	var upgrades []snack.Package
	for _, inst := range installed {
		aurPkg, ok := aurInfo[inst.Name]
		if !ok {
			continue // not in AUR (maybe from a custom repo)
		}
		cmp, err := versionCmp(ctx, inst.Version, aurPkg.Version)
		if err != nil {
			continue // skip packages where vercmp fails
		}
		if cmp < 0 {
			upgrades = append(upgrades, snack.Package{
				Name:       inst.Name,
				Version:    aurPkg.Version,
				Repository: "aur",
				Installed:  true,
			})
		}
	}
	return upgrades, nil
}

func upgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	inst, err := version(ctx, pkg)
	if err != nil {
		return false, err
	}
	latest, err := latestVersion(ctx, pkg)
	if err != nil {
		return false, err
	}
	cmp, err := versionCmp(ctx, inst, latest)
	if err != nil {
		return false, err
	}
	return cmp < 0, nil
}

func versionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	c := exec.CommandContext(ctx, "vercmp", ver1, ver2)
	out, err := c.Output()
	if err != nil {
		return 0, fmt.Errorf("vercmp: %w", err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("vercmp: unexpected output %q: %w", string(out), err)
	}
	switch {
	case n < 0:
		return -1, nil
	case n > 0:
		return 1, nil
	default:
		return 0, nil
	}
}

func autoremove(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)

	// Get orphans
	orphans, err := runPacman(ctx, []string{"-Qdtq"}, false)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil // no orphans
		}
		return fmt.Errorf("aur autoremove: %w", err)
	}
	orphans = strings.TrimSpace(orphans)
	if orphans == "" {
		return nil
	}

	pkgs := strings.Fields(orphans)
	args := append([]string{"-Rns", "--noconfirm"}, pkgs...)
	_, err = runPacman(ctx, args, o.Sudo)
	return err
}

// cleanBuildDir removes all subdirectories in the build directory.
func (a *AUR) cleanBuildDir() error {
	if a.BuildDir == "" {
		return nil // temp dirs are cleaned automatically
	}
	entries, err := os.ReadDir(a.BuildDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("aur clean: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			if err := os.RemoveAll(filepath.Join(a.BuildDir, e.Name())); err != nil {
				return fmt.Errorf("aur clean %s: %w", e.Name(), err)
			}
		}
	}
	return nil
}
