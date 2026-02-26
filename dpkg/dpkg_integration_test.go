//go:build integration

package dpkg_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/dpkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Dpkg(t *testing.T) {
	var mgr snack.Manager = dpkg.New()
	if !mgr.Available() {
		t.Skip("dpkg not available")
	}
	ctx := context.Background()

	assert.Equal(t, "dpkg", mgr.Name())

	caps := snack.GetCapabilities(mgr)
	assert.True(t, caps.FileOwnership, "dpkg should support FileOwnership")
	assert.True(t, caps.NameNormalize, "dpkg should support NameNormalize")
	assert.False(t, caps.VersionQuery)
	assert.False(t, caps.Hold)
	assert.False(t, caps.Clean)
	assert.False(t, caps.RepoManagement)
	assert.False(t, caps.KeyManagement)
	assert.False(t, caps.Groups)

	t.Run("List", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)

		found := false
		for _, p := range pkgs {
			if p.Name == "bash" {
				found = true
				assert.NotEmpty(t, p.Version)
				assert.True(t, p.Installed)
				break
			}
		}
		assert.True(t, found, "bash should be in list")
	})

	t.Run("IsInstalled_True", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "bash")
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("IsInstalled_False", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "xyznonexistentpackage999")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Version", func(t *testing.T) {
		ver, err := mgr.Version(ctx, "bash")
		require.NoError(t, err)
		assert.NotEmpty(t, ver)
		t.Logf("bash version: %s", ver)
	})

	t.Run("Version_NotInstalled", func(t *testing.T) {
		_, err := mgr.Version(ctx, "xyznonexistentpackage999")
		assert.Error(t, err)
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

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "bash")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "xyznonexistentpackage999")
		require.NoError(t, err)
		assert.Empty(t, pkgs)
	})

	// --- FileOwner ---
	t.Run("FileOwner", func(t *testing.T) {
		fo, ok := mgr.(snack.FileOwner)
		require.True(t, ok)

		t.Run("FileList", func(t *testing.T) {
			files, err := fo.FileList(ctx, "bash")
			require.NoError(t, err)
			require.NotEmpty(t, files)

			found := false
			for _, f := range files {
				if f == "/usr/bin/bash" || f == "/bin/bash" {
					found = true
					break
				}
			}
			assert.True(t, found, "bash binary should be in file list")
		})

		t.Run("FileList_NotInstalled", func(t *testing.T) {
			_, err := fo.FileList(ctx, "xyznonexistentpackage999")
			assert.Error(t, err)
		})

		t.Run("Owner", func(t *testing.T) {
			owner, err := fo.Owner(ctx, "/usr/bin/dpkg")
			require.NoError(t, err)
			assert.NotEmpty(t, owner)
			assert.Contains(t, owner, "dpkg")
		})

		t.Run("Owner_NotFound", func(t *testing.T) {
			_, err := fo.Owner(ctx, "/nonexistent/path/xyz")
			assert.Error(t, err)
		})
	})

	// --- NameNormalizer ---
	t.Run("NameNormalizer", func(t *testing.T) {
		nn, ok := mgr.(snack.NameNormalizer)
		require.True(t, ok)

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

	// --- Operations that dpkg doesn't really support (exercise error paths) ---
	t.Run("Install_Unsupported", func(t *testing.T) {
		// dpkg install requires a .deb file path, not a package name
		// This should fail gracefully
		err := mgr.Install(ctx, snack.Targets("tree"))
		assert.Error(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		// dpkg doesn't have update, should be a no-op or error
		err := mgr.Update(ctx)
		_ = err // exercise path
	})
}
