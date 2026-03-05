package apt

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseList_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"whitespace_only", "  \n  \n  ", 0},
		{"single_tab_field", "bash", 0},       // needs at least 2 tab-separated fields
		{"no_description", "bash\t5.2-1", 1},  // 2 fields is OK
		{"with_description", "bash\t5.2-1\tGNU Bourne Again SHell", 1},
		{"blank_lines_mixed", "\nbash\t5.2-1\n\ncurl\t7.88\n\n", 2},
		{"trailing_newline", "bash\t5.2-1\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseList(tt.input)
			if len(pkgs) != tt.want {
				t.Errorf("parseList() returned %d packages, want %d", len(pkgs), tt.want)
			}
		})
	}

	// Verify fields are populated correctly
	t.Run("field_values", func(t *testing.T) {
		input := "bash\t5.2-1\tGNU Bourne Again SHell"
		pkgs := parseList(input)
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		p := pkgs[0]
		if p.Name != "bash" {
			t.Errorf("Name = %q, want %q", p.Name, "bash")
		}
		if p.Version != "5.2-1" {
			t.Errorf("Version = %q, want %q", p.Version, "5.2-1")
		}
		if p.Description != "GNU Bourne Again SHell" {
			t.Errorf("Description = %q, want %q", p.Description, "GNU Bourne Again SHell")
		}
		if !p.Installed {
			t.Error("expected Installed=true")
		}
	})

	t.Run("no_description_fields", func(t *testing.T) {
		pkgs := parseList("vim\t9.0")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Description != "" {
			t.Errorf("Description = %q, want empty", pkgs[0].Description)
		}
	})

	t.Run("description_with_tabs", func(t *testing.T) {
		// Third field captures everything after the second tab
		pkgs := parseList("pkg\t1.0\tdesc\twith\ttabs")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Description != "desc\twith\ttabs" {
			t.Errorf("Description = %q, want %q", pkgs[0].Description, "desc\twith\ttabs")
		}
	})
}

func TestParseSearch_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"whitespace_only", "  \n  ", 0},
		{"no_dash_separator", "vim", 1}, // still parses, just no description
		{"normal", "vim - Vi IMproved", 1},
		{"blank_lines", "\nvim - Vi IMproved\n\nnano - small editor\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseSearch(tt.input)
			if len(pkgs) != tt.want {
				t.Errorf("parseSearch() returned %d packages, want %d", len(pkgs), tt.want)
			}
		})
	}

	t.Run("no_description", func(t *testing.T) {
		pkgs := parseSearch("vim")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "vim" {
			t.Errorf("Name = %q, want %q", pkgs[0].Name, "vim")
		}
		if pkgs[0].Description != "" {
			t.Errorf("Description = %q, want empty", pkgs[0].Description)
		}
	})

	t.Run("description_with_dashes", func(t *testing.T) {
		// Only splits on first " - "
		pkgs := parseSearch("gcc - GNU C Compiler - version 12")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "gcc" {
			t.Errorf("Name = %q, want %q", pkgs[0].Name, "gcc")
		}
		if pkgs[0].Description != "GNU C Compiler - version 12" {
			t.Errorf("Description = %q, want %q", pkgs[0].Description, "GNU C Compiler - version 12")
		}
	})

	t.Run("whitespace_trimming", func(t *testing.T) {
		pkgs := parseSearch("  vim  -  Vi IMproved  ")
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}
		if pkgs[0].Name != "vim" {
			t.Errorf("Name = %q, want %q", pkgs[0].Name, "vim")
		}
		if pkgs[0].Description != "Vi IMproved" {
			t.Errorf("Description = %q, want %q", pkgs[0].Description, "Vi IMproved")
		}
	})
}

func TestParseInfo_EdgeCases(t *testing.T) {
	t.Run("all_fields", func(t *testing.T) {
		input := "Package: bash\nVersion: 5.2-1\nArchitecture: amd64\nDescription: GNU Bourne Again SHell\n"
		p, err := parseInfo(input)
		if err != nil {
			t.Fatal(err)
		}
		if p.Name != "bash" || p.Version != "5.2-1" || p.Arch != "amd64" || p.Description != "GNU Bourne Again SHell" {
			t.Errorf("unexpected package: %+v", p)
		}
	})

	t.Run("empty_returns_not_found", func(t *testing.T) {
		_, err := parseInfo("")
		if err != snack.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("no_package_field", func(t *testing.T) {
		_, err := parseInfo("Version: 1.0\nArchitecture: amd64\n")
		if err != snack.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("extra_fields_ignored", func(t *testing.T) {
		input := "Package: curl\nMaintainer: someone\nVersion: 7.88\nPriority: optional\n"
		p, err := parseInfo(input)
		if err != nil {
			t.Fatal(err)
		}
		if p.Name != "curl" || p.Version != "7.88" {
			t.Errorf("unexpected: %+v", p)
		}
	})

	t.Run("lines_without_colon", func(t *testing.T) {
		input := "Package: vim\nsome continuation line\nVersion: 9.0\n"
		p, err := parseInfo(input)
		if err != nil {
			t.Fatal(err)
		}
		if p.Name != "vim" || p.Version != "9.0" {
			t.Errorf("unexpected: %+v", p)
		}
	})

	t.Run("version_with_epoch", func(t *testing.T) {
		input := "Package: systemd\nVersion: 1:252-2\nArchitecture: amd64\n"
		p, err := parseInfo(input)
		if err != nil {
			t.Fatal(err)
		}
		if p.Version != "1:252-2" {
			t.Errorf("Version = %q, want %q", p.Version, "1:252-2")
		}
	})
}
