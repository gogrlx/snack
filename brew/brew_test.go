//go:build darwin || linux

package brew

import (
	"testing"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks
var (
	_ snack.Manager        = (*Brew)(nil)
	_ snack.VersionQuerier = (*Brew)(nil)
	_ snack.Cleaner        = (*Brew)(nil)
	_ snack.FileOwner      = (*Brew)(nil)
	_ snack.NameNormalizer = (*Brew)(nil)
)

func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Fatal("New() returned nil")
	}
}

func TestName(t *testing.T) {
	b := New()
	if b.Name() != "brew" {
		t.Errorf("Name() = %q, want %q", b.Name(), "brew")
	}
}

func TestParseBrewList(t *testing.T) {
	input := `git 2.43.0
go 1.21.6
vim 9.1.0
`
	pkgs := parseBrewList(input)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "git" || pkgs[0].Version != "2.43.0" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseBrewList_Empty(t *testing.T) {
	pkgs := parseBrewList("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseBrewList_SinglePackage(t *testing.T) {
	input := "curl 8.6.0\n"
	pkgs := parseBrewList(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" {
		t.Errorf("expected name=curl, got %q", pkgs[0].Name)
	}
}

func TestParseBrewSearch(t *testing.T) {
	input := `==> Formulae
git
git-absorb
git-annex
==> Casks
git-credential-manager
`
	pkgs := parseBrewSearch(input)
	if len(pkgs) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "git" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
}

func TestParseBrewSearch_Empty(t *testing.T) {
	pkgs := parseBrewSearch("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseBrewInfo(t *testing.T) {
	input := `{"formulae":[{"name":"git","full_name":"git","desc":"Distributed revision control system","versions":{"stable":"2.43.0"},"installed":[{"version":"2.43.0"}]}],"casks":[]}`
	pkg := parseBrewInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "git" {
		t.Errorf("expected name 'git', got %q", pkg.Name)
	}
	if pkg.Version != "2.43.0" {
		t.Errorf("unexpected version: %q", pkg.Version)
	}
	if !pkg.Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseBrewInfo_NotInstalled(t *testing.T) {
	input := `{"formulae":[{"name":"wget","full_name":"wget","desc":"Internet file retriever","versions":{"stable":"1.21"},"installed":[]}],"casks":[]}`
	pkg := parseBrewInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Installed {
		t.Error("expected Installed=false")
	}
}

func TestParseBrewInfo_Cask(t *testing.T) {
	input := `{"formulae":[],"casks":[{"token":"visual-studio-code","name":["Visual Studio Code"],"desc":"Open-source code editor","version":"1.85.0"}]}`
	pkg := parseBrewInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "visual-studio-code" {
		t.Errorf("expected name 'visual-studio-code', got %q", pkg.Name)
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"git", "git"},
		{"python@3.12", "python"},
		{"node@18", "node"},
		{"go", "go"},
		{"ruby@3.2", "ruby"},
		{"postgresql@16@beta", "postgresql@16"},
	}

	b := New()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := b.NormalizeName(tt.input)
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
		{"git", "git", ""},
		// Homebrew doesn't embed arch in names - @ is version suffix
		{"python@3.12", "python@3.12", ""},
		{"node@18", "node@18", ""},
	}

	b := New()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotName, gotArch := b.ParseArch(tt.input)
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
		{"FileOwnership", caps.FileOwnership, true},
		{"NameNormalize", caps.NameNormalize, true},
		// Homebrew does not support these
		{"Hold", caps.Hold, false},
		{"RepoManagement", caps.RepoManagement, false},
		{"KeyManagement", caps.KeyManagement, false},
		{"Groups", caps.Groups, false},
		{"DryRun", caps.DryRun, false},
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

func TestParseBrewInfoVersion(t *testing.T) {
	t.Run("formula", func(t *testing.T) {
		input := `{"formulae":[{"name":"git","full_name":"git","desc":"Distributed revision control system","versions":{"stable":"2.43.0"},"installed":[]}],"casks":[]}`
		ver := parseBrewInfoVersion(input)
		if ver != "2.43.0" {
			t.Errorf("expected '2.43.0', got %q", ver)
		}
	})

	t.Run("cask", func(t *testing.T) {
		input := `{"formulae":[],"casks":[{"token":"visual-studio-code","name":["Visual Studio Code"],"desc":"Open-source code editor","version":"1.85.0"}]}`
		ver := parseBrewInfoVersion(input)
		if ver != "1.85.0" {
			t.Errorf("expected '1.85.0', got %q", ver)
		}
	})

	t.Run("empty", func(t *testing.T) {
		ver := parseBrewInfoVersion("")
		if ver != "" {
			t.Errorf("expected empty, got %q", ver)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		ver := parseBrewInfoVersion("not json")
		if ver != "" {
			t.Errorf("expected empty, got %q", ver)
		}
	})

	t.Run("no formulae or casks", func(t *testing.T) {
		input := `{"formulae":[],"casks":[]}`
		ver := parseBrewInfoVersion(input)
		if ver != "" {
			t.Errorf("expected empty, got %q", ver)
		}
	})
}

func TestParseBrewOutdated(t *testing.T) {
	t.Run("formulae only", func(t *testing.T) {
		input := `{"formulae":[{"name":"git","installed_versions":["2.43.0"],"current_version":"2.44.0"},{"name":"go","installed_versions":["1.21.6"],"current_version":"1.22.0"}],"casks":[]}`
		pkgs := parseBrewOutdated(input)
		if len(pkgs) != 2 {
			t.Fatalf("expected 2 packages, got %d", len(pkgs))
		}
		if pkgs[0].Name != "git" || pkgs[0].Version != "2.44.0" {
			t.Errorf("unexpected first package: %+v", pkgs[0])
		}
		if !pkgs[0].Installed {
			t.Error("expected Installed=true")
		}
	})

	t.Run("casks only", func(t *testing.T) {
		input := `{"formulae":[],"casks":[{"name":"firefox","installed_versions":"119.0","current_version":"120.0"}]}`
		pkgs := parseBrewOutdated(input)
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "firefox" || pkgs[0].Version != "120.0" {
			t.Errorf("unexpected package: %+v", pkgs[0])
		}
	})

	t.Run("mixed", func(t *testing.T) {
		input := `{"formulae":[{"name":"git","installed_versions":["2.43.0"],"current_version":"2.44.0"}],"casks":[{"name":"firefox","installed_versions":"119.0","current_version":"120.0"}]}`
		pkgs := parseBrewOutdated(input)
		if len(pkgs) != 2 {
			t.Fatalf("expected 2 packages, got %d", len(pkgs))
		}
	})

	t.Run("empty", func(t *testing.T) {
		pkgs := parseBrewOutdated("")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("no outdated", func(t *testing.T) {
		input := `{"formulae":[],"casks":[]}`
		pkgs := parseBrewOutdated(input)
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		pkgs := parseBrewOutdated("not json")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})
}

func TestSemverCmp(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"less major", "1.0.0", "2.0.0", -1},
		{"greater major", "2.0.0", "1.0.0", 1},
		{"less minor", "1.2.3", "1.3.0", -1},
		{"less patch", "1.2.3", "1.2.4", -1},
		{"multi-digit", "1.10.0", "1.9.0", 1},
		{"short vs long equal", "1.0", "1.0.0", 0},
		{"short vs long less", "1.0", "1.0.1", -1},
		{"short vs long greater", "1.1", "1.0.9", 1},
		{"single component", "5", "3", 1},
		{"single equal", "3", "3", 0},
		{"empty vs empty", "", "", 0},
		{"empty vs version", "", "1.0", -1},
		{"version vs empty", "1.0", "", 1},
		{"four components", "1.2.3.4", "1.2.3.5", -1},
		{"different lengths", "1.0.0.0", "1.0.0", 0},
		{"real brew versions", "2.43.0", "2.44.0", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := semverCmp(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("semverCmp(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestParseVersionSuffix(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion string
	}{
		{"python@3.12", "python", "3.12"},
		{"node@18", "node", "18"},
		{"git", "git", ""},
		{"ruby@3.2", "ruby", "3.2"},
		{"postgresql@16@beta", "postgresql@16", "beta"},
		{"", "", ""},
		{"@3.12", "@3.12", ""}, // @ at position 0, LastIndex returns 0 which is not > 0
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotName, gotVer := parseVersionSuffix(tt.input)
			if gotName != tt.wantName || gotVer != tt.wantVersion {
				t.Errorf("parseVersionSuffix(%q) = (%q, %q), want (%q, %q)",
					tt.input, gotName, gotVer, tt.wantName, tt.wantVersion)
			}
		})
	}
}

func TestParseBrewInfo_Empty(t *testing.T) {
	pkg := parseBrewInfo("")
	if pkg != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseBrewInfo_InvalidJSON(t *testing.T) {
	pkg := parseBrewInfo("not json")
	if pkg != nil {
		t.Error("expected nil for invalid JSON")
	}
}

func TestParseBrewInfo_NoFormulaeOrCasks(t *testing.T) {
	input := `{"formulae":[],"casks":[]}`
	pkg := parseBrewInfo(input)
	if pkg != nil {
		t.Error("expected nil when no formulae or casks")
	}
}

func TestParseBrewSearch_HeadersOnly(t *testing.T) {
	input := `==> Formulae

==> Casks
`
	pkgs := parseBrewSearch(input)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseBrewSearch_MultiplePerLine(t *testing.T) {
	input := "git  go  vim  curl\n"
	pkgs := parseBrewSearch(input)
	if len(pkgs) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(pkgs))
	}
	names := []string{"git", "go", "vim", "curl"}
	for i, want := range names {
		if pkgs[i].Name != want {
			t.Errorf("pkg[%d].Name = %q, want %q", i, pkgs[i].Name, want)
		}
	}
}

func TestParseBrewList_NameOnly(t *testing.T) {
	input := "git\ncurl\n"
	pkgs := parseBrewList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Version != "" {
		t.Errorf("expected empty version, got %q", pkgs[0].Version)
	}
}

func TestParseBrewList_WhitespaceLines(t *testing.T) {
	input := "  \n\n  git 2.43.0\n  \n"
	pkgs := parseBrewList(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
}

func TestInterfaceNonCompliance(t *testing.T) {
	var m snack.Manager = New()
	if _, ok := m.(snack.Holder); ok {
		t.Error("Brew should not implement Holder")
	}
	if _, ok := m.(snack.RepoManager); ok {
		t.Error("Brew should not implement RepoManager")
	}
	if _, ok := m.(snack.KeyManager); ok {
		t.Error("Brew should not implement KeyManager")
	}
	if _, ok := m.(snack.Grouper); ok {
		t.Error("Brew should not implement Grouper")
	}
}
