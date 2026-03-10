//go:build linux

package flatpak

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogrlx/snack"
)

func latestVersion(ctx context.Context, pkg string) (string, error) {
	out, err := run(ctx, []string{"remote-info", "flathub", pkg})
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("flatpak latestVersion %s: %w", pkg, snack.ErrNotFound)
		}
		return "", fmt.Errorf("flatpak latestVersion: %w", err)
	}
	// remote-info output is key:value like `flatpak info`
	p := parseInfo(out)
	if p == nil || p.Version == "" {
		return "", fmt.Errorf("flatpak latestVersion %s: %w", pkg, snack.ErrNotFound)
	}
	return p.Version, nil
}

func listUpgrades(ctx context.Context) ([]snack.Package, error) {
	out, err := run(ctx, []string{"remote-ls", "--updates", "--columns=name,application,version,origin"})
	if err != nil {
		// No updates available may produce an error on some versions
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
