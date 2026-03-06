package pacman

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseList(t *testing.T) {
	input := `linux 6.7.4.arch1-1
glibc 2.39-1
bash 5.2.026-2
`
	pkgs := parseList(input)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "linux" || pkgs[0].Version != "6.7.4.arch1-1" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseSearch(t *testing.T) {
	input := `core/linux 6.7.4.arch1-1 [installed]
    The Linux kernel and modules
extra/linux-lts 6.6.14-1
    The LTS Linux kernel and modules
`
	pkgs := parseSearch(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Repository != "core" || pkgs[0].Name != "linux" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if !pkgs[0].Installed {
		t.Error("expected first package to be installed")
	}
	if pkgs[1].Installed {
		t.Error("expected second package to not be installed")
	}
	if pkgs[1].Description != "The LTS Linux kernel and modules" {
		t.Errorf("unexpected description: %q", pkgs[1].Description)
	}
}

func TestParseInfo(t *testing.T) {
	input := `Repository      : core
Name            : linux
Version         : 6.7.4.arch1-1
Description     : The Linux kernel and modules
Architecture    : x86_64
`
	pkg := parseInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "linux" {
		t.Errorf("expected name 'linux', got %q", pkg.Name)
	}
	if pkg.Version != "6.7.4.arch1-1" {
		t.Errorf("unexpected version: %q", pkg.Version)
	}
	if pkg.Arch != "x86_64" {
		t.Errorf("unexpected arch: %q", pkg.Arch)
	}
	if pkg.Repository != "core" {
		t.Errorf("unexpected repo: %q", pkg.Repository)
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		base     []string
		opts     snack.Options
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "basic",
			base:     []string{"-S", "vim"},
			opts:     snack.Options{},
			wantCmd:  "pacman",
			wantArgs: []string{"-S", "vim"},
		},
		{
			name:     "with sudo",
			base:     []string{"-S", "vim"},
			opts:     snack.Options{Sudo: true},
			wantCmd:  "sudo",
			wantArgs: []string{"pacman", "-S", "vim"},
		},
		{
			name:     "with root and noconfirm",
			base:     []string{"-S", "vim"},
			opts:     snack.Options{Root: "/mnt", AssumeYes: true},
			wantCmd:  "pacman",
			wantArgs: []string{"-r", "/mnt", "-S", "vim", "--noconfirm"},
		},
		{
			name:     "dry run",
			base:     []string{"-S", "vim"},
			opts:     snack.Options{DryRun: true},
			wantCmd:  "pacman",
			wantArgs: []string{"-S", "vim", "--print"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildArgs(tt.base, tt.opts)
			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ snack.Manager = (*Pacman)(nil)
	var _ snack.VersionQuerier = (*Pacman)(nil)
	var _ snack.Cleaner = (*Pacman)(nil)
	var _ snack.FileOwner = (*Pacman)(nil)
	var _ snack.Grouper = (*Pacman)(nil)
	var _ snack.DryRunner = (*Pacman)(nil)
	var _ snack.PackageUpgrader = (*Pacman)(nil)
}

func TestInterfaceNonCompliance(t *testing.T) {
	p := New()
	var m snack.Manager = p

	if _, ok := m.(snack.Holder); ok {
		t.Error("Pacman should not implement Holder")
	}
	if _, ok := m.(snack.RepoManager); ok {
		t.Error("Pacman should not implement RepoManager")
	}
	if _, ok := m.(snack.KeyManager); ok {
		t.Error("Pacman should not implement KeyManager")
	}
	if _, ok := m.(snack.NameNormalizer); ok {
		t.Error("Pacman should not implement NameNormalizer")
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
		{"FileOwnership", caps.FileOwnership, true},
		{"Groups", caps.Groups, true},
		{"DryRun", caps.DryRun, true},
		{"Hold", caps.Hold, false},
		{"RepoManagement", caps.RepoManagement, false},
		{"KeyManagement", caps.KeyManagement, false},
		{"NameNormalize", caps.NameNormalize, false},
		{"PackageUpgrade", caps.PackageUpgrade, true},
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
	p := New()
	if p.Name() != "pacman" {
		t.Errorf("Name() = %q, want %q", p.Name(), "pacman")
	}
}
