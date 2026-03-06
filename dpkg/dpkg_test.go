package dpkg

import (
	"context"
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseList(t *testing.T) {
	input := "bash\t5.2-1\tinstall ok installed\ncoreutils\t9.1-1\tdeinstall ok config-files\n"
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || !pkgs[0].Installed {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[1].Installed {
		t.Errorf("expected second package not installed: %+v", pkgs[1])
	}
}

func TestParseDpkgList(t *testing.T) {
	input := `Desired=Unknown/Install/Remove/Purge/Hold
| Status=Not/Inst/Conf-files/Unpacked/halF-conf/Half-inst/trig-aWait/Trig-pend
|/ Err?=(none)/Reinst-required (Status,Err: uppercase=bad)
||/ Name           Version      Architecture Description
+++-==============-============-============-=================================
ii  bash           5.2-1        amd64        GNU Bourne Again SHell
rc  oldpkg         1.0-1        amd64        Some old package
`
	pkgs := parseDpkgList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || !pkgs[0].Installed {
		t.Errorf("unexpected: %+v", pkgs[0])
	}
	if pkgs[1].Installed {
		t.Errorf("expected oldpkg not installed: %+v", pkgs[1])
	}
}

func TestParseInfo(t *testing.T) {
	input := `Package: bash
Status: install ok installed
Version: 5.2-1
Architecture: amd64
Description: GNU Bourne Again SHell
`
	p, err := parseInfo(input)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "bash" || p.Version != "5.2-1" || !p.Installed {
		t.Errorf("unexpected: %+v", p)
	}
}

func TestParseInfoEmpty(t *testing.T) {
	_, err := parseInfo("")
	if err != snack.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNew(t *testing.T) {
	d := New()
	if d.Name() != "dpkg" {
		t.Errorf("expected 'dpkg', got %q", d.Name())
	}
}

func TestUpgradeUnsupported(t *testing.T) {
	d := New()
	if err := d.Upgrade(context.TODO()); err != snack.ErrUnsupportedPlatform {
		t.Errorf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

func TestUpdateUnsupported(t *testing.T) {
	d := New()
	if err := d.Update(context.TODO()); err != snack.ErrUnsupportedPlatform {
		t.Errorf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

// Compile-time interface assertions.
var (
	_ snack.Manager        = (*Dpkg)(nil)
	_ snack.FileOwner      = (*Dpkg)(nil)
	_ snack.NameNormalizer = (*Dpkg)(nil)
	_ snack.DryRunner      = (*Dpkg)(nil)
)

func TestSupportsDryRun(t *testing.T) {
	d := New()
	if !d.SupportsDryRun() {
		t.Error("expected SupportsDryRun() = true")
	}
}

func TestCapabilities(t *testing.T) {
	caps := snack.GetCapabilities(New())
	checks := []struct {
		name string
		got  bool
		want bool
	}{
		{"FileOwnership", caps.FileOwnership, true},
		{"NameNormalize", caps.NameNormalize, true},
		{"DryRun", caps.DryRun, true},
		{"VersionQuery", caps.VersionQuery, false},
		{"Hold", caps.Hold, false},
		{"Clean", caps.Clean, false},
		{"RepoManagement", caps.RepoManagement, false},
		{"KeyManagement", caps.KeyManagement, false},
		{"Groups", caps.Groups, false},
		{"PackageUpgrade", caps.PackageUpgrade, false},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if c.got != c.want {
				t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
			}
		})
	}
}

func TestNormalizeNameMethod(t *testing.T) {
	d := New()
	if got := d.NormalizeName("curl:amd64"); got != "curl" {
		t.Errorf("NormalizeName(curl:amd64) = %q, want %q", got, "curl")
	}
}

func TestParseArchMethod(t *testing.T) {
	d := New()
	name, arch := d.ParseArch("bash:arm64")
	if name != "bash" || arch != "arm64" {
		t.Errorf("ParseArch(bash:arm64) = (%q, %q), want (bash, arm64)", name, arch)
	}
}

func TestParseListEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"whitespace_only", "  \n  \n  ", 0},
		{"single_installed", "bash\t5.2-1\tinstall ok installed", 1},
		{"no_status_field", "bash\t5.2-1", 1},
		{"blank_lines_mixed", "\nbash\t5.2-1\tinstall ok installed\n\ncurl\t7.88\tinstall ok installed\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseList(tt.input)
			if len(pkgs) != tt.want {
				t.Errorf("parseList() returned %d packages, want %d", len(pkgs), tt.want)
			}
		})
	}
}

func TestParseDpkgListEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"header_only", "Desired=Unknown/Install/Remove/Purge/Hold\n||/ Name Version Architecture Description\n+++-====-====-====-====", 0},
		{"single_package", "ii  bash  5.2-1  amd64  GNU Bourne Again SHell", 1},
		{"held_package", "hi  nginx  1.24  amd64  web server", 1},
		{"purge_pending", "pn  oldpkg  1.0  amd64  old package", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseDpkgList(tt.input)
			if len(pkgs) != tt.want {
				t.Errorf("parseDpkgList() returned %d packages, want %d", len(pkgs), tt.want)
			}
		})
	}
}

func TestParseInfoEdgeCases(t *testing.T) {
	t.Run("no_package_field", func(t *testing.T) {
		_, err := parseInfo("Version: 1.0\nArchitecture: amd64\n")
		if err != snack.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("all_fields", func(t *testing.T) {
		input := "Package: vim\nStatus: install ok installed\nVersion: 9.0\nArchitecture: arm64\nDescription: Vi IMproved\n"
		p, err := parseInfo(input)
		if err != nil {
			t.Fatal(err)
		}
		if p.Name != "vim" || p.Version != "9.0" || p.Arch != "arm64" || !p.Installed {
			t.Errorf("unexpected: %+v", p)
		}
	})

	t.Run("version_with_epoch", func(t *testing.T) {
		input := "Package: systemd\nVersion: 1:252-2\n"
		p, err := parseInfo(input)
		if err != nil {
			t.Fatal(err)
		}
		if p.Version != "1:252-2" {
			t.Errorf("Version = %q, want %q", p.Version, "1:252-2")
		}
	})

	t.Run("not_installed_status", func(t *testing.T) {
		input := "Package: curl\nStatus: deinstall ok config-files\nVersion: 7.88\n"
		p, err := parseInfo(input)
		if err != nil {
			t.Fatal(err)
		}
		if p.Installed {
			t.Error("expected Installed=false for deinstall status")
		}
	})
}
