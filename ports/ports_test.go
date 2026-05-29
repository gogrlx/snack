package ports

import (
	"testing"

	"github.com/gogrlx/snack"
)

// Compile-time interface assertions.
var (
	_ snack.Manager         = (*Ports)(nil)
	_ snack.VersionQuerier  = (*Ports)(nil)
	_ snack.Cleaner         = (*Ports)(nil)
	_ snack.FileOwner       = (*Ports)(nil)
	_ snack.NameNormalizer  = (*Ports)(nil)
	_ snack.PackageUpgrader = (*Ports)(nil)
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.Name() != "ports" {
		t.Errorf("Name() = %q, want %q", p.Name(), "ports")
	}
}

func TestNormalizeNameMethod(t *testing.T) {
	p := New()
	tests := []struct {
		input string
		want  string
	}{
		{"curl-8.5.0", "curl"},
		{"python-3.11.7p0", "python"},
		{"quirks", "quirks"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := p.NormalizeName(tt.input); got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseArchMethod(t *testing.T) {
	p := New()
	tests := []struct {
		input    string
		wantName string
		wantArch string
	}{
		{"curl-8.5.0", "curl-8.5.0", ""},
		{"python-3.11.7p0", "python-3.11.7p0", ""},
		{"quirks", "quirks", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotName, gotArch := p.ParseArch(tt.input)
			if gotName != tt.wantName || gotArch != tt.wantArch {
				t.Errorf("ParseArch(%q) = (%q, %q), want (%q, %q)", tt.input, gotName, gotArch, tt.wantName, tt.wantArch)
			}
		})
	}
}

func TestParseList(t *testing.T) {
	input := `bash-5.2.21        GNU Bourne Again Shell
curl-8.5.0         command line tool for transferring data
python-3.11.7p0    interpreted object-oriented programming language
`
	pkgs := parseList(input)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || pkgs[0].Version != "5.2.21" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[0].Description != "GNU Bourne Again Shell" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
	if pkgs[2].Name != "python" || pkgs[2].Version != "3.11.7p0" {
		t.Errorf("unexpected third package: %+v", pkgs[2])
	}
}

func TestParseListEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantPkgs []snack.Package
	}{
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "whitespace only",
			input:   "   \n  \n\n",
			wantLen: 0,
		},
		{
			name:    "single entry",
			input:   "vim-9.0.2100 Vi IMproved\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "vim", Version: "9.0.2100", Description: "Vi IMproved", Installed: true},
			},
		},
		{
			name:    "no description",
			input:   "vim-9.0.2100\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "vim", Version: "9.0.2100", Description: "", Installed: true},
			},
		},
		{
			name:    "empty lines between entries",
			input:   "bash-5.2 Shell\n\n\ncurl-8.5.0 Transfer tool\n",
			wantLen: 2,
			wantPkgs: []snack.Package{
				{Name: "bash", Version: "5.2", Description: "Shell", Installed: true},
				{Name: "curl", Version: "8.5.0", Description: "Transfer tool", Installed: true},
			},
		},
		{
			name:    "package with p-suffix version (OpenBSD style)",
			input:   "python-3.11.7p0    interpreted language\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "python", Version: "3.11.7p0", Description: "interpreted language", Installed: true},
			},
		},
		{
			name:    "package name with multiple hyphens",
			input:   "py3-django-rest-3.14.0 REST framework\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "py3-django-rest", Version: "3.14.0", Description: "REST framework", Installed: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseList(tt.input)
			if len(pkgs) != tt.wantLen {
				t.Fatalf("expected %d packages, got %d", tt.wantLen, len(pkgs))
			}
			for i, want := range tt.wantPkgs {
				if i >= len(pkgs) {
					break
				}
				got := pkgs[i]
				if got.Name != want.Name {
					t.Errorf("[%d] Name = %q, want %q", i, got.Name, want.Name)
				}
				if got.Version != want.Version {
					t.Errorf("[%d] Version = %q, want %q", i, got.Version, want.Version)
				}
				if got.Description != want.Description {
					t.Errorf("[%d] Description = %q, want %q", i, got.Description, want.Description)
				}
				if got.Installed != want.Installed {
					t.Errorf("[%d] Installed = %v, want %v", i, got.Installed, want.Installed)
				}
			}
		})
	}
}

