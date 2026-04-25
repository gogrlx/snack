package snap

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseSnapList(t *testing.T) {
	input := `Name      Version    Rev    Tracking       Publisher   Notes
core22    20240111   1122   latest/stable  canonical✓  base
firefox   131.0      4647   latest/stable  mozilla✓    -
`
	pkgs := parseSnapList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "core22" || pkgs[0].Version != "20240111" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[1].Name != "firefox" || pkgs[1].Version != "131.0" {
		t.Errorf("unexpected second package: %+v", pkgs[1])
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseSnapListEmpty(t *testing.T) {
	input := `Name  Version  Rev  Tracking  Publisher  Notes
`
	pkgs := parseSnapList(input)
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseSnapListEdgeCases(t *testing.T) {
	t.Run("single entry", func(t *testing.T) {
		input := `Name      Version    Rev    Tracking       Publisher   Notes
core22    20240111   1122   latest/stable  canonical✓  base
`
		pkgs := parseSnapList(input)
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "core22" {
			t.Errorf("expected core22, got %q", pkgs[0].Name)
		}
	})

	t.Run("header only no trailing newline", func(t *testing.T) {
		input := "Name  Version  Rev  Tracking  Publisher  Notes"
		pkgs := parseSnapList(input)
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("single field line skipped", func(t *testing.T) {
		input := "Name  Version  Rev  Tracking  Publisher  Notes\nsinglefield\n"
		pkgs := parseSnapList(input)
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages (need >=2 fields), got %d", len(pkgs))
		}
	})

	t.Run("extra whitespace lines", func(t *testing.T) {
		input := "Name  Version  Rev  Tracking  Publisher  Notes\n  \n\n"
		pkgs := parseSnapList(input)
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("many columns", func(t *testing.T) {
		input := "Name  Version  Rev  Tracking  Publisher  Notes\nsnap1  2.0  100  latest/stable  pub  note  extra  more\n"
		pkgs := parseSnapList(input)
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "snap1" || pkgs[0].Version != "2.0" {
			t.Errorf("unexpected package: %+v", pkgs[0])
		}
	})
}

func TestParseSnapFind(t *testing.T) {
	input := `Name      Version   Publisher    Notes  Summary
firefox   131.0     mozilla✓     -      Mozilla Firefox web browser
chromium  129.0     nickvdp      -      Chromium web browser
`
	pkgs := parseSnapFind(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "firefox" || pkgs[0].Version != "131.0" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[0].Description != "Mozilla Firefox web browser" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
}

func TestParseSnapFindEdgeCases(t *testing.T) {
	t.Run("single result", func(t *testing.T) {
		input := "Name  Version  Publisher  Notes  Summary\nfirefox  131.0  mozilla  -  Web browser\n"
		pkgs := parseSnapFind(input)
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "firefox" {
			t.Errorf("expected firefox, got %q", pkgs[0].Name)
		}
		if pkgs[0].Description != "Web browser" {
			t.Errorf("expected 'Web browser', got %q", pkgs[0].Description)
		}
	})

	t.Run("header only", func(t *testing.T) {
		input := "Name  Version  Publisher  Notes  Summary\n"
		pkgs := parseSnapFind(input)
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("too few fields skipped", func(t *testing.T) {
		input := "Name  Version  Publisher  Notes  Summary\nfoo  1.0  pub\n"
		pkgs := parseSnapFind(input)
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages (need >=4 fields), got %d", len(pkgs))
		}
	})

	t.Run("exactly four fields no summary", func(t *testing.T) {
		input := "Name  Version  Publisher  Notes  Summary\nfoo  1.0  pub  note\n"
		pkgs := parseSnapFind(input)
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Description != "" {
			t.Errorf("expected empty description, got %q", pkgs[0].Description)
		}
	})

	t.Run("multi-word summary", func(t *testing.T) {
		input := "Name  Version  Publisher  Notes  Summary\nmysnap  2.0  me  -  A very long description with many words\n"
		pkgs := parseSnapFind(input)
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Description != "A very long description with many words" {
			t.Errorf("unexpected description: %q", pkgs[0].Description)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		pkgs := parseSnapFind("")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})
}

func TestParseSnapInfo(t *testing.T) {
	input := `name:      firefox
summary:   Mozilla Firefox web browser
publisher: Mozilla✓ (mozilla✓)
snap-id:   3wdHCAVyZEmYsCMFDE9qt92UV8rC8Wdk
installed: 131.0 (4647) 283MB -
`
	pkg := parseSnapInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "firefox" {
		t.Errorf("expected name 'firefox', got %q", pkg.Name)
	}
	if pkg.Version != "131.0" {
		t.Errorf("expected version '131.0', got %q", pkg.Version)
	}
	if pkg.Description != "Mozilla Firefox web browser" {
		t.Errorf("unexpected description: %q", pkg.Description)
	}
	if !pkg.Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseSnapInfoEmpty(t *testing.T) {
	pkg := parseSnapInfo("")
	if pkg != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseSnapInfoEdgeCases(t *testing.T) {
	t.Run("not installed snap", func(t *testing.T) {
		input := `name:      hello-world
summary:   A simple hello world snap
publisher: Canonical✓ (canonical✓)
snap-id:   buPKUD3TKqCOgLEjjHx5kSiCpIs5cMuQ
`
		pkg := parseSnapInfo(input)
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Name != "hello-world" {
			t.Errorf("expected hello-world, got %q", pkg.Name)
		}
		if pkg.Installed {
			t.Error("expected Installed=false for snap without installed field")
		}
		if pkg.Version != "" {
			t.Errorf("expected empty version, got %q", pkg.Version)
		}
	})

	t.Run("name only", func(t *testing.T) {
		pkg := parseSnapInfo("name: test\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Name != "test" {
			t.Errorf("expected test, got %q", pkg.Name)
		}
	})

	t.Run("no name returns nil", func(t *testing.T) {
		pkg := parseSnapInfo("summary: something\ninstalled: 1.0 (1) 10MB\n")
		if pkg != nil {
			t.Error("expected nil when no name field")
		}
	})

	t.Run("no colon lines ignored", func(t *testing.T) {
		input := "name: mysnap\nrandom text without colon\nsummary: A snap\n"
		pkg := parseSnapInfo(input)
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Description != "A snap" {
			t.Errorf("unexpected description: %q", pkg.Description)
		}
	})
}

func TestParseSnapInfoVersion(t *testing.T) {
	input := `name:      firefox
channels:
  latest/stable:    131.0       2024-10-01 (4647) 283MB -
  latest/candidate: 132.0b5     2024-10-05 (4650) 285MB -
  latest/beta:      132.0b5     2024-10-05 (4650) 285MB -
  latest/edge:      133.0a1     2024-10-06 (4655) 290MB -
`
	ver := parseSnapInfoVersion(input)
	if ver != "131.0" {
		t.Errorf("expected '131.0', got %q", ver)
	}
}

func TestParseSnapInfoVersionMissing(t *testing.T) {
	ver := parseSnapInfoVersion("name: test\n")
	if ver != "" {
		t.Errorf("expected empty, got %q", ver)
	}
}

func TestParseSnapInfoVersionEdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		ver := parseSnapInfoVersion("")
		if ver != "" {
			t.Errorf("expected empty, got %q", ver)
		}
	})

	t.Run("stable channel with dashes", func(t *testing.T) {
		input := "  latest/stable:    2024.01.15    2024-01-15 (100) 50MB -\n"
		ver := parseSnapInfoVersion(input)
		if ver != "2024.01.15" {
			t.Errorf("expected 2024.01.15, got %q", ver)
		}
	})

	t.Run("closed channel marked --", func(t *testing.T) {
		input := "  latest/stable:    --\n"
		ver := parseSnapInfoVersion(input)
		if ver != "" {
			t.Errorf("expected empty for closed channel, got %q", ver)
		}
	})

	t.Run("closed channel marked ^", func(t *testing.T) {
		input := "  latest/stable:    ^    \n"
		ver := parseSnapInfoVersion(input)
		if ver != "" {
			t.Errorf("expected empty for ^ channel, got %q", ver)
		}
	})
}

func TestParseSnapRefreshList(t *testing.T) {
	input := `Name      Version   Rev   Publisher    Notes
firefox   132.0     4650  mozilla✓     -
`
	pkgs := parseSnapRefreshList(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "firefox" || pkgs[0].Version != "132.0" {
		t.Errorf("unexpected package: %+v", pkgs[0])
	}
}

func TestParseSnapRefreshListUpToDate(t *testing.T) {
	input := `Name  Version  Rev  Publisher  Notes
All snaps up to date.
`
	pkgs := parseSnapRefreshList(input)
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseSnapRefreshListEdgeCases(t *testing.T) {
	t.Run("multiple upgrades", func(t *testing.T) {
		input := "Name  Version  Rev  Publisher  Notes\nfoo  2.0  10  pub  -\nbar  3.0  20  pub  -\n"
		pkgs := parseSnapRefreshList(input)
		if len(pkgs) != 2 {
			t.Fatalf("expected 2 packages, got %d", len(pkgs))
		}
		if pkgs[0].Name != "foo" || pkgs[1].Name != "bar" {
			t.Errorf("unexpected packages: %+v, %+v", pkgs[0], pkgs[1])
		}
	})

	t.Run("header only", func(t *testing.T) {
		input := "Name  Version  Rev  Publisher  Notes\n"
		pkgs := parseSnapRefreshList(input)
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
		{"less patch", "1.2.3", "1.2.4", -1},
		{"multi-digit minor", "1.10.0", "1.9.0", 1},
		{"short vs long equal", "1.0", "1.0.0", 0},
		{"real versions", "131.0", "132.0", -1},
		// Edge cases
		{"single component", "5", "3", 1},
		{"single equal", "3", "3", 0},
		{"empty vs empty", "", "", 0},
		{"empty vs version", "", "1.0", -1},
		{"version vs empty", "1.0", "", 1},
		{"non-numeric suffix", "1.0.0-beta", "1.0.0-rc1", 0},
		{"four components", "1.2.3.4", "1.2.3.5", -1},
		{"different lengths padded", "1.0.0.0", "1.0.0", 0},
		{"short less", "1.0", "1.0.1", -1},
		{"short greater", "1.1", "1.0.9", 1},
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
	var _ snack.Manager = (*Snap)(nil)
	var _ snack.VersionQuerier = (*Snap)(nil)
	var _ snack.Cleaner = (*Snap)(nil)
	var _ snack.PackageUpgrader = (*Snap)(nil)
}

// Compile-time interface checks
var (
	_ snack.Cleaner         = (*Snap)(nil)
	_ snack.PackageUpgrader = (*Snap)(nil)
)

func TestCapabilities(t *testing.T) {
	caps := snack.GetCapabilities(New())
	if !caps.VersionQuery {
		t.Error("expected VersionQuery=true")
	}
	if !caps.Clean {
		t.Error("expected Clean=true")
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
	if caps.RepoManagement {
		t.Error("expected RepoManagement=false")
	}
	if caps.KeyManagement {
		t.Error("expected KeyManagement=false")
	}
	if caps.Groups {
		t.Error("expected Groups=false")
	}
	if !caps.NameNormalize {
		t.Error("expected NameNormalize=true")
	}
	if !caps.PackageUpgrade {
		t.Error("expected PackageUpgrade=true")
	}
}

func TestName(t *testing.T) {
	s := New()
	if s.Name() != "snap" {
		t.Errorf("Name() = %q, want %q", s.Name(), "snap")
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"firefox", "firefox"},
		{"hello-world", "hello-world"},
		{"snap.with.dots", "snap.with.dots"},
		{"  keep-whitespace  ", "  keep-whitespace  "},
		{"", ""},
	}

	s := New()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := s.NormalizeName(tt.input); got != tt.want {
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
		{"firefox", "firefox", ""},
		{"hello-world", "hello-world", ""},
		{"snap.with.dots", "snap.with.dots", ""},
		{"", "", ""},
	}

	s := New()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotName, gotArch := s.ParseArch(tt.input)
			if gotName != tt.wantName || gotArch != tt.wantArch {
				t.Errorf("ParseArch(%q) = (%q, %q), want (%q, %q)",
					tt.input, gotName, gotArch, tt.wantName, tt.wantArch)
			}
		})
	}
}
