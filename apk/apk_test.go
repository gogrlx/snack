package apk

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestSplitNameVersion(t *testing.T) {
	tests := []struct {
		input   string
		name    string
		version string
	}{
		{"curl-8.5.0-r0", "curl", "8.5.0-r0"},
		{"musl-1.2.4-r2", "musl", "1.2.4-r2"},
		{"libcurl-doc-8.5.0-r0", "libcurl-doc", "8.5.0-r0"},
		{"go-1.21.5-r0", "go", "1.21.5-r0"},
		{"noversion", "noversion", ""},
	}
	for _, tt := range tests {
		name, ver := splitNameVersion(tt.input)
		if name != tt.name || ver != tt.version {
			t.Errorf("splitNameVersion(%q) = (%q, %q), want (%q, %q)",
				tt.input, name, ver, tt.name, tt.version)
		}
	}
}

func TestParseListInstalled(t *testing.T) {
	output := `curl-8.5.0-r0 x86_64 {curl} (MIT) [installed]
musl-1.2.4-r2 x86_64 {musl} (MIT) [installed]
`
	pkgs := parseListInstalled(output)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" || pkgs[0].Version != "8.5.0-r0" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
	if pkgs[0].Arch != "x86_64" {
		t.Errorf("expected arch x86_64, got %q", pkgs[0].Arch)
	}
}

func TestParseSearch(t *testing.T) {
	// verbose output
	output := `curl-8.5.0-r0 - URL retrieval utility and library
curl-doc-8.5.0-r0 - URL retrieval utility and library (documentation)
`
	pkgs := parseSearch(output)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" || pkgs[0].Version != "8.5.0-r0" {
		t.Errorf("unexpected package: %+v", pkgs[0])
	}
	if pkgs[0].Description != "URL retrieval utility and library" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
}

func TestParseSearchPlain(t *testing.T) {
	output := `curl-8.5.0-r0
curl-doc-8.5.0-r0
`
	pkgs := parseSearch(output)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" {
		t.Errorf("expected curl, got %q", pkgs[0].Name)
	}
}

func TestParseInfo(t *testing.T) {
	output := `curl-8.5.0-r0 installed size:
description: URL retrieval utility and library
arch: x86_64
webpage: https://curl.se/
`
	pkg := parseInfo(output)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Description != "URL retrieval utility and library" {
		t.Errorf("unexpected description: %q", pkg.Description)
	}
	if pkg.Arch != "x86_64" {
		t.Errorf("unexpected arch: %q", pkg.Arch)
	}
}

func TestParseInfoNameVersion(t *testing.T) {
	output := "curl-8.5.0-r0 description:\nsome stuff"
	name, ver := parseInfoNameVersion(output)
	if name != "curl" || ver != "8.5.0-r0" {
		t.Errorf("got (%q, %q), want (curl, 8.5.0-r0)", name, ver)
	}
}

func TestNewImplementsManager(t *testing.T) {
	var _ snack.Manager = New()
}

func TestName(t *testing.T) {
	a := New()
	if a.Name() != "apk" {
		t.Errorf("expected apk, got %q", a.Name())
	}
}
