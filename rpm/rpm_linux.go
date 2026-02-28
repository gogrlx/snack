//go:build linux

package rpm

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("rpm")
	return err == nil
}

func run(ctx context.Context, args []string) (string, error) {
	c := exec.CommandContext(ctx, "rpm", args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") {
			return "", fmt.Errorf("rpm: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("rpm: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
}

func runWithSudo(ctx context.Context, args []string, sudo bool) (string, error) {
	cmd := "rpm"
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
		if strings.Contains(se, "permission denied") {
			return "", fmt.Errorf("rpm: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("rpm: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
}

// formatSources extracts source paths or names from targets for rpm -i/-U.
func formatSources(targets []snack.Target) []string {
	args := make([]string, 0, len(targets))
	for _, t := range targets {
		if t.Source != "" {
			args = append(args, t.Source)
		} else {
			args = append(args, t.Name)
		}
	}
	return args
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
		args := append([]string{"-i"}, formatSources(toInstall)...)
		if _, err := runWithSudo(ctx, args, o.Sudo); err != nil {
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
		args := append([]string{"-e"}, snack.TargetNames(toRemove)...)
		if _, err := runWithSudo(ctx, args, o.Sudo); err != nil {
			return snack.RemoveResult{}, err
		}
	}
	var removed []snack.Package
	for _, t := range toRemove {
		removed = append(removed, snack.Package{Name: t.Name})
	}
	return snack.RemoveResult{Removed: removed, Unchanged: unchanged}, nil
}

func upgradeAll(ctx context.Context, opts ...snack.Option) error {
	// rpm -U requires specific files; upgradeAll with no targets is a no-op
	_ = opts
	return fmt.Errorf("rpm: upgrade requires specific package files: %w", snack.ErrUnsupportedPlatform)
}

func list(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"-qa", "--queryformat", `%{NAME}\t%{VERSION}-%{RELEASE}\t%{SUMMARY}\n`})
	if err != nil {
		return nil, fmt.Errorf("rpm list: %w", err)
	}
	return parseList(out), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	out, err := run(ctx, []string{"-qa", "*" + query + "*", "--queryformat", `%{NAME}\t%{VERSION}-%{RELEASE}\t%{SUMMARY}\n`})
	if err != nil {
		// exit 1 = no matches
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("rpm search: %w", err)
	}
	return parseList(out), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	out, err := run(ctx, []string{"-qi", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "is not installed") ||
			strings.Contains(err.Error(), "exit status 1") {
			return nil, fmt.Errorf("rpm info %s: %w", pkg, snack.ErrNotInstalled)
		}
		return nil, fmt.Errorf("rpm info: %w", err)
	}
	p := parseInfo(out)
	if p == nil {
		return nil, fmt.Errorf("rpm info %s: %w", pkg, snack.ErrNotInstalled)
	}
	p.Installed = true
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	c := exec.CommandContext(ctx, "rpm", "-q", pkg)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("rpm isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	c := exec.CommandContext(ctx, "rpm", "-q", "--queryformat", "%{VERSION}-%{RELEASE}", pkg)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "is not installed") || strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("rpm version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("rpm version: %w", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func fileList(ctx context.Context, pkg string) ([]string, error) {
	out, err := run(ctx, []string{"-ql", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "is not installed") {
			return nil, fmt.Errorf("rpm fileList %s: %w", pkg, snack.ErrNotInstalled)
		}
		return nil, fmt.Errorf("rpm fileList: %w", err)
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "(contains no files)") {
			files = append(files, line)
		}
	}
	return files, nil
}

func owner(ctx context.Context, path string) (string, error) {
	out, err := run(ctx, []string{"-qf", path})
	if err != nil {
		if strings.Contains(err.Error(), "is not owned by any package") {
			return "", fmt.Errorf("rpm owner %s: %w", path, snack.ErrNotFound)
		}
		return "", fmt.Errorf("rpm owner: %w", err)
	}
	return strings.TrimSpace(out), nil
}
