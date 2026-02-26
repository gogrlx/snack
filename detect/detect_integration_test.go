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
	mgr, err := detect.Default()
	require.NoError(t, err)
	require.NotNil(t, mgr)
	t.Logf("Detected: %s", mgr.Name())

	all := detect.All()
	require.NotEmpty(t, all)
	for _, m := range all {
		t.Logf("Available: %s", m.Name())
	}

	caps := snack.GetCapabilities(mgr)
	t.Logf("Capabilities: %+v", caps)
	assert.NotEmpty(t, mgr.Name())
}
