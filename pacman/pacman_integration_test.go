//go:build integration

package pacman_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/pacman"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Pacman(t *testing.T) {
	mgr := pacman.New()
	if !mgr.Available() {
		t.Skip("pacman not available")
	}
	ctx := context.Background()

	assert.Equal(t, "pacman", mgr.Name())

	caps := snack.GetCapabilities(mgr)
	assert.True(t, caps.VersionQuery, "pacman should support VersionQuery")
	assert.True(t, caps.Clean, "pacman should support Clean")
	assert.True(t, caps.FileOwnership, "pacman should support FileOwnership")
	assert.True(t, caps.Groups, "pacman should support Groups")
	assert.False(t, caps.Hold, "pacman should not support Hold")
	assert.False(t, caps.RepoManagement, "pacman should not support RepoManagement")
	assert.False(t, caps.KeyManagement, "pacman should not support KeyManagement")
	assert.False(t, caps.NameNormalize, "pacman should not support NameNormalize")

	t.Run("Update", func(t *testing.T) {
		require.NoError(t, mgr.Update(ctx))
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "curl")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)

		found := false
		for _, p := range pkgs {
			if p.Name == "curl" {
				found = true
				assert.NotEmpty(t, p.Description)
				break
			}
		}
		assert.True(t, found, "curl should appear in search results")
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "xyznonexistentpackage999")
		require.NoError(t, err)
		assert.Empty(t, pkgs)
	})

	t.Run("Info", func(t *testing.T) {
		pkg, err := mgr.Info(ctx, "bash")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Equal(t, "bash", pkg.Name)
		assert.NotEmpty(t, pkg.Version)
	})

	t.Run("Info_NotFound", func(t *testing.T) {
		_, err := mgr.Info(ctx, "xyznonexistentpackage999")
		assert.Error(t, err)
	})

	t.Run("Install_Single", func(t *testing.T) {
		_ = mgr.Remove(ctx, snack.Targets("tree"))
		err := mgr.Install(ctx, snack.Targets("tree"), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled_True", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "tree")
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("IsInstalled_False", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "xyznonexistentpackage999")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Version_Installed", func(t *testing.T) {
		ver, err := mgr.Version(ctx, "tree")
		require.NoError(t, err)
		assert.NotEmpty(t, ver)
		t.Logf("tree version: %s", ver)
	})

	t.Run("Version_NotInstalled", func(t *testing.T) {
		_, err := mgr.Version(ctx, "xyznonexistentpackage999")
		assert.Error(t, err)
	})

	t.Run("List_ContainsInstalled", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)

		found := false
		for _, p := range pkgs {
			if p.Name == "tree" {
				found = true
				assert.NotEmpty(t, p.Version)
				break
			}
		}
		assert.True(t, found, "tree should be in installed list")
	})

	t.Run("Install_Multiple", func(t *testing.T) {
		err := mgr.Install(ctx, snack.Targets("tree", "less"), snack.WithAssumeYes())
		require.NoError(t, err)

		for _, pkg := range []string{"tree", "less"} {
			installed, err := mgr.IsInstalled(ctx, pkg)
			require.NoError(t, err)
			assert.True(t, installed, "%s should be installed", pkg)
		}
	})

	t.Run("Purge", func(t *testing.T) {
		err := mgr.Purge(ctx, snack.Targets("less"), snack.WithAssumeYes())
		require.NoError(t, err)

		installed, err := mgr.IsInstalled(ctx, "less")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Remove", func(t *testing.T) {
		err := mgr.Remove(ctx, snack.Targets("tree"), snack.WithAssumeYes())
		require.NoError(t, err)

		installed, err := mgr.IsInstalled(ctx, "tree")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	// --- VersionQuerier ---
	t.Run("VersionQuerier", func(t *testing.T) {
		vq, ok := mgr.(snack.VersionQuerier)
		require.True(t, ok)

		t.Run("LatestVersion", func(t *testing.T) {
			ver, err := vq.LatestVersion(ctx, "curl")
			require.NoError(t, err)
			assert.NotEmpty(t, ver)
			t.Logf("curl latest: %s", ver)
		})

		t.Run("LatestVersion_NotFound", func(t *testing.T) {
			_, err := vq.LatestVersion(ctx, "xyznonexistentpackage999")
			assert.Error(t, err)
		})

		t.Run("ListUpgrades", func(t *testing.T) {
			pkgs, err := vq.ListUpgrades(ctx)
			require.NoError(t, err)
			t.Logf("upgradable packages: %d", len(pkgs))
		})

		t.Run("UpgradeAvailable", func(t *testing.T) {
			avail, err := vq.UpgradeAvailable(ctx, "bash")
			require.NoError(t, err)
			_ = avail
		})

		t.Run("VersionCmp", func(t *testing.T) {
			tests := []struct {
				v1, v2 string
				want   int
			}{
				{"1.0.0-1", "1.0.0-2", -1},
				{"2.0.0-1", "1.0.0-1", 1},
				{"1.0.0-1", "1.0.0-1", 0},
			}
			for _, tt := range tests {
				cmp, err := vq.VersionCmp(ctx, tt.v1, tt.v2)
				require.NoError(t, err, "VersionCmp(%s, %s)", tt.v1, tt.v2)
				assert.Equal(t, tt.want, cmp, "VersionCmp(%s, %s)", tt.v1, tt.v2)
			}
		})
	})

	// --- Cleaner ---
	t.Run("Cleaner", func(t *testing.T) {
		cl, ok := mgr.(snack.Cleaner)
		require.True(t, ok)

		t.Run("Autoremove", func(t *testing.T) {
			err := cl.Autoremove(ctx)
			require.NoError(t, err)
		})

		t.Run("Clean", func(t *testing.T) {
			err := cl.Clean(ctx)
			require.NoError(t, err)
		})
	})

	// --- FileOwner ---
	t.Run("FileOwner", func(t *testing.T) {
		fo, ok := mgr.(snack.FileOwner)
		require.True(t, ok)

		_ = mgr.Install(ctx, snack.Targets("tree"), snack.WithAssumeYes())

		t.Run("FileList", func(t *testing.T) {
			files, err := fo.FileList(ctx, "tree")
			require.NoError(t, err)
			require.NotEmpty(t, files)

			found := false
			for _, f := range files {
				if f == "/usr/bin/tree" {
					found = true
					break
				}
			}
			assert.True(t, found, "/usr/bin/tree should be in file list")
		})

		t.Run("FileList_NotInstalled", func(t *testing.T) {
			_, err := fo.FileList(ctx, "xyznonexistentpackage999")
			assert.Error(t, err)
		})

		t.Run("Owner", func(t *testing.T) {
			owner, err := fo.Owner(ctx, "/usr/bin/tree")
			require.NoError(t, err)
			assert.Contains(t, owner, "tree")
		})

		t.Run("Owner_NotFound", func(t *testing.T) {
			_, err := fo.Owner(ctx, "/nonexistent/path/xyz")
			assert.Error(t, err)
		})

		_ = mgr.Remove(ctx, snack.Targets("tree"), snack.WithAssumeYes())
	})

	// --- Grouper ---
	t.Run("Grouper", func(t *testing.T) {
		g, ok := mgr.(snack.Grouper)
		require.True(t, ok)

		t.Run("GroupList", func(t *testing.T) {
			groups, err := g.GroupList(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, groups)
			t.Logf("groups: %d", len(groups))

			found := false
			for _, grp := range groups {
				if grp == "base-devel" {
					found = true
					break
				}
			}
			assert.True(t, found, "base-devel group should exist")
		})

		t.Run("GroupInfo", func(t *testing.T) {
			pkgs, err := g.GroupInfo(ctx, "base-devel")
			require.NoError(t, err)
			require.NotEmpty(t, pkgs)
			t.Logf("base-devel packages: %d", len(pkgs))

			// Should contain gcc or make
			found := false
			for _, p := range pkgs {
				if p.Name == "gcc" || p.Name == "make" {
					found = true
					break
				}
			}
			assert.True(t, found, "base-devel should contain gcc or make")
		})

		t.Run("GroupInfo_NotFound", func(t *testing.T) {
			_, err := g.GroupInfo(ctx, "xyznonexistentgroup999")
			assert.Error(t, err)
		})
	})

	// --- Upgrade ---
	t.Run("Upgrade", func(t *testing.T) {
		err := mgr.Upgrade(ctx, snack.WithAssumeYes())
		require.NoError(t, err)
	})
}
