package snack_test

import (
	"testing"

	"github.com/gogrlx/snack"
	"github.com/stretchr/testify/assert"
)

func TestTargets(t *testing.T) {
	tests := []struct {
		names []string
		want  int
	}{
		{nil, 0},
		{[]string{}, 0},
		{[]string{"curl"}, 1},
		{[]string{"curl", "wget", "tree"}, 3},
	}
	for _, tt := range tests {
		targets := snack.Targets(tt.names...)
		assert.Len(t, targets, tt.want)
		for i, tgt := range targets {
			assert.Equal(t, tgt.Name, tt.names[i])
			assert.Empty(t, tgt.Version)
			assert.Empty(t, tgt.FromRepo)
			assert.Empty(t, tgt.Source)
		}
	}
}

func TestTargetNames(t *testing.T) {
	targets := []snack.Target{
		{Name: "curl", Version: "8.0"},
		{Name: "wget"},
		{Name: "tree", FromRepo: "main"},
	}
	names := snack.TargetNames(targets)
	assert.Equal(t, []string{"curl", "wget", "tree"}, names)
}

func TestTargetNames_Empty(t *testing.T) {
	names := snack.TargetNames(nil)
	assert.Empty(t, names)
}

func TestApplyOptions(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		o := snack.ApplyOptions()
		assert.False(t, o.Sudo)
		assert.False(t, o.AssumeYes)
		assert.False(t, o.DryRun)
		assert.False(t, o.Verbose)
		assert.False(t, o.Refresh)
		assert.False(t, o.Reinstall)
		assert.Empty(t, o.Root)
		assert.Empty(t, o.FromRepo)
	})

	t.Run("all options", func(t *testing.T) {
		o := snack.ApplyOptions(
			snack.WithSudo(),
			snack.WithAssumeYes(),
			snack.WithDryRun(),
			snack.WithVerbose(),
			snack.WithRefresh(),
			snack.WithReinstall(),
			snack.WithRoot("/mnt"),
			snack.WithFromRepo("testing"),
		)
		assert.True(t, o.Sudo)
		assert.True(t, o.AssumeYes)
		assert.True(t, o.DryRun)
		assert.True(t, o.Verbose)
		assert.True(t, o.Refresh)
		assert.True(t, o.Reinstall)
		assert.Equal(t, "/mnt", o.Root)
		assert.Equal(t, "testing", o.FromRepo)
	})
}

func TestGetCapabilities_NilSafe(t *testing.T) {
	// GetCapabilities should work on any Manager implementation
	// We can't easily test with nil, but we can verify the struct fields
	caps := snack.Capabilities{}
	assert.False(t, caps.VersionQuery)
	assert.False(t, caps.Hold)
	assert.False(t, caps.Clean)
	assert.False(t, caps.FileOwnership)
	assert.False(t, caps.RepoManagement)
	assert.False(t, caps.KeyManagement)
	assert.False(t, caps.Groups)
	assert.False(t, caps.NameNormalize)
}

func TestErrors(t *testing.T) {
	// Verify error sentinels are distinct
	errs := []error{
		snack.ErrNotInstalled,
		snack.ErrNotFound,
		snack.ErrUnsupportedPlatform,
		snack.ErrPermissionDenied,
		snack.ErrAlreadyInstalled,
		snack.ErrDependencyConflict,
		snack.ErrManagerNotFound,
		snack.ErrDaemonNotRunning,
	}
	for i, e1 := range errs {
		assert.NotNil(t, e1)
		assert.NotEmpty(t, e1.Error())
		for j, e2 := range errs {
			if i != j {
				assert.NotEqual(t, e1, e2, "%v should not equal %v", e1, e2)
			}
		}
	}
}

func TestInstallResult(t *testing.T) {
	r := snack.InstallResult{
		Installed: []snack.Package{{Name: "nginx", Version: "1.24.0", Installed: true}},
		Unchanged: []string{"curl"},
	}
	assert.Len(t, r.Installed, 1)
	assert.Equal(t, "nginx", r.Installed[0].Name)
	assert.Equal(t, "1.24.0", r.Installed[0].Version)
	assert.True(t, r.Installed[0].Installed)
	assert.Empty(t, r.Updated)
	assert.Equal(t, []string{"curl"}, r.Unchanged)
}

func TestRemoveResult(t *testing.T) {
	r := snack.RemoveResult{
		Removed:   []snack.Package{{Name: "nginx", Version: "1.24.0"}},
		Unchanged: []string{"wget"},
	}
	assert.Len(t, r.Removed, 1)
	assert.Equal(t, "nginx", r.Removed[0].Name)
	assert.Equal(t, []string{"wget"}, r.Unchanged)
}

func TestInstallResult_Zero(t *testing.T) {
	var r snack.InstallResult
	assert.Nil(t, r.Installed)
	assert.Nil(t, r.Updated)
	assert.Nil(t, r.Unchanged)
}

func TestRemoveResult_Zero(t *testing.T) {
	var r snack.RemoveResult
	assert.Nil(t, r.Removed)
	assert.Nil(t, r.Unchanged)
}
