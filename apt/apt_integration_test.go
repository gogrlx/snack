//go:build integration

package apt_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/apt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Apt(t *testing.T) {
	var mgr snack.Manager = apt.New()
	if !mgr.Available() {
		t.Skip("apt not available")
	}
	ctx := context.Background()

	// Verify Name
	assert.Equal(t, "apt", mgr.Name())

	// Verify capabilities reported correctly
	caps := snack.GetCapabilities(mgr)
	assert.True(t, caps.VersionQuery, "apt should support VersionQuery")
	assert.True(t, caps.Hold, "apt should support Hold")
	assert.True(t, caps.Clean, "apt should support Clean")
	assert.True(t, caps.FileOwnership, "apt should support FileOwnership")
	assert.True(t, caps.RepoManagement, "apt should support RepoManagement")
	assert.True(t, caps.KeyManagement, "apt should support KeyManagement")
	assert.True(t, caps.NameNormalize, "apt should support NameNormalize")
	assert.False(t, caps.Groups, "apt should not support Groups")

	t.Run("Update", func(t *testing.T) {
		require.NoError(t, mgr.Update(ctx))
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "curl")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs, "search for curl should return results")

		// Verify Package struct fields are populated
		found := false
		for _, p := range pkgs {
			if p.Name == "curl" {
				found = true
				assert.NotEmpty(t, p.Description, "curl should have a description")
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
		// Remove first to ensure clean state
		_ = mgr.Remove(ctx, snack.Targets("tree"), snack.WithAssumeYes())

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
				assert.True(t, p.Installed)
				break
			}
		}
		assert.True(t, found, "tree should be in installed list")
	})

	t.Run("Install_WithVersion", func(t *testing.T) {
		// Get current version and reinstall with explicit version
		ver, err := mgr.Version(ctx, "tree")
		require.NoError(t, err)

		err = mgr.Install(ctx, []snack.Target{{Name: "tree", Version: ver}}, snack.WithAssumeYes())
		// Should succeed (already installed at that version) or be a no-op
		// Some apt versions may return success, others may note it's already installed
		_ = err
	})

	t.Run("Purge", func(t *testing.T) {
		// Install then purge (removes config files too)
		_ = mgr.Install(ctx, snack.Targets("tree"), snack.WithAssumeYes())
		err := mgr.Purge(ctx, snack.Targets("tree"), snack.WithAssumeYes())
		require.NoError(t, err)

		installed, err := mgr.IsInstalled(ctx, "tree")
		require.NoError(t, err)
		assert.False(t, installed)
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
		require.True(t, ok, "apt must implement VersionQuerier")

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
			// May be empty in a freshly updated container, that's fine
			t.Logf("upgradable packages: %d", len(pkgs))
			for _, p := range pkgs {
				assert.NotEmpty(t, p.Name)
				assert.NotEmpty(t, p.Version)
			}
		})

		t.Run("UpgradeAvailable", func(t *testing.T) {
			// bash is typically at latest in a fresh container
			avail, err := vq.UpgradeAvailable(ctx, "bash")
			require.NoError(t, err)
			_ = avail // just exercising the code path
		})

		t.Run("VersionCmp", func(t *testing.T) {
			tests := []struct {
				v1, v2 string
				want   int
			}{
				{"1.0-1", "1.0-2", -1},
				{"2.0-1", "1.0-1", 1},
				{"1.0-1", "1.0-1", 0},
				{"1:1.0-1", "1.0-1", 1}, // epoch
			}
			for _, tt := range tests {
				cmp, err := vq.VersionCmp(ctx, tt.v1, tt.v2)
				require.NoError(t, err, "VersionCmp(%s, %s)", tt.v1, tt.v2)
				assert.Equal(t, tt.want, cmp, "VersionCmp(%s, %s)", tt.v1, tt.v2)
			}
		})
	})

	// --- Holder ---
	t.Run("Holder", func(t *testing.T) {
		h, ok := mgr.(snack.Holder)
		require.True(t, ok, "apt must implement Holder")

		// Ensure tree is installed for hold tests
		_ = mgr.Install(ctx, snack.Targets("tree"), snack.WithAssumeYes())

		t.Run("Hold", func(t *testing.T) {
			err := h.Hold(ctx, []string{"tree"})
			require.NoError(t, err)
		})

		t.Run("ListHeld", func(t *testing.T) {
			held, err := h.ListHeld(ctx)
			require.NoError(t, err)
			found := false
			for _, p := range held {
				if p.Name == "tree" {
					found = true
					break
				}
			}
			assert.True(t, found, "tree should be in held list")
		})

		t.Run("Unhold", func(t *testing.T) {
			err := h.Unhold(ctx, []string{"tree"})
			require.NoError(t, err)

			held, err := h.ListHeld(ctx)
			require.NoError(t, err)
			for _, p := range held {
				assert.NotEqual(t, "tree", p.Name, "tree should not be held after unhold")
			}
		})
	})

	// --- Cleaner ---
	t.Run("Cleaner", func(t *testing.T) {
		cl, ok := mgr.(snack.Cleaner)
		require.True(t, ok, "apt must implement Cleaner")

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
		require.True(t, ok, "apt must implement FileOwner")

		// Ensure tree is installed
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
	})

	// --- RepoManager ---
	t.Run("RepoManager", func(t *testing.T) {
		rm, ok := mgr.(snack.RepoManager)
		require.True(t, ok, "apt must implement RepoManager")

		t.Run("ListRepos", func(t *testing.T) {
			repos, err := rm.ListRepos(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, repos, "should have at least one repo")
			t.Logf("repos: %d", len(repos))
		})
	})

	// --- KeyManager ---
	t.Run("KeyManager", func(t *testing.T) {
		km, ok := mgr.(snack.KeyManager)
		require.True(t, ok, "apt must implement KeyManager")

		t.Run("ListKeys", func(t *testing.T) {
			keys, err := km.ListKeys(ctx)
			require.NoError(t, err)
			// May have keys or not, just exercising the path
			t.Logf("keys: %d", len(keys))
		})
	})

	// --- NameNormalizer ---
	t.Run("NameNormalizer", func(t *testing.T) {
		nn, ok := mgr.(snack.NameNormalizer)
		require.True(t, ok, "apt must implement NameNormalizer")

		tests := []struct {
			input, wantName, wantArch string
		}{
			{"curl:amd64", "curl", "amd64"},
			{"bash:arm64", "bash", "arm64"},
			{"python3", "python3", ""},
		}
		for _, tt := range tests {
			t.Run("NormalizeName_"+tt.input, func(t *testing.T) {
				got := nn.NormalizeName(tt.input)
				assert.Equal(t, tt.wantName, got)
			})
			t.Run("ParseArch_"+tt.input, func(t *testing.T) {
				name, arch := nn.ParseArch(tt.input)
				assert.Equal(t, tt.wantName, name)
				assert.Equal(t, tt.wantArch, arch)
			})
		}
	})

	// --- Upgrade (run last as it may change state) ---
	t.Run("Upgrade", func(t *testing.T) {
		err := mgr.Upgrade(ctx, snack.WithAssumeYes())
		require.NoError(t, err)
	})

	// Cleanup
	_ = mgr.Remove(ctx, snack.Targets("tree", "less"), snack.WithAssumeYes())
}
