//go:build integration

package flatpak_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/flatpak"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Flatpak(t *testing.T) {
	mgr := flatpak.New()
	if !mgr.Available() {
		t.Skip("flatpak not available")
	}
	ctx := context.Background()

	assert.Equal(t, "flatpak", mgr.Name())

	caps := snack.GetCapabilities(mgr)
	assert.True(t, caps.Clean, "flatpak should support Clean")
	assert.True(t, caps.RepoManagement, "flatpak should support RepoManagement")
	assert.False(t, caps.VersionQuery)
	assert.False(t, caps.Hold)
	assert.False(t, caps.FileOwnership)
	assert.False(t, caps.KeyManagement)
	assert.False(t, caps.Groups)
	assert.False(t, caps.NameNormalize)

	t.Run("Update", func(t *testing.T) {
		require.NoError(t, mgr.Update(ctx))
	})

	// --- RepoManager ---
	t.Run("RepoManager", func(t *testing.T) {
		rm, ok := mgr.(snack.RepoManager)
		require.True(t, ok)

		t.Run("ListRepos", func(t *testing.T) {
			repos, err := rm.ListRepos(ctx)
			require.NoError(t, err)
			found := false
			for _, r := range repos {
				if r.Name == "flathub" || r.ID == "flathub" {
					found = true
					break
				}
			}
			assert.True(t, found, "flathub repo should be configured")
		})
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "Flatseal")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "xyznonexistentpackage999")
		require.NoError(t, err)
		assert.Empty(t, pkgs)
	})

	t.Run("Install", func(t *testing.T) {
		err := mgr.Install(ctx, snack.Targets("com.github.tchx84.Flatseal"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled_True", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "com.github.tchx84.Flatseal")
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("IsInstalled_False", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "xyznonexistentpackage999")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Info", func(t *testing.T) {
		pkg, err := mgr.Info(ctx, "com.github.tchx84.Flatseal")
		require.NoError(t, err)
		require.NotNil(t, pkg)
	})

	t.Run("Info_NotFound", func(t *testing.T) {
		_, err := mgr.Info(ctx, "xyznonexistentpackage999")
		assert.Error(t, err)
	})

	t.Run("Version", func(t *testing.T) {
		ver, err := mgr.Version(ctx, "com.github.tchx84.Flatseal")
		require.NoError(t, err)
		assert.NotEmpty(t, ver)
		t.Logf("Flatseal version: %s", ver)
	})

	t.Run("List_ContainsInstalled", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		found := false
		for _, p := range pkgs {
			if p.Name == "Flatseal" || p.Description == "com.github.tchx84.Flatseal" {
				found = true
				break
			}
		}
		assert.True(t, found, "Flatseal should be in installed list")
	})

	t.Run("Remove", func(t *testing.T) {
		err := mgr.Remove(ctx, snack.Targets("com.github.tchx84.Flatseal"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled_After_Remove", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "com.github.tchx84.Flatseal")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	// --- Cleaner ---
	t.Run("Cleaner", func(t *testing.T) {
		cl, ok := mgr.(snack.Cleaner)
		require.True(t, ok)

		t.Run("Autoremove", func(t *testing.T) {
			err := cl.Autoremove(ctx, snack.WithAssumeYes())
			require.NoError(t, err)
		})

		t.Run("Clean", func(t *testing.T) {
			err := cl.Clean(ctx)
			require.NoError(t, err)
		})
	})

	t.Run("Upgrade", func(t *testing.T) {
		err := mgr.Upgrade(ctx, snack.WithAssumeYes())
		require.NoError(t, err)
	})
}
