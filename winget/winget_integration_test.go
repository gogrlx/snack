//go:build integration

package winget_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/winget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Winget(t *testing.T) {
	var mgr snack.Manager = winget.New()
	if !mgr.Available() {
		t.Skip("winget not available")
	}
	ctx := context.Background()

	assert.Equal(t, "winget", mgr.Name())

	caps := snack.GetCapabilities(mgr)
	assert.True(t, caps.VersionQuery, "winget should support VersionQuery")
	assert.True(t, caps.RepoManagement, "winget should support RepoManagement")
	assert.True(t, caps.NameNormalize, "winget should support NameNormalize")
	assert.True(t, caps.PackageUpgrade, "winget should support PackageUpgrade")
	assert.False(t, caps.Hold)
	assert.False(t, caps.Clean)
	assert.False(t, caps.FileOwnership)
	assert.False(t, caps.KeyManagement)
	assert.False(t, caps.Groups)
	assert.False(t, caps.DryRun)

	t.Run("Update", func(t *testing.T) {
		require.NoError(t, mgr.Update(ctx))
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "Microsoft.PowerToys")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)
		t.Logf("search results: %d", len(pkgs))
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "xyznonexistentpackage999zzz")
		require.NoError(t, err)
		assert.Empty(t, pkgs)
	})

	t.Run("Info", func(t *testing.T) {
		pkg, err := mgr.Info(ctx, "Microsoft.PowerToys")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Equal(t, "Microsoft.PowerToys", pkg.Name)
		assert.NotEmpty(t, pkg.Version)
		t.Logf("PowerToys version: %s", pkg.Version)
	})

	t.Run("Info_NotFound", func(t *testing.T) {
		_, err := mgr.Info(ctx, "xyznonexistentpackage999.notreal")
		assert.Error(t, err)
	})

	t.Run("IsInstalled_False", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "xyznonexistentpackage999.notreal")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("List", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		t.Logf("installed packages: %d", len(pkgs))
		// Windows runners should have at least some packages
		assert.NotEmpty(t, pkgs)
	})

	// --- VersionQuerier ---
	t.Run("VersionQuerier", func(t *testing.T) {
		vq, ok := mgr.(snack.VersionQuerier)
		require.True(t, ok)

		t.Run("LatestVersion", func(t *testing.T) {
			ver, err := vq.LatestVersion(ctx, "Microsoft.PowerToys")
			require.NoError(t, err)
			assert.NotEmpty(t, ver)
			t.Logf("PowerToys latest: %s", ver)
		})

		t.Run("LatestVersion_NotFound", func(t *testing.T) {
			_, err := vq.LatestVersion(ctx, "xyznonexistentpackage999.notreal")
			assert.Error(t, err)
		})

		t.Run("ListUpgrades", func(t *testing.T) {
			pkgs, err := vq.ListUpgrades(ctx)
			require.NoError(t, err)
			t.Logf("upgradable packages: %d", len(pkgs))
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

	// --- RepoManager ---
	t.Run("RepoManager", func(t *testing.T) {
		rm, ok := mgr.(snack.RepoManager)
		require.True(t, ok)

		t.Run("ListRepos", func(t *testing.T) {
			repos, err := rm.ListRepos(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, repos)
			t.Logf("sources: %d", len(repos))
			// Default sources should include "winget"
			found := false
			for _, r := range repos {
				if r.Name == "winget" {
					found = true
					break
				}
			}
			assert.True(t, found, "winget source should be in list")
		})
	})

	// --- NameNormalizer ---
	t.Run("NameNormalizer", func(t *testing.T) {
		nn, ok := mgr.(snack.NameNormalizer)
		require.True(t, ok)

		assert.Equal(t, "Microsoft.PowerToys", nn.NormalizeName("Microsoft.PowerToys"))
		assert.Equal(t, "Git.Git", nn.NormalizeName("  Git.Git  "))

		name, arch := nn.ParseArch("Microsoft.PowerToys")
		assert.Equal(t, "Microsoft.PowerToys", name)
		assert.Empty(t, arch)
	})
}
