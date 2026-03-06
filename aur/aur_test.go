package aur

import (
	"testing"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks
var (
	_ snack.Manager         = (*AUR)(nil)
	_ snack.VersionQuerier  = (*AUR)(nil)
	_ snack.Cleaner         = (*AUR)(nil)
	_ snack.PackageUpgrader = (*AUR)(nil)
	_ snack.NameNormalizer  = (*AUR)(nil)
)

func TestNew(t *testing.T) {
	a := New()
	if a == nil {
		t.Fatal("New() returned nil")
	}
}

func TestName(t *testing.T) {
	a := New()
	if a.Name() != "aur" {
		t.Errorf("Name() = %q, want %q", a.Name(), "aur")
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"yay", "yay"},
		{"paru", "paru"},
		{"google-chrome", "google-chrome"},
		{"visual-studio-code-bin", "visual-studio-code-bin"},
		{"", ""},
	}

	a := New()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := a.NormalizeName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseArch(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantArch string
	}{
		{"yay", "yay", ""},
		{"paru", "paru", ""},
		{"google-chrome", "google-chrome", ""},
	}

	a := New()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotName, gotArch := a.ParseArch(tt.input)
			if gotName != tt.wantName || gotArch != tt.wantArch {
				t.Errorf("ParseArch(%q) = (%q, %q), want (%q, %q)",
					tt.input, gotName, gotArch, tt.wantName, tt.wantArch)
			}
		})
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
		{"PackageUpgrade", caps.PackageUpgrade, true},
		{"NameNormalize", caps.NameNormalize, true},
		// AUR does not support these
		{"Hold", caps.Hold, false},
		{"FileOwnership", caps.FileOwnership, false},
		{"RepoManagement", caps.RepoManagement, false},
		{"KeyManagement", caps.KeyManagement, false},
		{"Groups", caps.Groups, false},
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

func TestInterfaceNonCompliance(t *testing.T) {
	var m snack.Manager = New()
	if _, ok := m.(snack.Holder); ok {
		t.Error("AUR should not implement Holder")
	}
	if _, ok := m.(snack.FileOwner); ok {
		t.Error("AUR should not implement FileOwner")
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
}
