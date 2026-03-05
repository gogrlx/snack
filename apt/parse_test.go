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

// --- New parse function tests ---

func TestParsePolicyCandidate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "normal_policy_output",
			input: `bash:
  Installed: 5.2-1
  Candidate: 5.2-2
  Version table:
     5.2-2 500
        500 http://archive.ubuntu.com/ubuntu jammy/main amd64 Packages
 *** 5.2-1 100
        100 /var/lib/dpkg/status`,
			want: "5.2-2",
		},
		{
			name: "candidate_none",
			input: `virtual-pkg:
  Installed: (none)
  Candidate: (none)
  Version table:`,
			want: "",
		},
		{
			name:  "empty_input",
			input: "",
			want:  "",
		},
		{
			name: "installed_equals_candidate",
			input: `curl:
  Installed: 7.88.1-10+deb12u4
  Candidate: 7.88.1-10+deb12u4
  Version table:
 *** 7.88.1-10+deb12u4 500`,
			want: "7.88.1-10+deb12u4",
		},
		{
			name: "epoch_version",
			input: `systemd:
  Installed: 1:252-2
  Candidate: 1:252-3
  Version table:`,
			want: "1:252-3",
		},
		{
			name:  "no_candidate_line",
			input: "bash:\n  Installed: 5.2-1\n",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePolicyCandidate(tt.input)
			if got != tt.want {
				t.Errorf("parsePolicyCandidate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePolicyInstalled(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "normal",
			input: `bash:
  Installed: 5.2-1
  Candidate: 5.2-2`,
			want: "5.2-1",
		},
		{
			name: "not_installed",
			input: `foo:
  Installed: (none)
  Candidate: 1.0`,
			want: "",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name: "epoch_version",
			input: `systemd:
  Installed: 1:252-2
  Candidate: 1:252-3`,
			want: "1:252-2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePolicyInstalled(tt.input)
			if got != tt.want {
				t.Errorf("parsePolicyInstalled() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseUpgradeSimulation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []snack.Package
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name: "single_upgrade",
			input: `Reading package lists...
Building dependency tree...
Reading state information...
The following packages will be upgraded:
  bash
1 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.
Inst bash [5.2-1] (5.2-2 Ubuntu:22.04/jammy [amd64])
Conf bash (5.2-2 Ubuntu:22.04/jammy [amd64])`,
			want: []snack.Package{
				{Name: "bash", Version: "5.2-2", Installed: true},
			},
		},
		{
			name: "multiple_upgrades",
			input: `Inst bash [5.2-1] (5.2-2 Ubuntu:22.04/jammy [amd64])
Inst curl [7.88.0] (7.88.1 Ubuntu:22.04/jammy [amd64])
Inst systemd [1:252-1] (1:252-2 Ubuntu:22.04/jammy [amd64])
Conf bash (5.2-2 Ubuntu:22.04/jammy [amd64])
Conf curl (7.88.1 Ubuntu:22.04/jammy [amd64])
Conf systemd (1:252-2 Ubuntu:22.04/jammy [amd64])`,
			want: []snack.Package{
				{Name: "bash", Version: "5.2-2", Installed: true},
				{Name: "curl", Version: "7.88.1", Installed: true},
				{Name: "systemd", Version: "1:252-2", Installed: true},
			},
		},
		{
			name:  "no_inst_lines",
			input: "Reading package lists...\nBuilding dependency tree...\n0 upgraded, 0 newly installed.\n",
			want:  nil,
		},
		{
			name:  "inst_without_parens",
			input: "Inst bash no-parens-here\n",
			want:  nil,
		},
		{
			name:  "inst_with_empty_parens",
			input: "Inst bash [5.2-1] ()\n",
			want:  nil,
		},
		{
			name:  "conf_lines_ignored",
			input: "Conf bash (5.2-2 Ubuntu:22.04/jammy [amd64])\n",
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUpgradeSimulation(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseUpgradeSimulation() returned %d packages, want %d", len(got), len(tt.want))
			}
			for i, g := range got {
				w := tt.want[i]
				if g.Name != w.Name || g.Version != w.Version || g.Installed != w.Installed {
					t.Errorf("package[%d] = %+v, want %+v", i, g, w)
				}
			}
		})
	}
}

func TestParseHoldList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"whitespace_only", "  \n  \n", nil},
		{"single_package", "bash\n", []string{"bash"}},
		{"multiple_packages", "bash\ncurl\nnginx\n", []string{"bash", "curl", "nginx"}},
		{"blank_lines_mixed", "\nbash\n\ncurl\n\n", []string{"bash", "curl"}},
		{"trailing_whitespace", "  bash  \n  curl  \n", []string{"bash", "curl"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseHoldList(tt.input)
			var names []string
			for _, p := range pkgs {
				names = append(names, p.Name)
				if !p.Installed {
					t.Errorf("expected Installed=true for %q", p.Name)
				}
			}
			if len(names) != len(tt.want) {
				t.Fatalf("got %d packages, want %d", len(names), len(tt.want))
			}
			for i, n := range names {
				if n != tt.want[i] {
					t.Errorf("package[%d] = %q, want %q", i, n, tt.want[i])
				}
			}
		})
	}
}

