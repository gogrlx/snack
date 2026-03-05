package aur

import (
	"testing"

	"github.com/gogrlx/snack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePackageList(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{
			name:   "empty",
			input:  "",
			expect: 0,
		},
		{
			name:   "single package",
			input:  "yay 12.5.7-1\n",
			expect: 1,
		},
		{
			name:   "multiple packages",
			input:  "yay 12.5.7-1\nparu 2.0.4-1\naur-helper 1.0-1\n",
			expect: 3,
		},
		{
			name:   "trailing whitespace",
			input:  "yay 12.5.7-1  \n  paru 2.0.4-1\n\n",
			expect: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parsePackageList(tt.input)
			assert.Len(t, pkgs, tt.expect)
			for _, p := range pkgs {
				assert.NotEmpty(t, p.Name)
				assert.NotEmpty(t, p.Version)
				assert.Equal(t, "aur", p.Repository)
				assert.True(t, p.Installed)
			}
		})
	}
}

func TestNew(t *testing.T) {
	a := New()
	assert.Equal(t, "aur", a.Name())
	assert.Empty(t, a.BuildDir)
	assert.Nil(t, a.MakepkgFlags)
}

func TestNewWithOptions(t *testing.T) {
	a := NewWithOptions(
		WithBuildDir("/tmp/aur-builds"),
		WithMakepkgFlags("--skippgpcheck", "--nocheck"),
	)
	assert.Equal(t, "/tmp/aur-builds", a.BuildDir)
	assert.Equal(t, []string{"--skippgpcheck", "--nocheck"}, a.MakepkgFlags)
}

func TestInterfaceCompliance(t *testing.T) {
	var _ snack.Manager = (*AUR)(nil)
	var _ snack.VersionQuerier = (*AUR)(nil)
	var _ snack.Cleaner = (*AUR)(nil)
	var _ snack.PackageUpgrader = (*AUR)(nil)
}

func TestInterfaceNonCompliance(t *testing.T) {
	a := New()
	var m snack.Manager = a

	if _, ok := m.(snack.FileOwner); ok {
		t.Error("AUR should not implement FileOwner")
	}
	if _, ok := m.(snack.Holder); ok {
		t.Error("AUR should not implement Holder")
	}
	if _, ok := m.(snack.RepoManager); ok {
		t.Error("AUR should not implement RepoManager")
	}
	if _, ok := m.(snack.KeyManager); ok {
		t.Error("AUR should not implement KeyManager")
	}
	if _, ok := m.(snack.Grouper); ok {
		t.Error("AUR should not implement Grouper")
	}
	if _, ok := m.(snack.NameNormalizer); ok {
		t.Error("AUR should not implement NameNormalizer")
	}
	if _, ok := m.(snack.DryRunner); ok {
		t.Error("AUR should not implement DryRunner")
	}
}

func TestCapabilities(t *testing.T) {
	caps := snack.GetCapabilities(New())

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{"VersionQuery", caps.VersionQuery, true},
		{"Clean", caps.Clean, true},
		{"FileOwnership", caps.FileOwnership, false},
		{"Hold", caps.Hold, false},
		{"RepoManagement", caps.RepoManagement, false},
		{"KeyManagement", caps.KeyManagement, false},
		{"Groups", caps.Groups, false},
		{"NameNormalize", caps.NameNormalize, false},
		{"DryRun", caps.DryRun, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	a := New()
	assert.Equal(t, "aur", a.Name())
}

func TestParsePackageList_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantNames []string
		wantVers  []string
	}{
		{
			name:    "empty string",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "whitespace only",
			input:   "  \n\t\n  \n",
			wantLen: 0,
		},
		{
			name:      "single package",
			input:     "yay 12.5.7-1\n",
			wantLen:   1,
			wantNames: []string{"yay"},
			wantVers:  []string{"12.5.7-1"},
		},
		{
			name:    "malformed single field",
			input:   "orphan\n",
			wantLen: 0,
		},
		{
			name:      "malformed mixed with valid",
			input:     "orphan\nyay 12.5.7-1\nbadline\nparu 2.0-1\n",
			wantLen:   2,
			wantNames: []string{"yay", "paru"},
			wantVers:  []string{"12.5.7-1", "2.0-1"},
		},
		{
			name:      "extra fields ignored",
			input:     "yay 12.5.7-1 extra stuff\n",
			wantLen:   1,
			wantNames: []string{"yay"},
			wantVers:  []string{"12.5.7-1"},
		},
		{
			name:      "trailing and leading whitespace on lines",
			input:     "  yay 12.5.7-1  \n  paru 2.0.4-1\n\n",
			wantLen:   2,
			wantNames: []string{"yay", "paru"},
			wantVers:  []string{"12.5.7-1", "2.0.4-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parsePackageList(tt.input)
			require.Len(t, pkgs, tt.wantLen)
			for i, p := range pkgs {
				assert.Equal(t, "aur", p.Repository, "all packages should have Repository=aur")
				assert.True(t, p.Installed, "all packages should have Installed=true")
				if i < len(tt.wantNames) {
					assert.Equal(t, tt.wantNames[i], p.Name)
				}
				if i < len(tt.wantVers) {
					assert.Equal(t, tt.wantVers[i], p.Version)
				}
			}
		})
	}
}
