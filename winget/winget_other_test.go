//go:build !windows

package winget

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
)

func TestUnavailableOnNonWindows(t *testing.T) {
	if New().Available() {
		t.Fatal("expected winget to be unavailable on non-Windows platforms")
	}
}

func TestUnsupportedOperationsOnNonWindows(t *testing.T) {
	ctx := context.Background()
	manager := New()
	targets := snack.Targets("Git.Git")

	installResult, err := manager.Install(ctx, targets)
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("Install() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if installResult.Installed != nil || installResult.Updated != nil || installResult.Unchanged != nil {
		t.Fatalf("Install() result = %+v, want zero value", installResult)
	}

	removeResult, err := manager.Remove(ctx, targets)
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("Remove() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if removeResult.Removed != nil || removeResult.Unchanged != nil {
		t.Fatalf("Remove() result = %+v, want zero value", removeResult)
	}

	installResult, err = manager.UpgradePackages(ctx, targets)
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("UpgradePackages() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if installResult.Installed != nil || installResult.Updated != nil || installResult.Unchanged != nil {
		t.Fatalf("UpgradePackages() result = %+v, want zero value", installResult)
	}

	checks := []struct {
		name string
		fn   func() error
	}{
		{"Purge", func() error { return manager.Purge(ctx, targets) }},
		{"Upgrade", func() error { return manager.Upgrade(ctx) }},
		{"Update", func() error { return manager.Update(ctx) }},
		{"AddRepo", func() error { return manager.AddRepo(ctx, snack.Repository{Name: "winget"}) }},
		{"RemoveRepo", func() error { return manager.RemoveRepo(ctx, "winget") }},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.fn(); err != snack.ErrUnsupportedPlatform {
				t.Fatalf("%s() error = %v, want %v", check.name, err, snack.ErrUnsupportedPlatform)
			}
		})
	}
}

func TestUnsupportedQueriesOnNonWindows(t *testing.T) {
	ctx := context.Background()
	manager := New()

	packages, err := manager.List(ctx)
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("List() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if packages != nil {
		t.Fatalf("List() packages = %#v, want nil", packages)
	}

	packages, err = manager.Search(ctx, "git")
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("Search() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if packages != nil {
		t.Fatalf("Search() packages = %#v, want nil", packages)
	}

	pkg, err := manager.Info(ctx, "Git.Git")
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("Info() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if pkg != nil {
		t.Fatalf("Info() package = %#v, want nil", pkg)
	}

	installed, err := manager.IsInstalled(ctx, "Git.Git")
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("IsInstalled() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if installed {
		t.Fatal("IsInstalled() = true, want false")
	}

	version, err := manager.Version(ctx, "Git.Git")
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("Version() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if version != "" {
		t.Fatalf("Version() = %q, want empty string", version)
	}

	version, err = manager.LatestVersion(ctx, "Git.Git")
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("LatestVersion() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if version != "" {
		t.Fatalf("LatestVersion() = %q, want empty string", version)
	}

	packages, err = manager.ListUpgrades(ctx)
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("ListUpgrades() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if packages != nil {
		t.Fatalf("ListUpgrades() packages = %#v, want nil", packages)
	}

	upgradeAvailable, err := manager.UpgradeAvailable(ctx, "Git.Git")
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("UpgradeAvailable() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if upgradeAvailable {
		t.Fatal("UpgradeAvailable() = true, want false")
	}

	cmp, err := manager.VersionCmp(ctx, "1.0.0", "1.0.1")
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("VersionCmp() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if cmp != 0 {
		t.Fatalf("VersionCmp() = %d, want 0", cmp)
	}

	repos, err := manager.ListRepos(ctx)
	if err != snack.ErrUnsupportedPlatform {
		t.Fatalf("ListRepos() error = %v, want %v", err, snack.ErrUnsupportedPlatform)
	}
	if repos != nil {
		t.Fatalf("ListRepos() repos = %#v, want nil", repos)
	}
}
