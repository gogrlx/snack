//go:build integration

package rpm_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/rpm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_RPM(t *testing.T) {
	mgr := rpm.New()
	if !mgr.Available() {
		t.Skip("rpm not available")
	}
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		pkgs, err := mgr.List(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, pkgs)
	})

	t.Run("IsInstalled", func(t *testing.T) {
		installed, err := mgr.IsInstalled(ctx, "bash")
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("Version", func(t *testing.T) {
		ver, err := mgr.Version(ctx, "bash")
		require.NoError(t, err)
		assert.NotEmpty(t, ver)
	})

	t.Run("Info", func(t *testing.T) {
		pkg, err := mgr.Info(ctx, "bash")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Equal(t, "bash", pkg.Name)
	})

	t.Run("FileOwner", func(t *testing.T) {
		if fo, ok := mgr.(snack.FileOwner); ok {
			owner, err := fo.Owner(ctx, "/bin/bash")
			require.NoError(t, err)
			assert.NotEmpty(t, owner)

			files, err := fo.FileList(ctx, "bash")
			require.NoError(t, err)
			assert.NotEmpty(t, files)
		}
	})
}
