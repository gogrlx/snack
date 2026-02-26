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

	t.Run("Update", func(t *testing.T) {
		err := mgr.Update(ctx)
		require.NoError(t, err)
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "curl")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)
	})

	t.Run("Info", func(t *testing.T) {
		// apk info only works on installed packages, use "tree" after install
		// or test with a pre-installed package like "busybox"
		pkg, err := mgr.Info(ctx, "busybox")
		if err != nil {
			t.Skip("busybox not installed, skipping Info test")
		}
		require.NotNil(t, pkg)
	})

	t.Run("Install", func(t *testing.T) {
		err := mgr.Install(ctx, snack.Targets("tree"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "tree")
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("Version", func(t *testing.T) {
		ver, err := mgr.Version(ctx, "tree")
		require.NoError(t, err)
		assert.NotEmpty(t, ver)
	})

	t.Run("List", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		found := false
		for _, p := range pkgs {
			if p.Name == "tree" {
				found = true
				break
			}
		}
		assert.True(t, found, "tree should be in installed list")
	})

	t.Run("Remove", func(t *testing.T) {
		err := mgr.Remove(ctx, snack.Targets("tree"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled_After_Remove", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "tree")
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("Capabilities", func(t *testing.T) {
		if vq, ok := mgr.(snack.VersionQuerier); ok {
			ver, err := vq.LatestVersion(ctx, "busybox")
			require.NoError(t, err)
			assert.NotEmpty(t, ver)

			upgrades, err := vq.ListUpgrades(ctx)
			require.NoError(t, err)
			_ = upgrades
		}

		if cl, ok := mgr.(snack.Cleaner); ok {
			err := cl.Clean(ctx)
			require.NoError(t, err)
		}

		if fo, ok := mgr.(snack.FileOwner); ok {
			owner, err := fo.Owner(ctx, "/usr/bin/apk")
			if err == nil {
				assert.NotEmpty(t, owner)
			}
		}
	})
}
