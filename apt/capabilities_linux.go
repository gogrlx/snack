//go:build linux

package apt

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gogrlx/snack"
)

// --- VersionQuerier ---

func latestVersion(ctx context.Context, pkg string) (string, error) {
	cmd := exec.CommandContext(ctx, "apt-cache", "policy", pkg)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("apt-cache policy %s: %w", pkg, snack.ErrNotFound)
	}
	candidate := parsePolicyCandidate(string(out))
	if candidate == "" {
		return "", fmt.Errorf("apt-cache policy %s: %w", pkg, snack.ErrNotFound)
	}
	return candidate, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	// Use apt-get --just-print upgrade instead of `apt list --upgradable`
	// because `apt` has unstable CLI output not intended for scripting.
	cmd := exec.CommandContext(ctx, "apt-get", "--just-print", "upgrade")
	cmd.Env = append(os.Environ(), "LANG=C", "DEBIAN_FRONTEND=noninteractive")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("apt-get --just-print upgrade: %w", err)
	}
	return parseUpgradeSimulation(string(out)), nil
}

func upgradeAvailable(ctx context.Context, pkg string) (bool, error) {
	cmd := exec.CommandContext(ctx, "apt-cache", "policy", pkg)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("apt-cache policy %s: %w", pkg, snack.ErrNotFound)
	}
	installed := parsePolicyInstalled(string(out))
	candidate := parsePolicyCandidate(string(out))
	if installed == "" {
		return false, fmt.Errorf("apt-cache policy %s: %w", pkg, snack.ErrNotInstalled)
	}
	if candidate == "" || candidate == installed {
		return false, nil
	}
	return true, nil
}

func versionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	// Check if ver1 < ver2
	cmd := exec.CommandContext(ctx, "dpkg", "--compare-versions", ver1, "lt", ver2)
	if err := cmd.Run(); err == nil {
		return -1, nil
	}
	// Check if ver1 = ver2
	cmd = exec.CommandContext(ctx, "dpkg", "--compare-versions", ver1, "eq", ver2)
	if err := cmd.Run(); err == nil {
		return 0, nil
	}
	// Must be ver1 > ver2
	return 1, nil
}

// --- Holder ---

func hold(ctx context.Context, pkgs []string) error {
	args := append([]string{"hold"}, pkgs...)
	cmd := exec.CommandContext(ctx, "apt-mark", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt-mark hold: %w: %s", err, stderr.String())
	}
	return nil
}

func unhold(ctx context.Context, pkgs []string) error {
	args := append([]string{"unhold"}, pkgs...)
	cmd := exec.CommandContext(ctx, "apt-mark", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt-mark unhold: %w: %s", err, stderr.String())
	}
	return nil
}

func listHeld(ctx context.Context) ([]snack.Package, error) {
	cmd := exec.CommandContext(ctx, "apt-mark", "showhold")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("apt-mark showhold: %w", err)
	}
	return parseHoldList(string(out)), nil
}

func isHeld(ctx context.Context, pkg string) (bool, error) {
	cmd := exec.CommandContext(ctx, "apt-mark", "showhold", pkg)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("apt-mark showhold %s: %w", pkg, err)
	}
	return strings.TrimSpace(string(out)) == pkg, nil
}

// --- Cleaner ---

func autoremove(ctx context.Context, opts ...snack.Option) error {
	return runAptGet(ctx, "autoremove", nil, append(opts, snack.WithAssumeYes())...)
}

func clean(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "apt-get", "clean")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt-get clean: %w: %s", err, stderr.String())
	}
	return nil
}

// --- FileOwner ---

func fileList(ctx context.Context, pkg string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-L", pkg)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		errMsg := stderr.String()
		if strings.Contains(errMsg, "is not installed") || strings.Contains(errMsg, "not found") {
			return nil, fmt.Errorf("dpkg-query -L %s: %w", pkg, snack.ErrNotInstalled)
		}
		return nil, fmt.Errorf("dpkg-query -L %s: %w: %s", pkg, err, errMsg)
	}
	return parseFileList(string(out)), nil
}

