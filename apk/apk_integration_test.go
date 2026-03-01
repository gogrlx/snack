//go:build integration

package apk_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/apk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Apk(t *testing.T) {
	var mgr snack.Manager = apk.New()
	if !mgr.Available() {
		t.Skip("apk not available")
	}
	ctx := context.Background()

	assert.Equal(t, "apk", mgr.Name())

	caps := snack.GetCapabilities(mgr)
	assert.True(t, caps.VersionQuery, "apk should support VersionQuery")
	assert.True(t, caps.Clean, "apk should support Clean")
	assert.True(t, caps.FileOwnership, "apk should support FileOwnership")
	assert.False(t, caps.Hold, "apk should not support Hold")
	assert.False(t, caps.RepoManagement, "apk should not support RepoManagement")
	assert.False(t, caps.KeyManagement, "apk should not support KeyManagement")
	assert.False(t, caps.Groups, "apk should not support Groups")
	assert.False(t, caps.NameNormalize, "apk should not support NameNormalize")

	t.Run("Update", func(t *testing.T) {
		require.NoError(t, mgr.Update(ctx))
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "curl")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs, "search for curl should return results")

		found := false
		for _, p := range pkgs {
			if p.Name == "curl" {
				found = true
				break
			}
		}
		assert.True(t, found, "curl should appear in search results")
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "xyznonexistentpackage999")
		// apk search may return error or empty results for no matches
		if err == nil {
			assert.Empty(t, pkgs)
		}
	})

	t.Run("Info_PreInstalled", func(t *testing.T) {
		pkg, err := mgr.Info(ctx, "busybox")
		if err != nil {
			t.Skip("busybox not installed")
		}
		require.NotNil(t, pkg)
		assert.Equal(t, "busybox", pkg.Name)
		assert.NotEmpty(t, pkg.Version)
	})

	t.Run("Info_NotFound", func(t *testing.T) {
		_, err := mgr.Info(ctx, "xyznonexistentpackage999")
		assert.Error(t, err)
	})

	t.Run("Install_Single", func(t *testing.T) {
		_, _ = mgr.Remove(ctx, snack.Targets("tree"))
		_, err := mgr.Install(ctx, snack.Targets("tree"))
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

	t.Run("Info_AfterInstall", func(t *testing.T) {
		pkg, err := mgr.Info(ctx, "tree")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Equal(t, "tree", pkg.Name)
		assert.NotEmpty(t, pkg.Version)
	})

	t.Run("Install_Multiple", func(t *testing.T) {
		_, err := mgr.Install(ctx, snack.Targets("tree", "less"))
		require.NoError(t, err)

		for _, pkg := range []string{"tree", "less"} {
			installed, err := mgr.IsInstalled(ctx, pkg)
			require.NoError(t, err)
			assert.True(t, installed, "%s should be installed", pkg)
		}
	})

	t.Run("Purge", func(t *testing.T) {
		// apk purge is same as remove
		err := mgr.Purge(ctx, snack.Targets("less"))
		require.NoError(t, err)

		installed, err := mgr.IsInstalled(ctx, "less")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Remove", func(t *testing.T) {
		_, err := mgr.Remove(ctx, snack.Targets("tree"))
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
			avail, err := vq.UpgradeAvailable(ctx, "busybox")
			require.NoError(t, err)
			_ = avail
		})

		t.Run("VersionCmp", func(t *testing.T) {
			tests := []struct {
				v1, v2 string
				want   int
			}{
				{"1.0.0-r0", "1.0.0-r1", -1},
				{"2.0.0-r0", "1.0.0-r0", 1},
				{"1.0.0-r0", "1.0.0-r0", 0},
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
			// apk doesn't have autoremove but we exercise the path
			err := cl.Autoremove(ctx)
			// May or may not be supported
			_ = err
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

		// Install tree for file tests
		_, _ = mgr.Install(ctx, snack.Targets("tree"))

		t.Run("FileList", func(t *testing.T) {
			files, err := fo.FileList(ctx, "tree")
			require.NoError(t, err)
			require.NotEmpty(t, files)
			t.Logf("tree files: %d", len(files))
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

		_, _ = mgr.Remove(ctx, snack.Targets("tree"))
	})

	// --- Upgrade ---
	t.Run("Upgrade", func(t *testing.T) {
		err := mgr.Upgrade(ctx)
		require.NoError(t, err)
	})
}
