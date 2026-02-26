//go:build integration

package flatpak_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/flatpak"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Flatpak(t *testing.T) {
	var mgr snack.Manager = flatpak.New()
	if !mgr.Available() {
		t.Skip("flatpak not available — install it first")
	}
	ctx := context.Background()

	t.Run("Update", func(t *testing.T) {
		err := mgr.Update(ctx)
		require.NoError(t, err)
	})

	t.Run("RepoManager", func(t *testing.T) {
		rm, ok := mgr.(snack.RepoManager)
		if !ok {
			t.Skip("RepoManager not implemented")
		}
		repos, err := rm.ListRepos(ctx)
		require.NoError(t, err)
		found := false
		for _, r := range repos {
			if r.Name == "flathub" || r.ID == "flathub" {
				found = true
				break
			}
		}
		assert.True(t, found, "flathub repo should be configured")
	})

	t.Run("Search", func(t *testing.T) {
		pkgs, err := mgr.Search(ctx, "Flatseal")
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)
	})

	t.Run("Install", func(t *testing.T) {
		err := mgr.Install(ctx, snack.Targets("com.github.tchx84.Flatseal"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "com.github.tchx84.Flatseal")
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("List", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		found := false
		for _, p := range pkgs {
			if p.Name == "Flatseal" || p.Description == "com.github.tchx84.Flatseal" {
				found = true
				break
			}
		}
		assert.True(t, found, "Flatseal should be in installed list (by Name or Application ID)")
	})

	t.Run("Remove", func(t *testing.T) {
		err := mgr.Remove(ctx, snack.Targets("com.github.tchx84.Flatseal"), snack.WithSudo(), snack.WithAssumeYes())
		require.NoError(t, err)
	})

	t.Run("IsInstalled_After_Remove", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "com.github.tchx84.Flatseal")
		require.NoError(t, err)
		assert.False(t, installed)
	})
}
