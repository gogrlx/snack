//go:build linux

package dpkg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("dpkg")
	return err == nil
}

func install(ctx context.Context, pkgs []string, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	var args []string
	if o.Sudo {
		args = append(args, "sudo")
	}
	args = append(args, "dpkg", "-i")
	if o.DryRun {
		args = append(args, "--simulate")
	}
	args = append(args, pkgs...)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if strings.Contains(errMsg, "Permission denied") || strings.Contains(errMsg, "are you root") {
			return fmt.Errorf("dpkg -i: %w", snack.ErrPermissionDenied)
		}
		return fmt.Errorf("dpkg -i: %w: %s", err, errMsg)
	}
	return nil
}

func remove(ctx context.Context, pkgs []string, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	var args []string
	if o.Sudo {
		args = append(args, "sudo")
	}
	args = append(args, "dpkg", "-r")
	if o.DryRun {
		args = append(args, "--simulate")
	}
	args = append(args, pkgs...)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dpkg -r: %w: %s", err, stderr.String())
	}
	return nil
}

func purge(ctx context.Context, pkgs []string, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	var args []string
	if o.Sudo {
		args = append(args, "sudo")
	}
	args = append(args, "dpkg", "-P")
	if o.DryRun {
		args = append(args, "--simulate")
	}
	args = append(args, pkgs...)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dpkg -P: %w: %s", err, stderr.String())
	}
	return nil
}

func list(ctx context.Context) ([]snack.Package, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Package}\\t${Version}\\t${Status}\\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dpkg-query list: %w", err)
	}
	return parseList(string(out)), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	pattern := fmt.Sprintf("*%s*", query)
	cmd := exec.CommandContext(ctx, "dpkg-query", "-l", pattern)
	out, err := cmd.Output()
	if err != nil {
		// dpkg-query -l returns exit 1 when no packages match
		return nil, nil
	}
	return parseDpkgList(string(out)), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-s", pkg)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(stderr.String(), "is not installed") || strings.Contains(stderr.String(), "not found") {
			return nil, fmt.Errorf("dpkg-query -s %s: %w", pkg, snack.ErrNotInstalled)
		}
		return nil, fmt.Errorf("dpkg-query -s %s: %w", pkg, err)
	}
	return parseInfo(string(out))
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Status}", pkg)
	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "install ok installed", nil
}

func version(ctx context.Context, pkg string) (string, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Version}", pkg)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("version %s: %w", pkg, snack.ErrNotInstalled)
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return "", fmt.Errorf("version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return v, nil
}
