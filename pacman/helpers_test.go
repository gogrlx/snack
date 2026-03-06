//go:build linux

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

func TestParseUpgrades(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantNames []string
		wantVers  []string
	}{
		{
			name:    "empty",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "whitespace only",
			input:   "   \n  \n\n\t\n",
			wantLen: 0,
		},
		{
			name:      "standard arrow format",
			input:     "linux 6.7.3.arch1-1 -> 6.7.4.arch1-1\nvim 9.0.2-1 -> 9.1.0-1\n",
			wantLen:   2,
			wantNames: []string{"linux", "vim"},
			wantVers:  []string{"6.7.4.arch1-1", "9.1.0-1"},
		},
		{
			name:      "single package arrow format",
			input:     "curl 8.6.0-1 -> 8.7.1-1\n",
			wantLen:   1,
			wantNames: []string{"curl"},
			wantVers:  []string{"8.7.1-1"},
		},
		{
			name:      "fallback two-field format",
			input:     "pkg 2.0\n",
			wantLen:   1,
			wantNames: []string{"pkg"},
			wantVers:  []string{"2.0"},
		},
		{
			name:      "mixed arrow and fallback",
			input:     "linux 6.7.3 -> 6.7.4\npkg 2.0\n",
			wantLen:   2,
			wantNames: []string{"linux", "pkg"},
			wantVers:  []string{"6.7.4", "2.0"},
		},
		{
			name:      "whitespace around entries",
			input:     "\n  \nlinux 6.7.3 -> 6.7.4\n\n",
			wantLen:   1,
			wantNames: []string{"linux"},
			wantVers:  []string{"6.7.4"},
		},
		{
			name:    "single field line skipped",
			input:   "orphan\nvalid 1.0 -> 2.0\n",
			wantLen: 1,
		},
		{
			name:      "epoch in version",
			input:     "java-runtime 1:21.0.2-1 -> 1:21.0.3-1\n",
			wantLen:   1,
			wantNames: []string{"java-runtime"},
			wantVers:  []string{"1:21.0.3-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseUpgrades(tt.input)
			require.Len(t, pkgs, tt.wantLen)
			for i, p := range pkgs {
				assert.True(t, p.Installed, "all upgrade entries should have Installed=true")
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

func TestParseGroupPkgSet_WhitespaceOnly(t *testing.T) {
	set := parseGroupPkgSet("   \n  \n\t\n")
	assert.Empty(t, set)
}

func TestParseGroupPkgSet_MultipleGroups(t *testing.T) {
	// Different group names, same package names — set uses pkg name (second field)
	input := `base-devel gcc
xorg xorg-server
base-devel gcc
`
	set := parseGroupPkgSet(input)
	assert.Len(t, set, 2)
	assert.Contains(t, set, "gcc")
	assert.Contains(t, set, "xorg-server")
}

func TestParseGroupPkgSet_ExtraFields(t *testing.T) {
	// Lines with more than 2 fields — should still use second field
	input := "group pkg extra stuff\n"
	set := parseGroupPkgSet(input)
	assert.Len(t, set, 1)
	assert.Contains(t, set, "pkg")
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
