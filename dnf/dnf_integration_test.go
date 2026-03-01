//go:build integration

package dnf_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/dnf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_DNF_Detection(t *testing.T) {
	d := dnf.New()
	if !d.Available() {
		t.Skip("dnf not available")
	}
	t.Logf("IsDNF5: %v", d.IsDNF5())
}

func TestIntegration_DNF(t *testing.T) {
	var mgr snack.Manager = dnf.New()
	if !mgr.Available() {
		t.Skip("dnf not available")
	}
	ctx := context.Background()

	assert.Equal(t, "dnf", mgr.Name())

	caps := snack.GetCapabilities(mgr)
	assert.True(t, caps.VersionQuery, "dnf should support VersionQuery")
	assert.True(t, caps.Hold, "dnf should support Hold")
	assert.True(t, caps.Clean, "dnf should support Clean")
	assert.True(t, caps.FileOwnership, "dnf should support FileOwnership")
	assert.True(t, caps.RepoManagement, "dnf should support RepoManagement")
	assert.True(t, caps.KeyManagement, "dnf should support KeyManagement")
	assert.True(t, caps.Groups, "dnf should support Groups")
	assert.True(t, caps.NameNormalize, "dnf should support NameNormalize")

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
		assert.ErrorIs(t, err, snack.ErrNotFound)
	})

	t.Run("Install_Single", func(t *testing.T) {
		_, _ = mgr.Remove(ctx, snack.Targets("tree"), snack.WithAssumeYes())
		_, err := mgr.Install(ctx, snack.Targets("tree"), snack.WithAssumeYes())
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
		assert.ErrorIs(t, err, snack.ErrNotInstalled)
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
		_, err := mgr.Install(ctx, snack.Targets("tree", "less"), snack.WithAssumeYes())
		require.NoError(t, err)

		for _, pkg := range []string{"tree", "less"} {
			installed, err := mgr.IsInstalled(ctx, pkg)
			require.NoError(t, err)
			assert.True(t, installed, "%s should be installed", pkg)
		}
	})

	t.Run("Install_WithRefresh", func(t *testing.T) {
		_, err := mgr.Install(ctx, snack.Targets("tree"), snack.WithAssumeYes(), snack.WithRefresh())
		require.NoError(t, err)
	})

	t.Run("Purge", func(t *testing.T) {
		err := mgr.Purge(ctx, snack.Targets("less"), snack.WithAssumeYes())
		require.NoError(t, err)

		installed, err := mgr.IsInstalled(ctx, "less")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Remove", func(t *testing.T) {
		_, err := mgr.Remove(ctx, snack.Targets("tree"), snack.WithAssumeYes())
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
			for _, p := range pkgs {
				assert.NotEmpty(t, p.Name)
				assert.NotEmpty(t, p.Version)
			}
		})

		t.Run("UpgradeAvailable", func(t *testing.T) {
			avail, err := vq.UpgradeAvailable(ctx, "bash")
			require.NoError(t, err)
			_ = avail
		})

		t.Run("VersionCmp", func(t *testing.T) {
			// rpmdev-vercmp may not be installed; skip if not available
			cmp, err := vq.VersionCmp(ctx, "1.0-1", "1.0-2")
			if err != nil {
				t.Skip("rpmdev-vercmp not available")
			}
			assert.Equal(t, -1, cmp)

			cmp, err = vq.VersionCmp(ctx, "2.0-1", "1.0-1")
			require.NoError(t, err)
			assert.Equal(t, 1, cmp)

			cmp, err = vq.VersionCmp(ctx, "1.0-1", "1.0-1")
			require.NoError(t, err)
			assert.Equal(t, 0, cmp)
		})
	})

	// --- Holder ---
	t.Run("Holder", func(t *testing.T) {
		h, ok := mgr.(snack.Holder)
		require.True(t, ok)

		// Install tree for hold tests; also ensure versionlock plugin
		_, _ = mgr.Install(ctx, snack.Targets("tree"), snack.WithAssumeYes())

		t.Run("Hold", func(t *testing.T) {
			err := h.Hold(ctx, []string{"tree"})
			if err != nil {
				t.Skipf("versionlock plugin not available: %v", err)
			}

			t.Run("IsHeld", func(t *testing.T) {
				held, err := h.IsHeld(ctx, "tree")
				require.NoError(t, err)
				assert.True(t, held, "tree should be held")

				notHeld, err := h.IsHeld(ctx, "curl")
				require.NoError(t, err)
				assert.False(t, notHeld, "curl should not be held")
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
			})
		})

		_, _ = mgr.Remove(ctx, snack.Targets("tree"), snack.WithAssumeYes())
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

		_, _ = mgr.Install(ctx, snack.Targets("tree"), snack.WithAssumeYes())

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

		_, _ = mgr.Remove(ctx, snack.Targets("tree"), snack.WithAssumeYes())
	})

	// --- RepoManager ---
	t.Run("RepoManager", func(t *testing.T) {
		rm, ok := mgr.(snack.RepoManager)
		require.True(t, ok)

		t.Run("ListRepos", func(t *testing.T) {
			repos, err := rm.ListRepos(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, repos)
			for _, r := range repos {
				assert.NotEmpty(t, r.ID)
			}
			t.Logf("repos: %d", len(repos))
		})
	})

	// --- KeyManager ---
	t.Run("KeyManager", func(t *testing.T) {
		km, ok := mgr.(snack.KeyManager)
		require.True(t, ok)

		t.Run("ListKeys", func(t *testing.T) {
			keys, err := km.ListKeys(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, keys, "Fedora should have GPG keys")
			t.Logf("keys: %d", len(keys))
		})
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
		})

		t.Run("GroupInfo", func(t *testing.T) {
			// Use a group that should exist
			groups, err := g.GroupList(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, groups)

			pkgs, err := g.GroupInfo(ctx, groups[0])
			// Some groups may not have packages queryable by name
			if err == nil {
				t.Logf("group %q packages: %d", groups[0], len(pkgs))
			}
		})

		t.Run("GroupInfo_NotFound", func(t *testing.T) {
			// dnf5 may return empty instead of error for unknown groups
			pkgs, err := g.GroupInfo(ctx, "xyznonexistentgroup999")
			if err == nil {
				assert.Empty(t, pkgs)
			}
		})
	})

	// --- NameNormalizer ---
	t.Run("NameNormalizer", func(t *testing.T) {
		nn, ok := mgr.(snack.NameNormalizer)
		require.True(t, ok)

		tests := []struct {
			input, wantName, wantArch string
		}{
			{"curl.x86_64", "curl", "x86_64"},
			{"bash.aarch64", "bash", "aarch64"},
			{"python3.noarch", "python3", "noarch"},
			{"python3.11.x86_64", "python3.11", "x86_64"},
			{"glibc", "glibc", ""},
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

	// --- Upgrade ---
	t.Run("Upgrade", func(t *testing.T) {
		err := mgr.Upgrade(ctx, snack.WithAssumeYes())
		require.NoError(t, err)
	})
}
