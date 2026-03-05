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

func TestParseListEdgeCases(t *testing.T) {
	t.Run("single entry", func(t *testing.T) {
		pkgs := parseList("Firefox\torg.mozilla.Firefox\t131.0\tflathub\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "Firefox" {
			t.Errorf("expected Firefox, got %q", pkgs[0].Name)
		}
	})

	t.Run("two fields only", func(t *testing.T) {
		pkgs := parseList("Firefox\torg.mozilla.Firefox\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "Firefox" {
			t.Errorf("expected Firefox, got %q", pkgs[0].Name)
		}
		if pkgs[0].Version != "" {
			t.Errorf("expected empty version, got %q", pkgs[0].Version)
		}
		if pkgs[0].Repository != "" {
			t.Errorf("expected empty repository, got %q", pkgs[0].Repository)
		}
	})

	t.Run("single field skipped", func(t *testing.T) {
		pkgs := parseList("Firefox\n")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages (needs >=2 fields), got %d", len(pkgs))
		}
	})

	t.Run("extra fields", func(t *testing.T) {
		pkgs := parseList("Firefox\torg.mozilla.Firefox\t131.0\tflathub\textra\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Repository != "flathub" {
			t.Errorf("expected flathub, got %q", pkgs[0].Repository)
		}
	})

	t.Run("whitespace only lines", func(t *testing.T) {
		pkgs := parseList("   \n\t\n  \n")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("three fields no repo", func(t *testing.T) {
		pkgs := parseList("GIMP\torg.gimp.GIMP\t2.10.38\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Version != "2.10.38" {
			t.Errorf("expected version 2.10.38, got %q", pkgs[0].Version)
		}
		if pkgs[0].Repository != "" {
			t.Errorf("expected empty repository, got %q", pkgs[0].Repository)
		}
	})
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

func TestParseSearchEdgeCases(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		pkgs := parseSearch("")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("single field line skipped", func(t *testing.T) {
		pkgs := parseSearch("JustAName\n")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages (needs >=2 fields), got %d", len(pkgs))
		}
	})

	t.Run("not installed result", func(t *testing.T) {
		pkgs := parseSearch("VLC\torg.videolan.VLC\t3.0.20\tflathub\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Installed {
			t.Error("search results should not be marked installed")
		}
	})
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

func TestParseInfoEdgeCases(t *testing.T) {
	t.Run("name only", func(t *testing.T) {
		pkg := parseInfo("Name: VLC\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Name != "VLC" {
			t.Errorf("expected VLC, got %q", pkg.Name)
		}
	})

	t.Run("no name returns nil", func(t *testing.T) {
		pkg := parseInfo("Version: 1.0\nArch: x86_64\n")
		if pkg != nil {
			t.Error("expected nil when no Name field")
		}
	})

	t.Run("no colon lines ignored", func(t *testing.T) {
		pkg := parseInfo("Name: Test\nsome random line without colon\nVersion: 2.0\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Version != "2.0" {
			t.Errorf("expected version 2.0, got %q", pkg.Version)
		}
	})

	t.Run("value with colons", func(t *testing.T) {
		pkg := parseInfo("Name: MyApp\nDescription: A tool: does things: well\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		// parseInfo uses strings.Index for first colon
		if pkg.Description != "A tool: does things: well" {
			t.Errorf("unexpected description: %q", pkg.Description)
		}
	})
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

func TestParseRemotesEdgeCases(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		repos := parseRemotes("")
		if len(repos) != 0 {
			t.Errorf("expected 0 repos, got %d", len(repos))
		}
	})

	t.Run("single enabled remote", func(t *testing.T) {
		repos := parseRemotes("flathub\thttps://dl.flathub.org/repo/\t\n")
		if len(repos) != 1 {
			t.Fatalf("expected 1 repo, got %d", len(repos))
		}
		if !repos[0].Enabled {
			t.Error("expected enabled")
		}
		if repos[0].Name != "flathub" {
			t.Errorf("expected Name=flathub, got %q", repos[0].Name)
		}
	})

	t.Run("single disabled remote", func(t *testing.T) {
		repos := parseRemotes("test-remote\thttps://example.com/\tdisabled\n")
		if len(repos) != 1 {
			t.Fatalf("expected 1 repo, got %d", len(repos))
		}
		if repos[0].Enabled {
			t.Error("expected disabled")
		}
	})

	t.Run("no URL field", func(t *testing.T) {
		repos := parseRemotes("myremote\n")
		if len(repos) != 1 {
			t.Fatalf("expected 1 repo, got %d", len(repos))
		}
		if repos[0].ID != "myremote" {
			t.Errorf("expected myremote, got %q", repos[0].ID)
		}
		if repos[0].URL != "" {
			t.Errorf("expected empty URL, got %q", repos[0].URL)
		}
		if !repos[0].Enabled {
			t.Error("expected enabled by default")
		}
	})

	t.Run("whitespace lines ignored", func(t *testing.T) {
		repos := parseRemotes("  \n\n  \n")
		if len(repos) != 0 {
			t.Errorf("expected 0 repos, got %d", len(repos))
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
		{"non-numeric suffix", "1.0.0-beta", "1.0.0-rc1", 0},
		{"pre-release stripped", "1.0.0beta2", "1.0.0rc1", 0},
		{"four components", "1.2.3.4", "1.2.3.5", -1},
		{"different lengths", "1.0.0.0", "1.0.0", 0},
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

func TestStripNonNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123", "123"},
		{"123abc", "123"},
		{"abc", ""},
		{"0beta", "0"},
		{"", ""},
		{"42-rc1", "42"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripNonNumeric(tt.input)
			if got != tt.want {
				t.Errorf("stripNonNumeric(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ snack.Manager = (*Flatpak)(nil)
	var _ snack.Cleaner = (*Flatpak)(nil)
	var _ snack.RepoManager = (*Flatpak)(nil)
	var _ snack.VersionQuerier = (*Flatpak)(nil)
	var _ snack.PackageUpgrader = (*Flatpak)(nil)
}

// Compile-time interface checks in test file
var (
	_ snack.VersionQuerier  = (*Flatpak)(nil)
	_ snack.PackageUpgrader = (*Flatpak)(nil)
)

func TestCapabilities(t *testing.T) {
	caps := snack.GetCapabilities(New())
	if !caps.Clean {
		t.Error("expected Clean=true")
	}
	if !caps.RepoManagement {
		t.Error("expected RepoManagement=true")
	}
	if !caps.VersionQuery {
		t.Error("expected VersionQuery=true")
	}
	// Should be false
	if caps.FileOwnership {
		t.Error("expected FileOwnership=false")
	}
	if caps.DryRun {
		t.Error("expected DryRun=false")
	}
	if caps.Hold {
		t.Error("expected Hold=false")
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

func TestName(t *testing.T) {
	f := New()
	if f.Name() != "flatpak" {
		t.Errorf("Name() = %q, want %q", f.Name(), "flatpak")
	}
}
