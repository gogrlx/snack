package flatpak

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseList(t *testing.T) {
	input := "Firefox\torg.mozilla.Firefox\t131.0\tflathub\n" +
		"GIMP\torg.gimp.GIMP\t2.10.38\tflathub\n"
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "Firefox" {
		t.Errorf("expected name 'Firefox', got %q", pkgs[0].Name)
	}
	if pkgs[0].Description != "org.mozilla.Firefox" {
		t.Errorf("expected description 'org.mozilla.Firefox', got %q", pkgs[0].Description)
	}
	if pkgs[0].Version != "131.0" {
		t.Errorf("expected version '131.0', got %q", pkgs[0].Version)
	}
	if pkgs[0].Repository != "flathub" {
		t.Errorf("expected repository 'flathub', got %q", pkgs[0].Repository)
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseListEmpty(t *testing.T) {
	pkgs := parseList("")
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseSearch(t *testing.T) {
	input := "Firefox\torg.mozilla.Firefox\t131.0\tflathub\n" +
		"Firefox ESR\torg.mozilla.FirefoxESR\t128.3\tflathub\n"
	pkgs := parseSearch(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "Firefox" {
		t.Errorf("unexpected name: %q", pkgs[0].Name)
	}
	if pkgs[1].Version != "128.3" {
		t.Errorf("unexpected version: %q", pkgs[1].Version)
	}
}

func TestParseSearchNoMatches(t *testing.T) {
	input := "No matches found\n"
	pkgs := parseSearch(input)
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseInfo(t *testing.T) {
	input := `Name: Firefox
Description: Fast, private web browser
Version: 131.0
Arch: x86_64
Origin: flathub
`
	pkg := parseInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "Firefox" {
		t.Errorf("expected name 'Firefox', got %q", pkg.Name)
	}
	if pkg.Version != "131.0" {
		t.Errorf("expected version '131.0', got %q", pkg.Version)
	}
	if pkg.Arch != "x86_64" {
		t.Errorf("expected arch 'x86_64', got %q", pkg.Arch)
	}
	if pkg.Repository != "flathub" {
		t.Errorf("expected repository 'flathub', got %q", pkg.Repository)
	}
}

func TestParseInfoEmpty(t *testing.T) {
	pkg := parseInfo("")
	if pkg != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseRemotes(t *testing.T) {
	input := "flathub\thttps://dl.flathub.org/repo/\t\n" +
		"gnome-nightly\thttps://nightly.gnome.org/repo/\tdisabled\n"
	repos := parseRemotes(input)
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].ID != "flathub" {
		t.Errorf("expected ID 'flathub', got %q", repos[0].ID)
	}
	if repos[0].URL != "https://dl.flathub.org/repo/" {
		t.Errorf("unexpected URL: %q", repos[0].URL)
	}
	if !repos[0].Enabled {
		t.Error("expected first repo to be enabled")
	}
	if repos[1].Enabled {
		t.Error("expected second repo to be disabled")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ snack.Manager = (*Flatpak)(nil)
	var _ snack.Cleaner = (*Flatpak)(nil)
	var _ snack.RepoManager = (*Flatpak)(nil)
}

func TestName(t *testing.T) {
	f := New()
	if f.Name() != "flatpak" {
		t.Errorf("Name() = %q, want %q", f.Name(), "flatpak")
	}
}
