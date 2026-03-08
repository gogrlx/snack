package snack_test

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/stretchr/testify/assert"
)

// mockManager is a minimal Manager implementation for testing.
type mockManager struct{}

func (m *mockManager) Install(context.Context, []snack.Target, ...snack.Option) (snack.InstallResult, error) {
	return snack.InstallResult{}, nil
}
func (m *mockManager) Remove(context.Context, []snack.Target, ...snack.Option) (snack.RemoveResult, error) {
	return snack.RemoveResult{}, nil
}
func (m *mockManager) Purge(context.Context, []snack.Target, ...snack.Option) error { return nil }
func (m *mockManager) Upgrade(context.Context, ...snack.Option) error               { return nil }
func (m *mockManager) Update(context.Context) error                                 { return nil }
func (m *mockManager) List(context.Context) ([]snack.Package, error)                { return nil, nil }
func (m *mockManager) Search(context.Context, string) ([]snack.Package, error)      { return nil, nil }
func (m *mockManager) Info(context.Context, string) (*snack.Package, error)         { return nil, nil }
func (m *mockManager) IsInstalled(context.Context, string) (bool, error)            { return false, nil }
func (m *mockManager) Version(context.Context, string) (string, error)              { return "", nil }
func (m *mockManager) Available() bool                                              { return true }
func (m *mockManager) Name() string                                                 { return "mock" }

// fullMockManager implements Manager plus all optional interfaces.
type fullMockManager struct {
	mockManager
}

func (m *fullMockManager) LatestVersion(context.Context, string) (string, error) { return "", nil }
func (m *fullMockManager) ListUpgrades(context.Context) ([]snack.Package, error) { return nil, nil }
func (m *fullMockManager) UpgradeAvailable(context.Context, string) (bool, error) {
	return false, nil
}
func (m *fullMockManager) VersionCmp(context.Context, string, string) (int, error) { return 0, nil }
func (m *fullMockManager) Hold(context.Context, []string) error                    { return nil }
func (m *fullMockManager) Unhold(context.Context, []string) error                  { return nil }
func (m *fullMockManager) ListHeld(context.Context) ([]snack.Package, error)       { return nil, nil }
func (m *fullMockManager) IsHeld(context.Context, string) (bool, error)            { return false, nil }
func (m *fullMockManager) Autoremove(context.Context, ...snack.Option) error       { return nil }
func (m *fullMockManager) Clean(context.Context) error                             { return nil }
func (m *fullMockManager) FileList(context.Context, string) ([]string, error)      { return nil, nil }
func (m *fullMockManager) Owner(context.Context, string) (string, error)           { return "", nil }
func (m *fullMockManager) ListRepos(context.Context) ([]snack.Repository, error)   { return nil, nil }
func (m *fullMockManager) AddRepo(context.Context, snack.Repository) error         { return nil }
func (m *fullMockManager) RemoveRepo(context.Context, string) error                { return nil }
func (m *fullMockManager) AddKey(context.Context, string) error                    { return nil }
func (m *fullMockManager) RemoveKey(context.Context, string) error                 { return nil }
func (m *fullMockManager) ListKeys(context.Context) ([]string, error)              { return nil, nil }
func (m *fullMockManager) GroupList(context.Context) ([]string, error)             { return nil, nil }
func (m *fullMockManager) GroupInfo(context.Context, string) ([]snack.Package, error) {
	return nil, nil
}
func (m *fullMockManager) GroupInstall(context.Context, string, ...snack.Option) error { return nil }
func (m *fullMockManager) GroupIsInstalled(context.Context, string) (bool, error)      { return false, nil }
func (m *fullMockManager) NormalizeName(name string) string                            { return name }
func (m *fullMockManager) ParseArch(name string) (string, string)                      { return name, "" }
func (m *fullMockManager) SupportsDryRun() bool                                        { return true }
func (m *fullMockManager) UpgradePackages(context.Context, []snack.Target, ...snack.Option) (snack.InstallResult, error) {
	return snack.InstallResult{}, nil
}

func TestGetCapabilities_BaseManager(t *testing.T) {
	caps := snack.GetCapabilities(&mockManager{})
	assert.False(t, caps.VersionQuery)
	assert.False(t, caps.Hold)
	assert.False(t, caps.Clean)
	assert.False(t, caps.FileOwnership)
	assert.False(t, caps.RepoManagement)
	assert.False(t, caps.KeyManagement)
	assert.False(t, caps.Groups)
	assert.False(t, caps.NameNormalize)
	assert.False(t, caps.DryRun)
	assert.False(t, caps.PackageUpgrade)
}

func TestGetCapabilities_FullManager(t *testing.T) {
	caps := snack.GetCapabilities(&fullMockManager{})
	assert.True(t, caps.VersionQuery)
	assert.True(t, caps.Hold)
	assert.True(t, caps.Clean)
	assert.True(t, caps.FileOwnership)
	assert.True(t, caps.RepoManagement)
	assert.True(t, caps.KeyManagement)
	assert.True(t, caps.Groups)
	assert.True(t, caps.NameNormalize)
	assert.True(t, caps.DryRun)
	assert.True(t, caps.PackageUpgrade)
}
