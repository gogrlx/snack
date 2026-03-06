package rpm

import (
	"testing"

	"github.com/gogrlx/snack"
)

// Compile-time interface assertions.
var (
	_ snack.Manager        = (*RPM)(nil)
	_ snack.FileOwner      = (*RPM)(nil)
	_ snack.NameNormalizer = (*RPM)(nil)
)

func TestName(t *testing.T) {
	r := New()
	if got := r.Name(); got != "rpm" {
		t.Errorf("Name() = %q, want %q", got, "rpm")
	}
}

func TestGetCapabilities(t *testing.T) {
	r := New()
	caps := snack.GetCapabilities(r)

	wantTrue := map[string]bool{
		"FileOwnership": caps.FileOwnership,
		"NameNormalize": caps.NameNormalize,
	}
	for name, got := range wantTrue {
		t.Run(name+"_true", func(t *testing.T) {
			if !got {
				t.Errorf("Capabilities.%s = false, want true", name)
			}
		})
	}

	wantFalse := map[string]bool{
		"VersionQuery":   caps.VersionQuery,
		"Hold":           caps.Hold,
		"Clean":          caps.Clean,
		"RepoManagement": caps.RepoManagement,
		"KeyManagement":  caps.KeyManagement,
		"Groups":         caps.Groups,
		"DryRun":         caps.DryRun,
		"PackageUpgrade": caps.PackageUpgrade,
	}
	for name, got := range wantFalse {
		t.Run(name+"_false", func(t *testing.T) {
			if got {
				t.Errorf("Capabilities.%s = true, want false", name)
			}
		})
	}
}

func TestNormalizeNameMethod(t *testing.T) {
	r := New()
	tests := []struct {
		input, want string
	}{
		{"nginx.x86_64", "nginx"},
		{"curl", "curl"},
		{"bash.noarch", "bash"},
	}
	for _, tt := range tests {
		if got := r.NormalizeName(tt.input); got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseArchMethod(t *testing.T) {
	r := New()
	name, arch := r.ParseArch("nginx.x86_64")
	if name != "nginx" || arch != "x86_64" {
		t.Errorf("ParseArch(\"nginx.x86_64\") = (%q, %q), want (\"nginx\", \"x86_64\")", name, arch)
	}
}

// RPM should NOT implement DryRunner.
func TestNotDryRunner(t *testing.T) {
	r := New()
	if _, ok := interface{}(r).(snack.DryRunner); ok {
		t.Error("RPM should not implement DryRunner")
	}
}
