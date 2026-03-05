package aur

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
