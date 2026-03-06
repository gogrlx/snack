package dnf

import (
	"testing"

	"github.com/gogrlx/snack"
)

// Compile-time interface assertions — DNF implements all optional interfaces.
var (
	_ snack.Manager        = (*DNF)(nil)
	_ snack.VersionQuerier = (*DNF)(nil)
	_ snack.Holder         = (*DNF)(nil)
	_ snack.Cleaner        = (*DNF)(nil)
	_ snack.FileOwner      = (*DNF)(nil)
	_ snack.RepoManager    = (*DNF)(nil)
	_ snack.KeyManager     = (*DNF)(nil)
	_ snack.Grouper        = (*DNF)(nil)
	_ snack.NameNormalizer = (*DNF)(nil)
	_ snack.DryRunner      = (*DNF)(nil)
	_ snack.PackageUpgrader = (*DNF)(nil)
)

func TestName(t *testing.T) {
	d := New()
	if got := d.Name(); got != "dnf" {
		t.Errorf("Name() = %q, want %q", got, "dnf")
	}
}

func TestSupportsDryRun(t *testing.T) {
	d := New()
	if !d.SupportsDryRun() {
		t.Error("SupportsDryRun() = false, want true")
	}
}

func TestGetCapabilities(t *testing.T) {
	d := New()
	caps := snack.GetCapabilities(d)

	tests := []struct {
		name string
		got  bool
	}{
		{"VersionQuery", caps.VersionQuery},
		{"Hold", caps.Hold},
		{"Clean", caps.Clean},
		{"FileOwnership", caps.FileOwnership},
		{"RepoManagement", caps.RepoManagement},
		{"KeyManagement", caps.KeyManagement},
		{"Groups", caps.Groups},
		{"NameNormalize", caps.NameNormalize},
		{"DryRun", caps.DryRun},
		{"PackageUpgrade", caps.PackageUpgrade},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.got {
				t.Errorf("Capabilities.%s = false, want true", tt.name)
			}
		})
	}
}

func TestNormalizeNameMethod(t *testing.T) {
	d := New()
	tests := []struct {
		input, want string
	}{
		{"nginx.x86_64", "nginx"},
		{"curl", "curl"},
	}
	for _, tt := range tests {
		if got := d.NormalizeName(tt.input); got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseArchMethod(t *testing.T) {
	d := New()
	name, arch := d.ParseArch("nginx.x86_64")
	if name != "nginx" || arch != "x86_64" {
		t.Errorf("ParseArch(\"nginx.x86_64\") = (%q, %q), want (\"nginx\", \"x86_64\")", name, arch)
	}
}