func owner(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "dpkg", "-S", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("dpkg -S %s: %w", path, snack.ErrNotFound)
	}
	pkg := parseOwner(string(out))
	if pkg == "" {
		return "", fmt.Errorf("dpkg -S %s: unexpected output", path)
	}
	return pkg, nil
}

// --- RepoManager ---

func listRepos(_ context.Context) ([]snack.Repository, error) {
	var repos []snack.Repository

	files := []string{"/etc/apt/sources.list"}
	// Glob sources.list.d
	matches, _ := filepath.Glob("/etc/apt/sources.list.d/*.list")
	files = append(files, matches...)
	sourcesMatches, _ := filepath.Glob("/etc/apt/sources.list.d/*.sources")
	files = append(files, sourcesMatches...)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			if r := parseSourcesLine(scanner.Text()); r != nil {
				repos = append(repos, *r)
			}
		}
	}
	return repos, nil
}

func addRepo(ctx context.Context, repo snack.Repository) error {
	repoLine := repo.URL
	if repo.Type != "" {
		repoLine = repo.Type + " " + repo.URL
	}
	cmd := exec.CommandContext(ctx, "add-apt-repository", "-y", repoLine)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("add-apt-repository: %w: %s", err, stderr.String())
	}
	return nil
}

func removeRepo(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, "add-apt-repository", "--remove", "-y", id)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("add-apt-repository --remove: %w: %s", err, stderr.String())
	}
	return nil
}

// --- KeyManager ---

func addKey(ctx context.Context, key string) error {
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		// Modern approach: download and dearmor to /etc/apt/keyrings/
		name := filepath.Base(key)
		name = strings.TrimSuffix(name, filepath.Ext(name)) + ".gpg"
		dest := filepath.Join("/etc/apt/keyrings", name)

		// Ensure keyrings dir exists
		_ = os.MkdirAll("/etc/apt/keyrings", 0o755)

		// Download and dearmor
		shell := fmt.Sprintf("curl -fsSL %q | gpg --dearmor -o %q", key, dest)
		cmd := exec.CommandContext(ctx, "sh", "-c", shell)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			// Fallback to apt-key
			cmd2 := exec.CommandContext(ctx, "apt-key", "adv", "--fetch-keys", key)
			var stderr2 bytes.Buffer
			cmd2.Stderr = &stderr2
			if err2 := cmd2.Run(); err2 != nil {
				return fmt.Errorf("add key %s: %w: %s", key, err2, stderr2.String())
			}
		}
		return nil
	}
	// Treat as key ID or file path
	cmd := exec.CommandContext(ctx, "apt-key", "add", key)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt-key add: %w: %s", err, stderr.String())
	}
	return nil
}

func removeKey(ctx context.Context, keyID string) error {
	// Try removing from keyrings dir first
	matches, _ := filepath.Glob("/etc/apt/keyrings/*")
	for _, m := range matches {
		if strings.Contains(filepath.Base(m), keyID) {
			if err := os.Remove(m); err == nil {
				return nil
			}
		}
	}
	// Fallback to apt-key del
	cmd := exec.CommandContext(ctx, "apt-key", "del", keyID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt-key del %s: %w: %s", keyID, err, stderr.String())
	}
	return nil
}

func listKeys(ctx context.Context) ([]string, error) {
	var keys []string

	// List keyring files
	matches, _ := filepath.Glob("/etc/apt/keyrings/*.gpg")
	keys = append(keys, matches...)
	ascMatches, _ := filepath.Glob("/etc/apt/keyrings/*.asc")
	keys = append(keys, ascMatches...)

	// Also list apt-key entries
	cmd := exec.CommandContext(ctx, "apt-key", "list")
	out, _ := cmd.Output()
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		// Key fingerprint lines are hex strings
		if len(line) > 0 && !strings.Contains(line, "/") && !strings.Contains(line, "pub") && !strings.Contains(line, "uid") && !strings.Contains(line, "sub") && !strings.HasPrefix(line, "-") {
			keys = append(keys, line)
		}
	}
	return keys, nil
}

// NameNormalizer functions are in normalize.go (no build constraints needed).
