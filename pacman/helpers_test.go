package pacman

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gogrlx/snack"
)

func TestFormatTargets_Empty(t *testing.T) {
	assert.Empty(t, formatTargets(nil))
}

func TestFormatTargets_NamesOnly(t *testing.T) {
	targets := []snack.Target{
		{Name: "vim"},
		{Name: "git"},
		{Name: "curl"},
	}
	got := formatTargets(targets)
	assert.Equal(t, []string{"vim", "git", "curl"}, got)
}

func TestFormatTargets_WithVersions(t *testing.T) {
	targets := []snack.Target{
		{Name: "vim", Version: "9.1.0-1"},
		{Name: "git"},
		{Name: "curl", Version: "8.7.1-1"},
	}
	got := formatTargets(targets)
	assert.Equal(t, []string{"vim=9.1.0-1", "git", "curl=8.7.1-1"}, got)
}

func TestBuildArgs_AllOptions(t *testing.T) {
	opts := snack.Options{
		Root:      "/mnt",
		Sudo:      true,
		AssumeYes: true,
		DryRun:    true,
	}
	cmd, args := buildArgs([]string{"-S", "pkg"}, opts)
	assert.Equal(t, "sudo", cmd)
	// Should have: pacman -r /mnt -S pkg --noconfirm --print
	assert.Contains(t, args, "pacman")
	assert.Contains(t, args, "-r")
	assert.Contains(t, args, "/mnt")
	assert.Contains(t, args, "--noconfirm")
	assert.Contains(t, args, "--print")
}

func TestBuildArgs_NoOptions(t *testing.T) {
	cmd, args := buildArgs([]string{"-Q"}, snack.Options{})
	assert.Equal(t, "pacman", cmd)
	assert.Equal(t, []string{"-Q"}, args)
}

func TestBuildArgs_SudoPrependsCommand(t *testing.T) {
	cmd, args := buildArgs([]string{"-S", "vim"}, snack.Options{Sudo: true})
	assert.Equal(t, "sudo", cmd)
	require.True(t, len(args) >= 1)
	assert.Equal(t, "pacman", args[0])
}

func TestBuildArgs_RootBeforeBaseArgs(t *testing.T) {
	_, args := buildArgs([]string{"-S", "vim"}, snack.Options{Root: "/alt"})
	// -r /alt should come before -S vim
	rIdx := -1
	sIdx := -1
	for i, a := range args {
		if a == "-r" {
			rIdx = i
		}
		if a == "-S" {
			sIdx = i
		}
	}
	assert.Greater(t, sIdx, rIdx, "root flag should come before base args")
}

func TestParseUpgrades_Empty(t *testing.T) {
	assert.Empty(t, parseUpgrades(""))
}

func TestParseUpgrades_Standard(t *testing.T) {
	input := `linux 6.7.3.arch1-1 -> 6.7.4.arch1-1
vim 9.0.2-1 -> 9.1.0-1
`
	pkgs := parseUpgrades(input)
	require.Len(t, pkgs, 2)
	assert.Equal(t, "linux", pkgs[0].Name)
	assert.Equal(t, "6.7.4.arch1-1", pkgs[0].Version)
	assert.True(t, pkgs[0].Installed)
	assert.Equal(t, "vim", pkgs[1].Name)
	assert.Equal(t, "9.1.0-1", pkgs[1].Version)
}

func TestParseUpgrades_FallbackFormat(t *testing.T) {
	// Some versions of pacman might output "pkg newver" without the arrow
	input := "pkg 2.0\n"
	pkgs := parseUpgrades(input)
	require.Len(t, pkgs, 1)
	assert.Equal(t, "pkg", pkgs[0].Name)
	assert.Equal(t, "2.0", pkgs[0].Version)
}

func TestParseUpgrades_WhitespaceLines(t *testing.T) {
	input := "\n  \nlinux 6.7.3 -> 6.7.4\n\n"
	pkgs := parseUpgrades(input)
	require.Len(t, pkgs, 1)
}

func TestParseGroupPkgSet_Empty(t *testing.T) {
	set := parseGroupPkgSet("")
	assert.Empty(t, set)
}

func TestParseGroupPkgSet_Standard(t *testing.T) {
	input := `base-devel autoconf
base-devel automake
base-devel binutils
base-devel gcc
base-devel make
`
	set := parseGroupPkgSet(input)
	assert.Len(t, set, 5)
	assert.Contains(t, set, "autoconf")
	assert.Contains(t, set, "automake")
	assert.Contains(t, set, "binutils")
	assert.Contains(t, set, "gcc")
	assert.Contains(t, set, "make")
}

func TestParseGroupPkgSet_SingleField(t *testing.T) {
	// Lines with fewer than 2 fields should be skipped
	input := "orphan\ngroup pkg\n"
	set := parseGroupPkgSet(input)
	assert.Len(t, set, 1)
	assert.Contains(t, set, "pkg")
}

func TestParseGroupPkgSet_Duplicates(t *testing.T) {
	input := `group pkg1
group pkg1
group pkg2
`
	set := parseGroupPkgSet(input)
	assert.Len(t, set, 2)
}

func TestNew(t *testing.T) {
	p := New()
	assert.NotNil(t, p)
	assert.Equal(t, "pacman", p.Name())
}

func TestSupportsDryRun(t *testing.T) {
	p := New()
	assert.True(t, p.SupportsDryRun())
}
