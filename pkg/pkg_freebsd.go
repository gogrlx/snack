//go:build freebsd

package pkg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func available() bool {
	_, err := exec.LookPath("pkg")
	return err == nil
}

func run(ctx context.Context, args []string, opts snack.Options) (string, error) {
	cmdName := "pkg"
	cmdArgs := make([]string, 0, len(args)+2)
	cmdArgs = append(cmdArgs, args...)

	if opts.Sudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmdName = "sudo"
	}

	c := exec.CommandContext(ctx, cmdName, cmdArgs...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		se := stderr.String()
		if strings.Contains(se, "permission denied") || strings.Contains(se, "Insufficient privileges") {
			return "", fmt.Errorf("pkg: %w", snack.ErrPermissionDenied)
		}
		return "", fmt.Errorf("pkg: %s: %w", strings.TrimSpace(se), err)
	}
	return stdout.String(), nil
}

func formatTargets(targets []snack.Target) []string {
	args := make([]string, 0, len(targets))
	for _, t := range targets {
		if t.Version != "" {
			args = append(args, t.Name+"-"+t.Version)
		} else {
			args = append(args, t.Name)
		}
	}
	return args
}

func install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	args := append([]string{"install", "-y"}, formatTargets(pkgs)...)
	_, err := run(ctx, args, o)
	return err
}

func remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	args := append([]string{"delete", "-y"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args, o)
	return err
}

func purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	args := append([]string{"delete", "-y", "-f"}, snack.TargetNames(pkgs)...)
	_, err := run(ctx, args, o)
	return err
}

func upgrade(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	_, err := run(ctx, []string{"upgrade", "-y"}, o)
	return err
}

func update(ctx context.Context) error {
	_, err := run(ctx, []string{"update"}, snack.Options{})
	return err
}

func list(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"query", "%n\t%v\t%c"}, snack.Options{})
	if err != nil {
		return nil, fmt.Errorf("pkg list: %w", err)
	}
	return parseQuery(out), nil
}

func search(ctx context.Context, query string) ([]snack.Package, error) {
	out, err := run(ctx, []string{"search", query}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("pkg search: %w", err)
	}
	return parseSearch(out), nil
}

func info(ctx context.Context, pkg string) (*snack.Package, error) {
	out, err := run(ctx, []string{"info", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") || strings.Contains(err.Error(), "exit status 70") {
			return nil, fmt.Errorf("pkg info %s: %w", pkg, snack.ErrNotFound)
		}
		return nil, fmt.Errorf("pkg info: %w", err)
	}
	p := parseInfo(out)
	if p == nil {
		return nil, fmt.Errorf("pkg info %s: %w", pkg, snack.ErrNotFound)
	}
	return p, nil
}

func isInstalled(ctx context.Context, pkg string) (bool, error) {
	c := exec.CommandContext(ctx, "pkg", "info", pkg)
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return false, nil
		}
		return false, fmt.Errorf("pkg isInstalled: %w", err)
	}
	return true, nil
}

func version(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"query", "%v", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") || strings.Contains(err.Error(), "exit status 70") {
			return "", fmt.Errorf("pkg version %s: %w", pkg, snack.ErrNotInstalled)
		}
		return "", fmt.Errorf("pkg version: %w", err)
	}
	v := strings.TrimSpace(out)
	if v == "" {
		return "", fmt.Errorf("pkg version %s: %w", pkg, snack.ErrNotInstalled)
	}
	return v, nil
}
