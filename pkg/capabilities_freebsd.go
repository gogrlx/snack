//go:build freebsd

package pkg

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func latestVersion(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"rquery", "%v", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") || strings.Contains(err.Error(), "exit status 70") {
			return "", fmt.Errorf("pkg latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("pkg latestVersion: %w", err)
	}
	v := strings.TrimSpace(out)
	if v == "" {
		return "", fmt.Errorf("pkg latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return v, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"upgrade", "-n"}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		return nil, fmt.Errorf("pkg listUpgrades: %w", err)
	}
	return parseUpgrades(out), nil
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

func versionCmp(ctx context.Context, ver1, ver2 string) (int, error) {
	c := exec.CommandContext(ctx, "pkg", "version", "-t", ver1, ver2)
	out, err := c.Output()
	if err != nil {
		return 0, fmt.Errorf("pkg version -t: %w", err)
	}
	switch strings.TrimSpace(string(out)) {
	case "<":
		return -1, nil
	case ">":
		return 1, nil
	case "=":
		return 0, nil
	default:
		return 0, fmt.Errorf("pkg version -t: unexpected output %q", string(out))
	}
}

func autoremove(ctx context.Context, opts ...snack.Option) error {
	o := snack.ApplyOptions(opts...)
	_, err := run(ctx, []string{"autoremove", "-y"}, o)
	return err
}

func clean(ctx context.Context) error {
	_, err := run(ctx, []string{"clean", "-y"}, snack.Options{})
	return err
}

func fileList(ctx context.Context, pkg string) ([]string, error) {
	out, err := run(ctx, []string{"info", "-l", pkg}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") || strings.Contains(err.Error(), "exit status 70") {
			return nil, fmt.Errorf("pkg fileList %s: %w", pkg, snack.ErrNotInstalled)
		}
		return nil, fmt.Errorf("pkg fileList: %w", err)
	}
	return parseFileList(out), nil
}

func owner(ctx context.Context, path string) (string, error) {
	out, err := run(ctx, []string{"which", path}, snack.Options{})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") || strings.Contains(err.Error(), "was not found") {
			return "", fmt.Errorf("pkg owner %s: %w", path, snack.ErrNotFound)
		}
		return "", fmt.Errorf("pkg owner: %w", err)
	}
	return parseOwner(out), nil
}
