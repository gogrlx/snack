package rpm

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseList(t *testing.T) {
	input := "bash\t5.1.8-6.el9\tThe GNU Bourne Again shell\ncurl\t7.76.1-23.el9\tA utility for getting files from remote servers\n"
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || pkgs[0].Version != "5.1.8-6.el9" {
		t.Errorf("unexpected pkg[0]: %+v", pkgs[0])
	}
	if pkgs[0].Description != "The GNU Bourne Again shell" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseInfo(t *testing.T) {
	input := `Name        : bash
Version     : 5.1.8
Release     : 6.el9
Architecture: x86_64
Install Date: Mon 01 Jan 2024 12:00:00 AM UTC
Group       : System Environment/Shells
Size        : 7896043
License     : GPLv3+
Signature   : RSA/SHA256, Mon 01 Jan 2024 12:00:00 AM UTC, Key ID abc123
Source RPM  : bash-5.1.8-6.el9.src.rpm
Build Date  : Mon 01 Jan 2024 12:00:00 AM UTC
Build Host  : builder.example.com
Packager    : CentOS Buildsys <bugs@centos.org>
Vendor      : CentOS
URL         : https://www.gnu.org/software/bash
Summary     : The GNU Bourne Again shell
Description :
The GNU Bourne Again shell (Bash) is a shell or command language
interpreter that is compatible with the Bourne shell (sh).
`
	p := parseInfo(input)
	if p == nil {
		t.Fatal("expected package, got nil")
	}
	if p.Name != "bash" {
		t.Errorf("Name = %q, want bash", p.Name)
	}
	if p.Version != "5.1.8-6.el9" {
		t.Errorf("Version = %q, want 5.1.8-6.el9", p.Version)
	}
	if p.Arch != "x86_64" {
		t.Errorf("Arch = %q, want x86_64", p.Arch)
	}
	if p.Description != "The GNU Bourne Again shell" {
		t.Errorf("Description = %q, want 'The GNU Bourne Again shell'", p.Description)
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"nginx.x86_64", "nginx"},
		{"curl.aarch64", "curl"},
		{"bash.noarch", "bash"},
		{"python3", "python3"},
	}
	for _, tt := range tests {
		got := normalizeName(tt.input)
		if got != tt.want {
			t.Errorf("normalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseArchSuffix(t *testing.T) {
	tests := []struct {
		input, wantName, wantArch string
	}{
		{"nginx.x86_64", "nginx", "x86_64"},
		{"bash", "bash", ""},
		{"glibc.i686", "glibc", "i686"},
	}
	for _, tt := range tests {
		name, arch := parseArchSuffix(tt.input)
		if name != tt.wantName || arch != tt.wantArch {
			t.Errorf("parseArchSuffix(%q) = (%q, %q), want (%q, %q)", tt.input, name, arch, tt.wantName, tt.wantArch)
		}
	}
}

// Compile-time interface checks.
var (
	_ snack.Manager        = (*RPM)(nil)
	_ snack.FileOwner      = (*RPM)(nil)
	_ snack.NameNormalizer = (*RPM)(nil)
)