func TestParseSearchResults(t *testing.T) {
	input := `nginx-1.24.0
nginx-1.25.3
`
	pkgs := parseSearchResults(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "nginx" || pkgs[0].Version != "1.24.0" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
}

func TestParseSearchResultsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantPkgs []snack.Package
	}{
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "whitespace only",
			input:   "  \n\n  \n",
			wantLen: 0,
		},
		{
			name:    "single result",
			input:   "curl-8.5.0\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "curl", Version: "8.5.0"},
			},
		},
		{
			name:    "name with many hyphens",
			input:   "py3-django-rest-framework-3.14.0\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "py3-django-rest-framework", Version: "3.14.0"},
			},
		},
		{
			name:    "no version (no hyphen)",
			input:   "quirks\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "quirks", Version: ""},
			},
		},
		{
			name:    "p-suffix version",
			input:   "python-3.11.7p0\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "python", Version: "3.11.7p0"},
			},
		},
		{
			name:    "multiple results with blank lines",
			input:   "a-1.0\n\nb-2.0\n\nc-3.0\n",
			wantLen: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseSearchResults(tt.input)
			if len(pkgs) != tt.wantLen {
				t.Fatalf("expected %d packages, got %d", tt.wantLen, len(pkgs))
			}
			for i, want := range tt.wantPkgs {
				if i >= len(pkgs) {
					break
				}
				got := pkgs[i]
				if got.Name != want.Name {
					t.Errorf("[%d] Name = %q, want %q", i, got.Name, want.Name)
				}
				if got.Version != want.Version {
					t.Errorf("[%d] Version = %q, want %q", i, got.Version, want.Version)
				}
			}
		})
	}
}

func TestParseInfoOutput(t *testing.T) {
	input := `Information for nginx-1.24.0:

Comment:
robust and small WWW server
Description:
nginx is an HTTP and reverse proxy server, a mail proxy server,
and a generic TCP/UDP proxy server.
`
	pkg := parseInfoOutput(input, "nginx-1.24.0")
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "nginx" {
		t.Errorf("expected name 'nginx', got %q", pkg.Name)
	}
	if pkg.Version != "1.24.0" {
		t.Errorf("unexpected version: %q", pkg.Version)
	}
}

func TestParseInfoOutputWithComment(t *testing.T) {
	input := `Information for curl-8.5.0:

Comment: command line tool for transferring data
Description:
curl is a tool to transfer data from or to a server.
`
	pkg := parseInfoOutput(input, "curl-8.5.0")
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "curl" {
		t.Errorf("expected name 'curl', got %q", pkg.Name)
	}
	if pkg.Description != "command line tool for transferring data" {
		t.Errorf("unexpected description: %q", pkg.Description)
	}
}

func TestParseInfoOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pkgArg  string
		wantNil bool
		want    *snack.Package
	}{
		{
			name:    "empty input and empty pkg arg",
			input:   "",
			pkgArg:  "",
			wantNil: true,
		},
		{
			name:   "empty input with pkg arg fallback",
			input:  "",
			pkgArg: "curl-8.5.0",
			want:   &snack.Package{Name: "curl", Version: "8.5.0", Installed: true},
		},
		{
			name:   "no header, falls back to pkg arg",
			input:  "Some random output\nwithout the expected header",
			pkgArg: "vim-9.0",
			want:   &snack.Package{Name: "vim", Version: "9.0", Installed: true},
		},
		{
			name: "comment on same line as Comment: label",
			input: `Information for zsh-5.9:

Comment: Zsh shell
`,
			pkgArg: "zsh-5.9",
			want:   &snack.Package{Name: "zsh", Version: "5.9", Description: "Zsh shell", Installed: true},
		},
		{
			name: "comment on next line after Comment: label (not captured as description)",
			input: `Information for bash-5.2:

Comment:
GNU Bourne Again Shell
`,
			pkgArg: "bash-5.2",
			// Comment: with nothing after the colon sets Description="",
			// and the next line isn't in a Description: block so it's ignored.
			want: &snack.Package{Name: "bash", Version: "5.2", Description: "", Installed: true},
		},
		{
			name: "description spans multiple lines",
			input: `Information for git-2.43.0:

Comment: distributed version control system
Description:
Git is a fast, scalable, distributed revision control system
with an unusually rich command set.
`,
			pkgArg: "git-2.43.0",
			want:   &snack.Package{Name: "git", Version: "2.43.0", Description: "distributed version control system", Installed: true},
		},
		{
			name: "extra fields are ignored",
			input: `Information for tmux-3.3a:

Comment: terminal multiplexer
Maintainer: someone@openbsd.org
WWW: https://tmux.github.io
Description:
tmux is a terminal multiplexer.
`,
			pkgArg: "tmux-3.3a",
			want:   &snack.Package{Name: "tmux", Version: "3.3a", Description: "terminal multiplexer", Installed: true},
		},
		{
			name:   "pkg arg with no version (no hyphen)",
			input:  "",
			pkgArg: "quirks",
			want:   &snack.Package{Name: "quirks", Version: "", Installed: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInfoOutput(tt.input, tt.pkgArg)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil package")
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %q, want %q", got.Version, tt.want.Version)
			}
			if tt.want.Description != "" && got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.Installed != tt.want.Installed {
				t.Errorf("Installed = %v, want %v", got.Installed, tt.want.Installed)
			}
		})
	}
}

