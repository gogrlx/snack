//go:build darwin || linux

package brew

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

func run(ctx context.Context, args []string) (string, error) {
	c := exec.CommandContext(ctx, "brew", args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") {
			return "", fmt.Errorf("brew: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("brew: %s: %w", strings.TrimSpace(se), err)
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
		args := []string{"install"}
		pkg := t.Name
		if t.Version != "" {
			pkg = t.Name + "@" + t.Version
		}
		args = append(args, pkg)
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
		args := append([]string{"uninstall"}, snack.TargetNames(toRemove)...)
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
	args := append([]string{"uninstall", "--zap"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args)
	return err
}

func upgrade(ctx context.Context, _ ...snack.Option) error {
	_, err := run(ctx, []string{"upgrade"})
	return err
}

func update(ctx context.Context) error {
	_, err := run(ctx, []string{"update"})
	return err
}

func list(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"list", "--versions"})
	if err != nil {
		return nil, fmt.Errorf("brew list: %w", err)
	}
	return parseBrewList(out), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	out, err := run(ctx, []string{"search", query})
	if err != nil {
		if strings.Contains(err.Error(), "No formulae or casks found") {
			return nil, nil
		}
		return nil, fmt.Errorf("brew search: %w", err)
	}
	return parseBrewSearch(out), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	out, err := run(ctx, []string{"info", "--json=v2", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "No available formula") ||
			strings.Contains(err.Error(), "No formulae or casks found") {
			return nil, fmt.Errorf("brew info %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("brew info: %w", err)
	}
	p := parseBrewInfo(out)
	if p == nil {
		return nil, fmt.Errorf("brew info %s: %w", pkg, snack.ErrNotFound)
	}
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	c := exec.CommandContext(ctx, "brew", "list", pkg)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("brew isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"list", "--versions", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("brew version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("brew version: %w", err)
	}
	pkgs := parseBrewList(out)
	if len(pkgs) == 0 {
		return "", fmt.Errorf("brew version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return pkgs[0].Version, nil
}

func latestVersion(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"info", "--json=v2", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "No available formula") {
			return "", fmt.Errorf("brew latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("brew latestVersion: %w", err)
	}
	ver := parseBrewInfoVersion(out)
	if ver == "" {
		return "", fmt.Errorf("brew latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return ver, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"outdated", "--json=v2"})
	if err != nil {
		return nil, fmt.Errorf("brew listUpgrades: %w", err)
	}
	return parseBrewOutdated(out), nil
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

func autoremove(ctx context.Context, _ ...snack.Option) error {
	_, err := run(ctx, []string{"autoremove"})
	return err
}

func clean(ctx context.Context) error {
	_, err := run(ctx, []string{"cleanup"})
	return err
}

// brewInfoJSON represents the JSON output from `brew info --json=v2`.
type brewInfoJSON struct {
	Formulae []struct {
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		Desc        string `json:"desc"`
		Versions    struct {
			Stable string `json:"stable"`
		} `json:"versions"`
		Installed []struct {
			Version string `json:"version"`
		} `json:"installed"`
	} `json:"formulae"`
	Casks []struct {
		Token   string `json:"token"`
		Name    []string `json:"name"`
		Desc    string `json:"desc"`
		Version string `json:"version"`
	} `json:"casks"`
}

// brewOutdatedJSON represents the JSON output from `brew outdated --json=v2`.
type brewOutdatedJSON struct {
	Formulae []struct {
		Name             string `json:"name"`
		InstalledVersions []string `json:"installed_versions"`
		CurrentVersion   string `json:"current_version"`
	} `json:"formulae"`
	Casks []struct {
		Name             string `json:"name"`
		InstalledVersions string `json:"installed_versions"`
		CurrentVersion   string `json:"current_version"`
	} `json:"casks"`
}

func parseBrewList(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		pkg := snack.Package{
			Name:      fields[0],
			Installed: true,
		}
		if len(fields) >= 2 {
			pkg.Version = fields[1]
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func parseBrewSearch(output string) []snack.Package {
	var pkgs []snack.Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "==>") {
			continue
		}
		for _, name := range strings.Fields(line) {
			pkgs = append(pkgs, snack.Package{Name: name})
		}
	}
	return pkgs
}

func parseBrewInfo(output string) *snack.Package {
	var data brewInfoJSON
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return nil
	}
	if len(data.Formulae) > 0 {
		f := data.Formulae[0]
		pkg := &snack.Package{
			Name:        f.Name,
			Version:     f.Versions.Stable,
			Description: f.Desc,
		}
		if len(f.Installed) > 0 {
			pkg.Installed = true
			pkg.Version = f.Installed[0].Version
		}
		return pkg
	}
	if len(data.Casks) > 0 {
		c := data.Casks[0]
		return &snack.Package{
			Name:        c.Token,
			Version:     c.Version,
			Description: c.Desc,
		}
	}
	return nil
}

func parseBrewInfoVersion(output string) string {
	var data brewInfoJSON
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return ""
	}
	if len(data.Formulae) > 0 {
		return data.Formulae[0].Versions.Stable
	}
	if len(data.Casks) > 0 {
		return data.Casks[0].Version
	}
	return ""
}

func parseBrewOutdated(output string) []snack.Package {
	var data brewOutdatedJSON
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return nil
	}
	var pkgs []snack.Package
	for _, f := range data.Formulae {
		pkgs = append(pkgs, snack.Package{
			Name:      f.Name,
			Version:   f.CurrentVersion,
			Installed: true,
		})
	}
	for _, c := range data.Casks {
		pkgs = append(pkgs, snack.Package{
			Name:      c.Name,
			Version:   c.CurrentVersion,
			Installed: true,
		})
	}
	return pkgs
}

func semverCmp(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var numA, numB int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &numA)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &numB)
		}
		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
	}
	return 0
}

func fileList(ctx context.Context, pkg string) ([]string, error) {
	out, err := run(ctx, []string{"list", "--formula", pkg})
	if err != nil {
		// Try cask
		out, err = run(ctx, []string{"list", "--cask", pkg})
		if err != nil {
			if strings.Contains(err.Error(), "exit status 1") {
				return nil, fmt.Errorf("brew fileList %s: %w", pkg, snack.ErrNotInstalled)
			}
			return nil, fmt.Errorf("brew fileList: %w", err)
		}
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func owner(ctx context.Context, path string) (string, error) {
	// brew doesn't have a direct "which package owns this file" command
	// We need to iterate through installed packages and check their files
	out, err := run(ctx, []string{"list", "--formula"})
	if err != nil {
		return "", fmt.Errorf("brew owner: %w", err)
	}
	for _, pkg := range strings.Fields(out) {
		files, err := fileList(ctx, pkg)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f == path || strings.HasSuffix(f, "/"+path) {
				return pkg, nil
			}
		}
	}
	return "", fmt.Errorf("brew owner %s: %w", path, snack.ErrNotFound)
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
	for _, t := range toUpgrade {
		if _, err := run(ctx, []string{"upgrade", t.Name}); err != nil {
			return snack.InstallResult{}, fmt.Errorf("brew upgrade %s: %w", t.Name, err)
		}
	}
	var upgraded []snack.Package
	for _, t := range toUpgrade {
		v, _ := version(ctx, t.Name)
		upgraded = append(upgraded, snack.Package{Name: t.Name, Version: v, Installed: true})
	}
	return snack.InstallResult{Installed: upgraded, Unchanged: unchanged}, nil
}
