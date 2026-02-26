//go:build integration

package detect_test

import (
	"testing"

	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/detect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Detect(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		mgr, err := detect.Default()
		require.NoError(t, err)
		require.NotNil(t, mgr)
		assert.NotEmpty(t, mgr.Name())
		assert.True(t, mgr.Available())
		t.Logf("Detected default: %s", mgr.Name())
	})

	t.Run("All", func(t *testing.T) {
		all := detect.All()
		require.NotEmpty(t, all)
		for _, m := range all {
			assert.NotEmpty(t, m.Name())
			assert.True(t, m.Available())
			t.Logf("Available: %s", m.Name())
		}
	})

	t.Run("ByName_Valid", func(t *testing.T) {
		all := detect.All()
		require.NotEmpty(t, all)

		// Should be able to find each detected manager by name
		for _, m := range all {
			found, err := detect.ByName(m.Name())
			require.NoError(t, err, "ByName(%s)", m.Name())
			require.NotNil(t, found)
			assert.Equal(t, m.Name(), found.Name())
		}
	})

	t.Run("ByName_Invalid", func(t *testing.T) {
		_, err := detect.ByName("xyznonexistentmanager999")
		assert.Error(t, err)
		assert.ErrorIs(t, err, snack.ErrManagerNotFound)
	})

	t.Run("Capabilities", func(t *testing.T) {
		mgr, err := detect.Default()
		require.NoError(t, err)

		caps := snack.GetCapabilities(mgr)
		t.Logf("Default manager %s capabilities: %+v", mgr.Name(), caps)

		// Every manager should have at least the base Manager interface
		// (which isn't in Capabilities, but let's verify some basics)
		assert.NotEmpty(t, mgr.Name())
	})
}