func TestParseFileList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"whitespace_only", "  \n\n  ", nil},
		{
			name:  "single_file",
			input: "/usr/bin/bash\n",
			want:  []string{"/usr/bin/bash"},
		},
		{
			name:  "multiple_files",
			input: "/.\n/usr\n/usr/bin\n/usr/bin/bash\n/usr/share/man/man1/bash.1.gz\n",
			want:  []string{"/.", "/usr", "/usr/bin", "/usr/bin/bash", "/usr/share/man/man1/bash.1.gz"},
		},
		{
			name:  "blank_lines_mixed",
			input: "\n/usr/bin/curl\n\n/usr/share/doc/curl\n\n",
			want:  []string{"/usr/bin/curl", "/usr/share/doc/curl"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFileList(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseFileList() returned %d files, want %d", len(got), len(tt.want))
			}
			for i, f := range got {
				if f != tt.want[i] {
					t.Errorf("file[%d] = %q, want %q", i, f, tt.want[i])
				}
			}
		})
	}
}

func TestParseOwner(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single_package",
			input: "bash: /usr/bin/bash\n",
			want:  "bash",
		},
		{
			name:  "multiple_packages",
			input: "bash, dash: /usr/bin/sh\n",
			want:  "bash",
		},
		{
			name:  "package_with_arch",
			input: "libc6:amd64: /lib/x86_64-linux-gnu/libc.so.6\n",
			want:  "libc6", // parseOwner splits on first colon, arch suffix is stripped
		},
		{
			name:  "multiple_lines",
			input: "coreutils: /usr/bin/ls\ncoreutils: /usr/bin/cat\n",
			want:  "coreutils",
		},
		{
			name:  "no_colon",
			input: "unexpected output without colon",
			want:  "",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace_around_package",
			input: "  nginx  : /usr/sbin/nginx\n",
			want:  "nginx",
		},
		{
			name:  "three_packages_comma_separated",
			input: "pkg1, pkg2, pkg3: /some/path\n",
			want:  "pkg1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOwner(tt.input)
			if got != tt.want {
				t.Errorf("parseOwner() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSourcesLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantURL string
		wantTyp string
	}{
		{
			name:    "empty",
			input:   "",
			wantNil: true,
		},
		{
			name:    "comment",
			input:   "# This is a comment",
			wantNil: true,
		},
		{
			name:    "non_deb_line",
			input:   "some random text",
			wantNil: true,
		},
		{
			name:    "basic_deb",
			input:   "deb http://archive.ubuntu.com/ubuntu/ jammy main",
			wantURL: "http://archive.ubuntu.com/ubuntu/",
			wantTyp: "deb",
		},
		{
			name:    "deb_src",
			input:   "deb-src http://archive.ubuntu.com/ubuntu/ jammy main",
			wantURL: "http://archive.ubuntu.com/ubuntu/",
			wantTyp: "deb-src",
		},
		{
			name:    "with_options",
			input:   "deb [arch=amd64] https://repo.example.com/deb stable main",
			wantURL: "https://repo.example.com/deb",
			wantTyp: "deb",
		},
		{
			name:    "with_signed_by",
			input:   "deb [arch=amd64 signed-by=/etc/apt/keyrings/key.gpg] https://repo.example.com stable main",
			wantURL: "https://repo.example.com",
			wantTyp: "deb",
		},
		{
			name:    "leading_whitespace",
			input:   "   deb http://example.com/repo stable main",
			wantURL: "http://example.com/repo",
			wantTyp: "deb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := parseSourcesLine(tt.input)
			if tt.wantNil {
				if r != nil {
					t.Errorf("parseSourcesLine() = %+v, want nil", r)
				}
				return
			}
			if r == nil {
				t.Fatal("parseSourcesLine() = nil, want non-nil")
			}
			if r.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", r.URL, tt.wantURL)
			}
			if r.Type != tt.wantTyp {
				t.Errorf("Type = %q, want %q", r.Type, tt.wantTyp)
			}
			if !r.Enabled {
				t.Error("expected Enabled=true")
			}
		})
	}
}

func TestExtractURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic_deb",
			input: "deb http://archive.ubuntu.com/ubuntu/ jammy main",
			want:  "http://archive.ubuntu.com/ubuntu/",
		},
		{
			name:  "deb_src",
			input: "deb-src http://archive.ubuntu.com/ubuntu/ jammy main",
			want:  "http://archive.ubuntu.com/ubuntu/",
		},
		{
			name:  "with_options",
			input: "deb [arch=amd64] https://apt.example.com/repo stable main",
			want:  "https://apt.example.com/repo",
		},
		{
			name:  "with_signed_by",
			input: "deb [arch=amd64 signed-by=/etc/apt/keyrings/key.gpg] https://repo.example.com/deb stable main",
			want:  "https://repo.example.com/deb",
		},
		{
			name:  "just_type",
			input: "deb",
			want:  "",
		},
		{
			name:  "empty_options_bracket",
			input: "deb [] http://example.com/repo stable",
			want:  "http://example.com/repo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractURL(tt.input)
			if got != tt.want {
				t.Errorf("extractURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
