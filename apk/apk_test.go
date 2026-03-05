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
		// Edge cases
		{"", "", ""},
		{"a-1", "a", "1"},
		{"my-pkg-name-0.1-r0", "my-pkg-name", "0.1-r0"},
		{"a-b-c-3.0", "a-b-c", "3.0"},
		{"single", "single", ""},
		{"-1.0", "-1.0", ""},  // no digit follows last hyphen at position > 0
		{"pkg-0", "pkg", "0"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, ver := splitNameVersion(tt.input)
			if name != tt.name || ver != tt.version {
				t.Errorf("splitNameVersion(%q) = (%q, %q), want (%q, %q)",
					tt.input, name, ver, tt.name, tt.version)
			}
		})
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

func TestParseListInstalledEdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		pkgs := parseListInstalled("")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		pkgs := parseListInstalled("   \n  \n  ")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("single package", func(t *testing.T) {
		pkgs := parseListInstalled("busybox-1.36.1-r5 x86_64 {busybox} (GPL-2.0-only) [installed]\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "busybox" || pkgs[0].Version != "1.36.1-r5" {
			t.Errorf("unexpected package: %+v", pkgs[0])
		}
	})

	t.Run("not installed", func(t *testing.T) {
		pkgs := parseListInstalled("curl-8.5.0-r0 x86_64 {curl} (MIT)\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Installed {
			t.Error("expected Installed=false")
		}
	})
}

func TestParseListLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantName  string
		wantVer   string
		wantArch  string
		installed bool
	}{
		{
			name:      "full line",
			line:      "curl-8.5.0-r0 x86_64 {curl} (MIT) [installed]",
			wantName:  "curl",
			wantVer:   "8.5.0-r0",
			wantArch:  "x86_64",
			installed: true,
		},
		{
			name:      "no installed marker",
			line:      "vim-9.0-r0 x86_64 {vim} (Vim)",
			wantName:  "vim",
			wantVer:   "9.0-r0",
			wantArch:  "x86_64",
			installed: false,
		},
		{
			name:      "name only",
			line:      "curl-8.5.0-r0",
			wantName:  "curl",
			wantVer:   "8.5.0-r0",
			wantArch:  "",
			installed: false,
		},
		{
			name:      "empty line",
			line:      "",
			wantName:  "",
			wantVer:   "",
			wantArch:  "",
			installed: false,
		},
		{
			name:      "aarch64 arch",
			line:      "openssl-3.1.4-r0 aarch64 {openssl} (Apache-2.0) [installed]",
			wantName:  "openssl",
			wantVer:   "3.1.4-r0",
			wantArch:  "aarch64",
			installed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := parseListLine(tt.line)
			if pkg.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", pkg.Name, tt.wantName)
			}
			if pkg.Version != tt.wantVer {
				t.Errorf("Version = %q, want %q", pkg.Version, tt.wantVer)
			}
			if pkg.Arch != tt.wantArch {
				t.Errorf("Arch = %q, want %q", pkg.Arch, tt.wantArch)
			}
			if pkg.Installed != tt.installed {
				t.Errorf("Installed = %v, want %v", pkg.Installed, tt.installed)
			}
		})
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

func TestParseSearchEdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		pkgs := parseSearch("")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("single result verbose", func(t *testing.T) {
		pkgs := parseSearch("nginx-1.24.0-r0 - HTTP and reverse proxy server\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "nginx" {
			t.Errorf("expected nginx, got %q", pkgs[0].Name)
		}
		if pkgs[0].Description != "HTTP and reverse proxy server" {
			t.Errorf("unexpected description: %q", pkgs[0].Description)
		}
	})

	t.Run("single result plain", func(t *testing.T) {
		pkgs := parseSearch("nginx-1.24.0-r0\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "nginx" || pkgs[0].Version != "1.24.0-r0" {
			t.Errorf("unexpected package: %+v", pkgs[0])
		}
		if pkgs[0].Description != "" {
			t.Errorf("expected empty description, got %q", pkgs[0].Description)
		}
	})

	t.Run("description with hyphens", func(t *testing.T) {
		pkgs := parseSearch("git-2.43.0-r0 - Distributed version control system - fast\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Description != "Distributed version control system - fast" {
			t.Errorf("unexpected description: %q", pkgs[0].Description)
		}
	})
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

func TestParseInfoEdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		pkg := parseInfo("")
		if pkg == nil {
			t.Fatal("expected non-nil (parseInfo always returns a pkg)")
		}
	})

	t.Run("no description", func(t *testing.T) {
		pkg := parseInfo("arch: aarch64\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Arch != "aarch64" {
			t.Errorf("expected aarch64, got %q", pkg.Arch)
		}
		if pkg.Description != "" {
			t.Errorf("expected empty description, got %q", pkg.Description)
		}
	})

	t.Run("multiple colons in value", func(t *testing.T) {
		pkg := parseInfo("description: A tool: does things: really well\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		// Note: strings.Cut splits on first colon only
		if pkg.Description != "A tool: does things: really well" {
			t.Errorf("unexpected description: %q", pkg.Description)
		}
	})
}

func TestParseInfoNameVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantN   string
		wantV   string
	}{
		{
			name:  "standard",
			input: "curl-8.5.0-r0 description:\nsome stuff",
			wantN: "curl",
			wantV: "8.5.0-r0",
		},
		{
			name:  "single line no version",
			input: "noversion",
			wantN: "noversion",
			wantV: "",
		},
		{
			name:  "multi-hyphen name",
			input: "lib-ssl-dev-3.0.0-r0 some text",
			wantN: "lib-ssl-dev",
			wantV: "3.0.0-r0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ver := parseInfoNameVersion(tt.input)
			if name != tt.wantN || ver != tt.wantV {
				t.Errorf("got (%q, %q), want (%q, %q)", name, ver, tt.wantN, tt.wantV)
			}
		})
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

// Compile-time interface compliance checks
var (
	_ snack.VersionQuerier  = (*Apk)(nil)
	_ snack.Cleaner         = (*Apk)(nil)
	_ snack.FileOwner       = (*Apk)(nil)
	_ snack.DryRunner       = (*Apk)(nil)
	_ snack.PackageUpgrader = (*Apk)(nil)
)

func TestInterfaceCompliance(t *testing.T) {
	// Verify at test time as well
	var m snack.Manager = New()
	if _, ok := m.(snack.VersionQuerier); !ok {
		t.Error("Apk should implement VersionQuerier")
	}
	if _, ok := m.(snack.Cleaner); !ok {
		t.Error("Apk should implement Cleaner")
	}
	if _, ok := m.(snack.FileOwner); !ok {
		t.Error("Apk should implement FileOwner")
	}
	if _, ok := m.(snack.DryRunner); !ok {
		t.Error("Apk should implement DryRunner")
	}
	if _, ok := m.(snack.PackageUpgrader); !ok {
		t.Error("Apk should implement PackageUpgrader")
	}
}

func TestCapabilities(t *testing.T) {
	caps := snack.GetCapabilities(New())
	if !caps.VersionQuery {
		t.Error("expected VersionQuery=true")
	}
	if !caps.Clean {
		t.Error("expected Clean=true")
	}
	if !caps.FileOwnership {
		t.Error("expected FileOwnership=true")
	}
	if !caps.DryRun {
		t.Error("expected DryRun=true")
	}
	// Should be false
	if caps.Hold {
		t.Error("expected Hold=false")
	}
	if caps.RepoManagement {
		t.Error("expected RepoManagement=false")
	}
	if caps.KeyManagement {
		t.Error("expected KeyManagement=false")
	}
	if caps.Groups {
		t.Error("expected Groups=false")
	}
	if caps.NameNormalize {
		t.Error("expected NameNormalize=false")
	}
}

func TestSupportsDryRun(t *testing.T) {
	a := New()
	if !a.SupportsDryRun() {
		t.Error("SupportsDryRun() should return true")
	}
}

func TestParseUpgradeSimulation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantPkgs []snack.Package
	}{
		{
			name:    "empty",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "OK only",
			input:   "OK: 123 MiB in 45 packages\n",
			wantLen: 0,
		},
		{
			name:  "single upgrade",
			input: "(1/1) Upgrading curl (8.5.0-r0 -> 8.6.0-r0)\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "curl", Version: "8.6.0-r0", Installed: true},
			},
		},
		{
			name: "multiple upgrades",
			input: `(1/3) Upgrading musl (1.2.4-r2 -> 1.2.5-r0)
(2/3) Upgrading openssl (3.1.4-r0 -> 3.2.0-r0)
(3/3) Upgrading curl (8.5.0-r0 -> 8.6.0-r0)
OK: 123 MiB in 45 packages
`,
			wantLen: 3,
			wantPkgs: []snack.Package{
				{Name: "musl", Version: "1.2.5-r0", Installed: true},
				{Name: "openssl", Version: "3.2.0-r0", Installed: true},
				{Name: "curl", Version: "8.6.0-r0", Installed: true},
			},
		},
		{
			name:    "non-upgrade lines only",
			input:   "Purging old package\nInstalling new-pkg\n",
			wantLen: 0,
		},
		{
			name:  "upgrade without version parens",
			input: "(1/1) Upgrading busybox\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "busybox", Version: "", Installed: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseUpgradeSimulation(tt.input)
			if len(pkgs) != tt.wantLen {
				t.Fatalf("expected %d packages, got %d", tt.wantLen, len(pkgs))
			}
			for i, want := range tt.wantPkgs {
				if i >= len(pkgs) {
					break
				}
				if pkgs[i].Name != want.Name {
					t.Errorf("pkg[%d].Name = %q, want %q", i, pkgs[i].Name, want.Name)
				}
				if pkgs[i].Version != want.Version {
					t.Errorf("pkg[%d].Version = %q, want %q", i, pkgs[i].Version, want.Version)
				}
				if pkgs[i].Installed != want.Installed {
					t.Errorf("pkg[%d].Installed = %v, want %v", i, pkgs[i].Installed, want.Installed)
				}
			}
		})
	}
}
