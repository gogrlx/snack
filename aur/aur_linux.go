//go:build linux

package aur

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gogrlx/snack"
)

const aurRPC = "https://aur.archlinux.org/rpc/v5"

func available() bool {
	_, err := exec.LookPath("makepkg")
	return err == nil
}

// aurSearchResponse is the JSON response from the AUR RPC API.
type aurSearchResponse struct {
	ResultCount int `json:"resultcount"`
	Results     []struct {
		Name        string  `json:"Name"`
		Version     string  `json:"Version"`
		Description string  `json:"Description"`
		URL         string  `json:"URL"`
		OutOfDate   *int64  `json:"OutOfDate"`
		Maintainer  string  `json:"Maintainer"`
		Popularity  float64 `json:"Popularity"`
	} `json:"results"`
}

func aurQuery(ctx context.Context, queryType, arg string) (*aurSearchResponse, error) {
	url := fmt.Sprintf("%s/%s/%s", aurRPC, queryType, arg)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aur: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("aur: %w", err)
	}
	var result aurSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("aur: %w", err)
	}
	return &result, nil
}

func install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	o := snack.ApplyOptions(opts...)
	var installed []snack.Package
	var unchanged []string

	for _, t := range pkgs {
		// Check if already installed
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

		if err := installPkg(ctx, t.Name, o); err != nil {
			return snack.InstallResult{}, err
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

func installPkg(ctx context.Context, pkg string, opts snack.Options) error {
	tmpDir, err := os.MkdirTemp("", "aur-"+pkg)
	if err != nil {
		return fmt.Errorf("aur: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	gitURL := fmt.Sprintf("https://aur.archlinux.org/%s.git", pkg)
	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", gitURL, tmpDir)
	var stderr bytes.Buffer
	cloneCmd.Stderr = &stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("aur: git clone %s: %s: %w", pkg, stderr.String(), err)
	}

	makepkgArgs := []string{"-si", "--noconfirm"}
	if opts.AssumeYes {
		makepkgArgs = append(makepkgArgs, "--noconfirm")
	}

	makeCmd := exec.CommandContext(ctx, "makepkg", makepkgArgs...)
	makeCmd.Dir = tmpDir
	makeCmd.Stderr = &stderr
	makeCmd.Stdout = os.Stdout

	if err := makeCmd.Run(); err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") {
			return fmt.Errorf("aur: %w", snack.ErrPermissionDenied)
		}
		return fmt.Errorf("aur: makepkg %s: %s: %w", pkg, se, err)
	}
	return nil
}

func remove(_ context.Context, _ []snack.Target, _ ...snack.Option) (snack.RemoveResult, error) {
	return snack.RemoveResult{}, fmt.Errorf("aur: remove not supported, use pacman instead")
}

func purge(_ context.Context, _ []snack.Target, _ ...snack.Option) error {
	return fmt.Errorf("aur: purge not supported, use pacman instead")
}

func upgrade(ctx context.Context, opts ...snack.Option) error {
	aurPkgs, err := list(ctx)
	if err != nil {
		return err
	}
	for _, p := range aurPkgs {
		result, err := aurQuery(ctx, "info", p.Name)
		if err != nil {
			continue
		}
		if result.ResultCount == 0 {
			continue
		}
		if result.Results[0].Version != p.Version {
			if _, err := install(ctx, []snack.Target{{Name: p.Name}}, opts...); err != nil {
				return err
			}
		}
	}
	return nil
}

func update(_ context.Context) error {
	return nil
}

func list(ctx context.Context) ([]snack.Package, error) {
	c := exec.CommandContext(ctx, "pacman", "-Qm")
	var stdout bytes.Buffer
	c.Stdout = &stdout
	if err := c.Run(); err != nil {
		// exit status 1 means no foreign packages
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("aur list: %w", err)
	}
	return parsePackageList(stdout.String()), nil
}

func parsePackageList(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		pkgs = append(pkgs, snack.Package{
			Name:       parts[0],
			Version:    parts[1],
			Repository: "aur",
			Installed:  true,
		})
	}
	return pkgs
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	result, err := aurQuery(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var pkgs []snack.Package
	for _, r := range result.Results {
		pkgs = append(pkgs, snack.Package{
			Name:        r.Name,
			Version:     r.Version,
			Description: r.Description,
		})
	}
	return pkgs, nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	result, err := aurQuery(ctx, "info", pkg)
	if err != nil {
		return nil, err
	}
	if result.ResultCount == 0 {
		return nil, fmt.Errorf("aur info %s: %w", pkg, snack.ErrNotFound)
	}
	r := result.Results[0]
	return &snack.Package{
		Name:        r.Name,
		Version:     r.Version,
		Description: r.Description,
	}, nil
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
	aurPkgs, err := list(ctx)
	if err != nil {
		return false, err
	}
	for _, p := range aurPkgs {
		if p.Name == pkg {
			return true, nil
		}
	}
	return false, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	c := exec.CommandContext(ctx, "pacman", "-Q", pkg)
	var stdout bytes.Buffer
	c.Stdout = &stdout
	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", fmt.Errorf("aur version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("aur version: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(stdout.String()))
	if len(parts) < 2 {
		return "", fmt.Errorf("aur version %s: unexpected output", pkg)
	}
	return parts[1], nil
}

func latestVersion(ctx context.Context, pkg string) (string, error) {
	result, err := aurQuery(ctx, "info", pkg)
	if err != nil {
		return "", err
	}
	if result.ResultCount == 0 {
		return "", fmt.Errorf("aur latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return result.Results[0].Version, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	aurPkgs, err := list(ctx)
	if err != nil {
		return nil, err
	}
	var upgrades []snack.Package
	for _, p := range aurPkgs {
		result, err := aurQuery(ctx, "info", p.Name)
		if err != nil || result.ResultCount == 0 {
			continue
		}
		if result.Results[0].Version != p.Version {
			upgrades = append(upgrades, snack.Package{
				Name:       p.Name,
				Version:    result.Results[0].Version,
				Repository: "aur",
				Installed:  true,
			})
		}
	}
	return upgrades, nil
}

func upgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	installed, err := version(ctx, pkg)
	if err != nil {
		return false, err
	}
	latest, err := latestVersion(ctx, pkg)
	if err != nil {
		return false, err
	}
	cmp, err := versionCmp(ctx, installed, latest)
	if err != nil {
		return false, err
	}
	return cmp < 0, nil
}

func versionCmp(_ context.Context, ver1, ver2 string) (int, error) {
	c := exec.Command("vercmp", ver1, ver2)
	var stdout bytes.Buffer
	c.Stdout = &stdout
	if err := c.Run(); err != nil {
		return 0, fmt.Errorf("aur versionCmp: %w", err)
	}
	result := strings.TrimSpace(stdout.String())
	switch result {
	case "-1":
		return -1, nil
	case "0":
		return 0, nil
	case "1":
		return 1, nil
	default:
		return 0, fmt.Errorf("aur versionCmp: unexpected output %q", result)
	}
}

func autoremove(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)

	// Get orphans via pacman
	c := exec.CommandContext(ctx, "pacman", "-Qdtq")
	var stdout bytes.Buffer
	c.Stdout = &stdout
	if err := c.Run(); err != nil {
		// exit status 1 means no orphans
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("aur autoremove: %w", err)
	}

	orphans := strings.TrimSpace(stdout.String())
	if orphans == "" {
		return nil
	}

	args := []string{"-Rns", "--noconfirm"}
	args = append(args, strings.Fields(orphans)...)
	cmd := "pacman"
	if o.Sudo {
		args = append([]string{cmd}, args...)
		cmd = "sudo"
	}

	removeCmd := exec.CommandContext(ctx, cmd, args...)
	return removeCmd.Run()
}

func clean(_ context.Context) error {
	// AUR builds in temp dirs, nothing persistent to clean
	return nil
}

func upgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	// For AUR, upgrading is just reinstalling from source
	allOpts := append([]snack.Option{snack.WithReinstall()}, opts...)
	return install(ctx, pkgs, allOpts...)
}

// getAURBuildDir returns the directory to use for AUR builds.
func getAURBuildDir() string {
	if dir := os.Getenv("AURDEST"); dir != "" {
		return dir
	}
	if cache := os.Getenv("XDG_CACHE_HOME"); cache != "" {
		return filepath.Join(cache, "aur")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "aur")
}
