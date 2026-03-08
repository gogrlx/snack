package rpm

import (
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"nginx", "nginx"},
		{"nginx.x86_64", "nginx"},
		{"curl.noarch", "curl"},
		{"kernel.aarch64", "kernel"},
		{"bash.i686", "bash"},
		{"glibc.i386", "glibc"},
		{"libfoo.armv7hl", "libfoo"},
		{"module.ppc64le", "module"},
		{"app.s390x", "app"},
		{"source.src", "source"},
		{"nodot", "nodot"},
		{"", ""},
		{"pkg.unknown", "pkg.unknown"},
		{"multi.dot.x86_64", "multi.dot"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseArchSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantArch string
	}{
		{"x86_64", "nginx.x86_64", "nginx", "x86_64"},
		{"noarch", "bash.noarch", "bash", "noarch"},
		{"aarch64", "kernel.aarch64", "kernel", "aarch64"},
		{"i686", "glibc.i686", "glibc", "i686"},
		{"i386", "compat.i386", "compat", "i386"},
		{"armv7hl", "lib.armv7hl", "lib", "armv7hl"},
		{"ppc64le", "app.ppc64le", "app", "ppc64le"},
		{"s390x", "z.s390x", "z", "s390x"},
		{"src", "pkg.src", "pkg", "src"},
		{"no dot", "curl", "curl", ""},
		{"unknown arch", "pkg.foobar", "pkg.foobar", ""},
		{"empty", "", "", ""},
		{"multiple dots", "a.b.x86_64", "a.b", "x86_64"},
		{"dot but not arch", "libfoo.so", "libfoo.so", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotArch := parseArchSuffix(tt.input)
			if gotName != tt.wantName || gotArch != tt.wantArch {
				t.Errorf("parseArchSuffix(%q) = (%q, %q), want (%q, %q)",
					tt.input, gotName, gotArch, tt.wantName, tt.wantArch)
			}
		})
	}
}

func TestParseList(t *testing.T) {
	input := "bash\t5.2.15-3.fc38\tThe GNU Bourne Again shell\n" +
		"curl\t8.0.1-1.fc38\tA utility for getting files from remote servers\n"
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || pkgs[0].Version != "5.2.15-3.fc38" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[0].Description != "The GNU Bourne Again shell" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseListEdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		pkgs := parseList("")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		pkgs := parseList("  \n\n  \n")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages, got %d", len(pkgs))
		}
	})

	t.Run("single entry no description", func(t *testing.T) {
		pkgs := parseList("vim\t9.0.1\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "vim" || pkgs[0].Version != "9.0.1" {
			t.Errorf("unexpected package: %+v", pkgs[0])
		}
		if pkgs[0].Description != "" {
			t.Errorf("expected empty description, got %q", pkgs[0].Description)
		}
	})

	t.Run("single field line skipped", func(t *testing.T) {
		pkgs := parseList("justname\n")
		if len(pkgs) != 0 {
			t.Errorf("expected 0 packages (need >=2 tab fields), got %d", len(pkgs))
		}
	})

	t.Run("description with tabs", func(t *testing.T) {
		pkgs := parseList("pkg\t1.0\tA description\twith tabs\n")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		// SplitN with 3 means the third part includes everything after the second tab
		if pkgs[0].Description != "A description\twith tabs" {
			t.Errorf("unexpected description: %q", pkgs[0].Description)
		}
	})
}

func TestParseInfo(t *testing.T) {
	input := `Name        : curl
Version     : 8.0.1
Release     : 1.fc38
Architecture: x86_64
Summary     : A utility for getting files from remote servers
`
	pkg := parseInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "curl" {
		t.Errorf("expected name 'curl', got %q", pkg.Name)
	}
	if pkg.Version != "8.0.1-1.fc38" {
		t.Errorf("expected version '8.0.1-1.fc38', got %q", pkg.Version)
	}
	if pkg.Arch != "x86_64" {
		t.Errorf("expected arch 'x86_64', got %q", pkg.Arch)
	}
	if pkg.Description != "A utility for getting files from remote servers" {
		t.Errorf("unexpected description: %q", pkg.Description)
	}
}

func TestParseInfoEdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		pkg := parseInfo("")
		if pkg != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("name only", func(t *testing.T) {
		pkg := parseInfo("Name : bash\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Name != "bash" {
			t.Errorf("expected bash, got %q", pkg.Name)
		}
	})

	t.Run("no name returns nil", func(t *testing.T) {
		pkg := parseInfo("Version : 1.0\nArch : x86_64\n")
		if pkg != nil {
			t.Error("expected nil when no Name field")
		}
	})

	t.Run("version without release", func(t *testing.T) {
		pkg := parseInfo("Name : test\nVersion : 2.5\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Version != "2.5" {
			t.Errorf("expected version '2.5', got %q", pkg.Version)
		}
	})

	t.Run("release without version", func(t *testing.T) {
		// Release only appends if version is non-empty
		pkg := parseInfo("Name : test\nRelease : 3.el9\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Version != "" {
			t.Errorf("expected empty version (release alone shouldn't set it), got %q", pkg.Version)
		}
	})

	t.Run("arch key variant", func(t *testing.T) {
		pkg := parseInfo("Name : test\nArch : aarch64\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Arch != "aarch64" {
			t.Errorf("expected aarch64, got %q", pkg.Arch)
		}
	})

	t.Run("no colon lines ignored", func(t *testing.T) {
		pkg := parseInfo("Name : test\nrandom line\nSummary : A tool\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Description != "A tool" {
			t.Errorf("unexpected description: %q", pkg.Description)
		}
	})

	t.Run("value with colons", func(t *testing.T) {
		pkg := parseInfo("Name : myapp\nSummary : A tool: does things: well\n")
		if pkg == nil {
			t.Fatal("expected non-nil")
		}
		if pkg.Description != "A tool: does things: well" {
			t.Errorf("unexpected description: %q", pkg.Description)
		}
	})
}
