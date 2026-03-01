//go:build integration

package snap_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/snap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Snap(t *testing.T) {
	var mgr snack.Manager = snap.New()
	if !mgr.Available() {
		t.Skip("snap not available")
	}
	ctx := context.Background()

	assert.Equal(t, "snap", mgr.Name())

	caps := snack.GetCapabilities(mgr)
	assert.True(t, caps.VersionQuery, "snap should support VersionQuery")
	assert.False(t, caps.Hold)
	assert.False(t, caps.Clean)
	assert.False(t, caps.FileOwnership)
	assert.False(t, caps.RepoManagement)
	assert.False(t, caps.KeyManagement)
	assert.False(t, caps.Groups)
	assert.False(t, caps.NameNormalize)

	t.Run("Update", func(t *testing.T) {
		require.NoError(t, mgr.Update(ctx))
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "hello-world")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "xyznonexistentpackage999zzz")
		require.NoError(t, err)
		assert.Empty(t, pkgs)
	})

	t.Run("Info", func(t *testing.T) {
		pkg, err := mgr.Info(ctx, "hello-world")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Equal(t, "hello-world", pkg.Name)
		// Version may be empty for uninstalled snaps queried via snap info
	})

	t.Run("Info_NotFound", func(t *testing.T) {
		_, err := mgr.Info(ctx, "xyznonexistentpackage999")
		assert.Error(t, err)
	})

	t.Run("Install", func(t *testing.T) {
		_, _ = mgr.Remove(ctx, snack.Targets("hello-world"), snack.WithSudo())
		_, err := mgr.Install(ctx, snack.Targets("hello-world"), snack.WithSudo())
		require.NoError(t, err)
	})

	t.Run("IsInstalled_True", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "hello-world")
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("IsInstalled_False", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "xyznonexistentpackage999")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Version_Installed", func(t *testing.T) {
		ver, err := mgr.Version(ctx, "hello-world")
		require.NoError(t, err)
		assert.NotEmpty(t, ver)
		t.Logf("hello-world version: %s", ver)
	})

	t.Run("Version_NotInstalled", func(t *testing.T) {
		_, err := mgr.Version(ctx, "xyznonexistentpackage999")
		assert.Error(t, err)
	})

	t.Run("List_ContainsInstalled", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		found := false
		for _, p := range pkgs {
			if p.Name == "hello-world" {
				found = true
				assert.NotEmpty(t, p.Version)
				break
			}
		}
		assert.True(t, found, "hello-world should be in installed list")
	})

	// --- VersionQuerier ---
	t.Run("VersionQuerier", func(t *testing.T) {
		vq, ok := mgr.(snack.VersionQuerier)
		require.True(t, ok)

		t.Run("LatestVersion", func(t *testing.T) {
			ver, err := vq.LatestVersion(ctx, "hello-world")
			require.NoError(t, err)
			assert.NotEmpty(t, ver)
			t.Logf("hello-world latest: %s", ver)
		})

		t.Run("LatestVersion_NotFound", func(t *testing.T) {
			_, err := vq.LatestVersion(ctx, "xyznonexistentpackage999")
			assert.Error(t, err)
		})

		t.Run("ListUpgrades", func(t *testing.T) {
			pkgs, err := vq.ListUpgrades(ctx)
			require.NoError(t, err)
			t.Logf("upgradable snaps: %d", len(pkgs))
		})

		t.Run("UpgradeAvailable", func(t *testing.T) {
			avail, err := vq.UpgradeAvailable(ctx, "hello-world")
			require.NoError(t, err)
			_ = avail
		})

		t.Run("VersionCmp", func(t *testing.T) {
			tests := []struct {
				v1, v2 string
				want   int
			}{
				{"1.0", "2.0", -1},
				{"2.0", "1.0", 1},
				{"1.0", "1.0", 0},
				{"1.0.1", "1.0.0", 1},
			}
			for _, tt := range tests {
				cmp, err := vq.VersionCmp(ctx, tt.v1, tt.v2)
				require.NoError(t, err, "VersionCmp(%s, %s)", tt.v1, tt.v2)
				assert.Equal(t, tt.want, cmp, "VersionCmp(%s, %s)", tt.v1, tt.v2)
			}
		})
	})

	t.Run("Remove", func(t *testing.T) {
		_, err := mgr.Remove(ctx, snack.Targets("hello-world"), snack.WithSudo())
		require.NoError(t, err)

		installed, err := mgr.IsInstalled(ctx, "hello-world")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Purge", func(t *testing.T) {
		_, _ = mgr.Install(ctx, snack.Targets("hello-world"), snack.WithSudo())
		err := mgr.Purge(ctx, snack.Targets("hello-world"), snack.WithSudo())
		require.NoError(t, err)
	})

	t.Run("Upgrade", func(t *testing.T) {
		err := mgr.Upgrade(ctx, snack.WithSudo())
		require.NoError(t, err)
	})
}