func TestParseInfoOutputEmpty(t *testing.T) {
	pkg := parseInfoOutput("", "")
	if pkg != nil {
		t.Error("expected nil for empty input and empty pkg name")
	}
}

func TestParseInfoOutputFallbackToPkgArg(t *testing.T) {
	input := "Some random output\nwithout the expected header"
	pkg := parseInfoOutput(input, "curl-8.5.0")
	if pkg == nil {
		t.Fatal("expected non-nil package from fallback")
	}
	if pkg.Name != "curl" || pkg.Version != "8.5.0" {
		t.Errorf("unexpected fallback parse: %+v", pkg)
	}
}

func TestSplitNameVersion(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion string
	}{
		{"nginx-1.24.0", "nginx", "1.24.0"},
		{"py3-pip-23.1", "py3-pip", "23.1"},
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

func TestSplitNameVersionEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantVersion string
	}{
		{
			name:        "empty string",
			input:       "",
			wantName:    "",
			wantVersion: "",
		},
		{
			name:        "no hyphen",
			input:       "singleword",
			wantName:    "singleword",
			wantVersion: "",
		},
		{
			name:        "multiple hyphens",
			input:       "py3-django-rest-3.14.0",
			wantName:    "py3-django-rest",
			wantVersion: "3.14.0",
		},
		{
			name:        "leading hyphen (idx=0, returns whole string)",
			input:       "-1.0",
			wantName:    "-1.0",
			wantVersion: "",
		},
		{
			name:        "trailing hyphen",
			input:       "nginx-",
			wantName:    "nginx",
			wantVersion: "",
		},
		{
			name:        "only hyphen",
			input:       "-",
			wantName:    "-",
			wantVersion: "",
		},
		{
			name:        "hyphen at index 1",
			input:       "a-1.0",
			wantName:    "a",
			wantVersion: "1.0",
		},
		{
			name:        "p-suffix version",
			input:       "python-3.11.7p0",
			wantName:    "python",
			wantVersion: "3.11.7p0",
		},
		{
			name:        "version with v prefix",
			input:       "go-v1.21.5",
			wantName:    "go",
			wantVersion: "v1.21.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ver := splitNameVersion(tt.input)
			if name != tt.wantName || ver != tt.wantVersion {
				t.Errorf("splitNameVersion(%q) = (%q, %q), want (%q, %q)",
					tt.input, name, ver, tt.wantName, tt.wantVersion)
			}
		})
	}
}

func TestSplitNameVersionNoHyphen(t *testing.T) {
	name, ver := splitNameVersion("singleword")
	if name != "singleword" || ver != "" {
		t.Errorf("splitNameVersion(\"singleword\") = (%q, %q), want (\"singleword\", \"\")", name, ver)
	}
}

func TestSplitNameVersionLeadingHyphen(t *testing.T) {
	name, ver := splitNameVersion("-1.0")
	if name != "-1.0" || ver != "" {
		t.Errorf("splitNameVersion(\"-1.0\") = (%q, %q), want (\"-1.0\", \"\")", name, ver)
	}
}

func TestParseUpgradeOutput(t *testing.T) {
	input := `quirks-7.14 -> quirks-7.18
curl-8.5.0 -> curl-8.6.0
python-3.11.7p0 -> python-3.11.8p0
`
	pkgs := parseUpgradeOutput(input)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "quirks" || pkgs[0].Version != "7.18" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[1].Name != "curl" || pkgs[1].Version != "8.6.0" {
		t.Errorf("unexpected second package: %+v", pkgs[1])
	}
	if pkgs[2].Name != "python" || pkgs[2].Version != "3.11.8p0" {
		t.Errorf("unexpected third package: %+v", pkgs[2])
	}
}

func TestParseUpgradeOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantPkgs []snack.Package
	}{
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "whitespace only",
			input:   "  \n\n  \n",
			wantLen: 0,
		},
		{
			name:    "single upgrade",
			input:   "bash-5.2 -> bash-5.3\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "bash", Version: "5.3", Installed: true},
			},
		},
		{
			name:    "line without -> is skipped",
			input:   "Some info line\nbash-5.2 -> bash-5.3\nAnother line\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "bash", Version: "5.3", Installed: true},
			},
		},
		{
			name:    "malformed line (-> but not enough fields)",
			input:   "-> bash-5.3\n",
			wantLen: 0,
		},
		{
			name:    "wrong arrow (=> instead of ->)",
			input:   "bash-5.2 => bash-5.3\n",
			wantLen: 0,
		},
		{
			name:    "package name with multiple hyphens",
			input:   "py3-django-rest-3.14.0 -> py3-django-rest-3.15.0\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "py3-django-rest", Version: "3.15.0", Installed: true},
			},
		},
		{
			name:    "p-suffix versions",
			input:   "python-3.11.7p0 -> python-3.11.8p1\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "python", Version: "3.11.8p1", Installed: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseUpgradeOutput(tt.input)
			if len(pkgs) != tt.wantLen {
				t.Fatalf("expected %d packages, got %d", tt.wantLen, len(pkgs))
			}
			for i, want := range tt.wantPkgs {
				if i >= len(pkgs) {
					break
				}
				got := pkgs[i]
				if got.Name != want.Name {
					t.Errorf("[%d] Name = %q, want %q", i, got.Name, want.Name)
				}
				if got.Version != want.Version {
					t.Errorf("[%d] Version = %q, want %q", i, got.Version, want.Version)
				}
				if got.Installed != want.Installed {
					t.Errorf("[%d] Installed = %v, want %v", i, got.Installed, want.Installed)
				}
			}
		})
	}
}

func TestParseUpgradeOutputEmpty(t *testing.T) {
	pkgs := parseUpgradeOutput("")
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseFileListOutput(t *testing.T) {
	input := `Information for curl-8.5.0:

Files:
/usr/local/bin/curl
/usr/local/include/curl/curl.h
/usr/local/lib/libcurl.so.26.0
/usr/local/man/man1/curl.1
`
	files := parseFileListOutput(input)
	if len(files) != 4 {
		t.Fatalf("expected 4 files, got %d", len(files))
	}
	if files[0] != "/usr/local/bin/curl" {
		t.Errorf("unexpected first file: %q", files[0])
	}
}

func TestParseFileListOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		want    []string
	}{
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "header only no files",
			input:   "Information for pkg-1.0:\n\nFiles:\n",
			wantLen: 0,
		},
		{
			name:    "paths with spaces",
			input:   "Information for pkg-1.0:\n\nFiles:\n/usr/local/share/my dir/file name.txt\n/usr/local/share/another path/test\n",
			wantLen: 2,
			want: []string{
				"/usr/local/share/my dir/file name.txt",
				"/usr/local/share/another path/test",
			},
		},
		{
			name:    "single file",
			input:   "Files:\n/usr/local/bin/bash\n",
			wantLen: 1,
			want:    []string{"/usr/local/bin/bash"},
		},
		{
			name:    "no header just paths",
			input:   "/usr/local/bin/a\n/usr/local/bin/b\n",
			wantLen: 2,
		},
		{
			name:    "blank lines between files",
			input:   "Files:\n/usr/local/bin/a\n\n/usr/local/bin/b\n",
			wantLen: 2,
		},
		{
			name:    "non-path lines are skipped",
			input:   "Information for pkg-1.0:\n\nFiles:\nNot a path\n/usr/local/bin/real\n",
			wantLen: 1,
			want:    []string{"/usr/local/bin/real"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := parseFileListOutput(tt.input)
			if len(files) != tt.wantLen {
				t.Fatalf("expected %d files, got %d", tt.wantLen, len(files))
			}
			for i, w := range tt.want {
				if i >= len(files) {
					break
				}
				if files[i] != w {
					t.Errorf("[%d] got %q, want %q", i, files[i], w)
				}
			}
		})
	}
}

func TestParseFileListOutputEmpty(t *testing.T) {
	files := parseFileListOutput("")
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestParseOwnerOutput(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"curl-8.5.0", "curl"},
		{"python-3.11.7p0", "python"},
		{"", ""},
	}
	for _, tt := range tests {
		got := parseOwnerOutput(tt.input)
		if got != tt.want {
			t.Errorf("parseOwnerOutput(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseOwnerOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace only",
			input: "   \n  ",
			want:  "",
		},
		{
			name:  "name with many hyphens",
			input: "py3-django-rest-framework-3.14.0",
			want:  "py3-django-rest-framework",
		},
		{
			name:  "no version (no hyphen)",
			input: "quirks",
			want:  "quirks",
		},
		{
			name:  "leading/trailing whitespace",
			input: "  curl-8.5.0  ",
			want:  "curl",
		},
		{
			name:  "p-suffix version",
			input: "python-3.11.7p0",
			want:  "python",
		},
		{
			name:  "trailing newline",
			input: "bash-5.2\n",
			want:  "bash",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOwnerOutput(tt.input)
			if got != tt.want {
				t.Errorf("parseOwnerOutput(%q) = %q, want %q", tt.input, got, tt.want)
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
		{"PackageUpgrade", caps.PackageUpgrade, true},
		{"Hold", caps.Hold, false},
		{"RepoManagement", caps.RepoManagement, false},
		{"KeyManagement", caps.KeyManagement, false},
		{"Groups", caps.Groups, false},
		{"DryRun", caps.DryRun, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}
