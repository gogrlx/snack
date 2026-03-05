package pacman

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseList_Empty(t *testing.T) {
	assert.Empty(t, parseList(""))
}

func TestParseList_WhitespaceOnly(t *testing.T) {
	assert.Empty(t, parseList("   \n  \n\n"))
}

func TestParseList_SinglePackage(t *testing.T) {
	pkgs := parseList("vim 9.1.0-1\n")
	require.Len(t, pkgs, 1)
	assert.Equal(t, "vim", pkgs[0].Name)
	assert.Equal(t, "9.1.0-1", pkgs[0].Version)
	assert.True(t, pkgs[0].Installed)
}

func TestParseList_MalformedLine(t *testing.T) {
	// Line with only one field should be skipped
	pkgs := parseList("orphaned-line\nvalid 1.0\n")
	require.Len(t, pkgs, 1)
	assert.Equal(t, "valid", pkgs[0].Name)
}

func TestParseList_ExtraFields(t *testing.T) {
	// Extra fields beyond name+version should be ignored
	pkgs := parseList("pkg 1.0 extra stuff\n")
	require.Len(t, pkgs, 1)
	assert.Equal(t, "pkg", pkgs[0].Name)
	assert.Equal(t, "1.0", pkgs[0].Version)
}

func TestParseList_TrailingNewlines(t *testing.T) {
	pkgs := parseList("a 1.0\nb 2.0\n\n\n")
	require.Len(t, pkgs, 2)
}

func TestParseSearch_Empty(t *testing.T) {
	assert.Empty(t, parseSearch(""))
}

func TestParseSearch_NoDescription(t *testing.T) {
	// Package line with no following description line
	pkgs := parseSearch("core/pkg 1.0\n")
	require.Len(t, pkgs, 1)
	assert.Equal(t, "core", pkgs[0].Repository)
	assert.Equal(t, "pkg", pkgs[0].Name)
	assert.Equal(t, "1.0", pkgs[0].Version)
	assert.Empty(t, pkgs[0].Description)
}

func TestParseSearch_NoRepo(t *testing.T) {
	// Name without repo/ prefix
	pkgs := parseSearch("standalone 3.0\n    A standalone package\n")
	require.Len(t, pkgs, 1)
	assert.Empty(t, pkgs[0].Repository)
	assert.Equal(t, "standalone", pkgs[0].Name)
	assert.Equal(t, "A standalone package", pkgs[0].Description)
}

func TestParseSearch_InstalledBrackets(t *testing.T) {
	pkgs := parseSearch("extra/tmux 3.4-1 [installed]\n    A terminal multiplexer\n")
	require.Len(t, pkgs, 1)
	assert.True(t, pkgs[0].Installed)
}

func TestParseSearch_InstalledPartialVersion(t *testing.T) {
	// [installed: 3.3-1] format
	pkgs := parseSearch("extra/tmux 3.4-1 [installed: 3.3-1]\n    A terminal multiplexer\n")
	require.Len(t, pkgs, 1)
	assert.True(t, pkgs[0].Installed)
}

func TestParseSearch_NotInstalled(t *testing.T) {
	pkgs := parseSearch("community/rare-pkg 0.1-1\n    Something obscure\n")
	require.Len(t, pkgs, 1)
	assert.False(t, pkgs[0].Installed)
}

func TestParseSearch_MultiplePackages(t *testing.T) {
	input := `core/linux 6.7.4.arch1-1 [installed]
    The Linux kernel and modules
extra/linux-lts 6.6.14-1
    The LTS Linux kernel and modules
community/linux-zen 6.7.4.zen1-1
    The Zen Linux kernel
`
	pkgs := parseSearch(input)
	require.Len(t, pkgs, 3)
	assert.Equal(t, "linux", pkgs[0].Name)
	assert.Equal(t, "linux-lts", pkgs[1].Name)
	assert.Equal(t, "linux-zen", pkgs[2].Name)
}

func TestParseSearch_DescriptionOnlyLines(t *testing.T) {
	// Lines starting with whitespace without a preceding header should be skipped
	input := `    orphaned description
core/valid 1.0
    Real description
`
	pkgs := parseSearch(input)
	require.Len(t, pkgs, 1)
	assert.Equal(t, "valid", pkgs[0].Name)
	assert.Equal(t, "Real description", pkgs[0].Description)
}

func TestParseInfo_Empty(t *testing.T) {
	assert.Nil(t, parseInfo(""))
}

func TestParseInfo_NoName(t *testing.T) {
	// If Name field is missing, returns nil
	pkg := parseInfo("Version : 1.0\nDescription : Something\n")
	assert.Nil(t, pkg)
}

func TestParseInfo_AllFields(t *testing.T) {
	input := `Repository      : extra
Name            : neovim
Version         : 0.10.0-1
Description     : Fork of Vim aiming to improve user experience
Architecture    : x86_64
URL             : https://neovim.io
`
	pkg := parseInfo(input)
	require.NotNil(t, pkg)
	assert.Equal(t, "neovim", pkg.Name)
	assert.Equal(t, "0.10.0-1", pkg.Version)
	assert.Equal(t, "Fork of Vim aiming to improve user experience", pkg.Description)
	assert.Equal(t, "x86_64", pkg.Arch)
	assert.Equal(t, "extra", pkg.Repository)
}

func TestParseInfo_ColonInValue(t *testing.T) {
	// Value itself contains colons — only the first colon should be used as delimiter
	input := `Name            : pkg
Description     : A tool: does things: many of them
Version         : 1.0
`
	pkg := parseInfo(input)
	require.NotNil(t, pkg)
	assert.Equal(t, "A tool: does things: many of them", pkg.Description)
}

func TestParseInfo_UnrecognizedKeys(t *testing.T) {
	input := `Name            : test
Licenses        : MIT
Version         : 2.0
Packager        : Someone
`
	pkg := parseInfo(input)
	require.NotNil(t, pkg)
	assert.Equal(t, "test", pkg.Name)
	assert.Equal(t, "2.0", pkg.Version)
}
