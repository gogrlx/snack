//go:build integration

package pkg_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Pkg(t *testing.T) {
	var mgr snack.Manager = pkg.New()
	if !mgr.Available() {
		t.Skip("pkg not available")
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
		p, err := mgr.Info(ctx, "curl")
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, "curl", p.Name)
	})

	t.Run("Install", func(t *testing.T) {
		_, err := mgr.Install(ctx, snack.Targets("tree"), snack.WithSudo(), snack.WithAssumeYes())
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
		_, err := mgr.Remove(ctx, snack.Targets("tree"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled_After_Remove", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "tree")
		require.NoError(t, err)
		assert.False(t, installed)
	})
}
