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
	mgr := snap.New()
	if !mgr.Available() {
		t.Skip("snap not available")
	}
	ctx := context.Background()

	t.Run("Update", func(t *testing.T) {
		err := mgr.Update(ctx)
		require.NoError(t, err)
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "hello-world")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)
	})

	t.Run("Info", func(t *testing.T) {
		pkg, err := mgr.Info(ctx, "hello-world")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Equal(t, "hello-world", pkg.Name)
	})

	t.Run("Install", func(t *testing.T) {
		err := mgr.Install(ctx, snack.Targets("hello-world"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "hello-world")
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("Version", func(t *testing.T) {
		ver, err := mgr.Version(ctx, "hello-world")
		require.NoError(t, err)
		assert.NotEmpty(t, ver)
	})

	t.Run("List", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		found := false
		for _, p := range pkgs {
			if p.Name == "hello-world" {
				found = true
				break
			}
		}
		assert.True(t, found, "hello-world should be in installed list")
	})

	t.Run("Remove", func(t *testing.T) {
		err := mgr.Remove(ctx, snack.Targets("hello-world"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled_After_Remove", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "hello-world")
		require.NoError(t, err)
		assert.False(t, installed)
	})
}
