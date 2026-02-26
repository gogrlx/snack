package pkg

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseQuery(t *testing.T) {
	input := "nginx\t1.24.0\tRobust and small WWW server\ncurl\t8.5.0\tCommand line tool for transferring data\n"
	pkgs := parseQuery(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "nginx" || pkgs[0].Version != "1.24.0" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[0].Description != "Robust and small WWW server" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseSearch(t *testing.T) {
	input := `nginx-1.24.0                   Robust and small WWW server
curl-8.5.0                     Command line tool for transferring data
`
	pkgs := parseSearch(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "nginx" || pkgs[0].Version != "1.24.0" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[1].Name != "curl" || pkgs[1].Version != "8.5.0" {
		t.Errorf("unexpected second package: %+v", pkgs[1])
	}
}

func TestParseInfo(t *testing.T) {
	input := `Name           : nginx
Version        : 1.24.0
Comment        : Robust and small WWW server
Arch           : FreeBSD:14:amd64
`
	pkg := parseInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "nginx" {
		t.Errorf("expected name 'nginx', got %q", pkg.Name)
	}
	if pkg.Version != "1.24.0" {
		t.Errorf("unexpected version: %q", pkg.Version)
	}
	if pkg.Description != "Robust and small WWW server" {
		t.Errorf("unexpected description: %q", pkg.Description)
	}
	if pkg.Arch != "FreeBSD:14:amd64" {
		t.Errorf("unexpected arch: %q", pkg.Arch)
	}
}

func TestParseUpgrades(t *testing.T) {
	input := `Updating FreeBSD repository catalogue...
The following 2 package(s) will be affected:

Upgrading nginx: 1.24.0 -> 1.26.0
Upgrading curl: 8.5.0 -> 8.6.0

Number of packages to be upgraded: 2
`
	pkgs := parseUpgrades(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "nginx" || pkgs[0].Version != "1.26.0" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[1].Name != "curl" || pkgs[1].Version != "8.6.0" {
		t.Errorf("unexpected second package: %+v", pkgs[1])
	}
}

func TestParseFileList(t *testing.T) {
	input := `nginx-1.24.0:
	/usr/local/sbin/nginx
	/usr/local/etc/nginx/nginx.conf
	/usr/local/share/doc/nginx/README
`
	files := parseFileList(input)
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}
	if files[0] != "/usr/local/sbin/nginx" {
		t.Errorf("unexpected file: %q", files[0])
	}
}

func TestParseOwner(t *testing.T) {
	input := "/usr/local/sbin/nginx was installed by package nginx-1.24.0\n"
	name := parseOwner(input)
	if name != "nginx" {
		t.Errorf("expected 'nginx', got %q", name)
	}
}

func TestSplitNameVersion(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion string
	}{
		{"nginx-1.24.0", "nginx", "1.24.0"},
		{"py39-pip-23.1", "py39-pip", "23.1"},
		{"bash", "bash", ""},
	}
	for _, tt := range tests {
		name, ver := splitNameVersion(tt.input)
		if name != tt.wantName || ver != tt.wantVersion {
			t.Errorf("splitNameVersion(%q) = (%q, %q), want (%q, %q)",
				tt.input, name, ver, tt.wantName, tt.wantVersion)
		}
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ snack.Manager = (*Pkg)(nil)
	var _ snack.VersionQuerier = (*Pkg)(nil)
	var _ snack.Cleaner = (*Pkg)(nil)
	var _ snack.FileOwner = (*Pkg)(nil)
}

func TestName(t *testing.T) {
	p := New()
	if p.Name() != "pkg" {
		t.Errorf("Name() = %q, want %q", p.Name(), "pkg")
	}
}
