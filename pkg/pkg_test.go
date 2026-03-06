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

func TestParseQueryEdgeCases(t *testing.T) {
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
			input:   "vim\t9.0\tVi IMproved\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "vim", Version: "9.0", Description: "Vi IMproved", Installed: true},
			},
		},
		{
			name:    "missing description (two fields only)",
			input:   "bash\t5.2\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "bash", Version: "5.2", Description: "", Installed: true},
			},
		},
		{
			name:    "single field only (no tabs, skipped)",
			input:   "justname\n",
			wantLen: 0,
		},
		{
			name:    "description with tabs",
			input:   "pkg\t1.0\tA\ttabbed\tdescription\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "pkg", Version: "1.0", Description: "A\ttabbed\tdescription", Installed: true},
			},
		},
		{
			name:    "trailing and leading whitespace on lines",
			input:   "  nginx\t1.24.0\tWeb server  \n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "nginx", Version: "1.24.0", Description: "Web server", Installed: true},
			},
		},
		{
			name:    "multiple entries with blank lines between",
			input:   "a\t1.0\tAlpha\n\nb\t2.0\tBeta\n",
			wantLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseQuery(tt.input)
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

func TestParseSearchEdgeCases(t *testing.T) {
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
			name:    "name with many hyphens",
			input:   "py39-django-rest-framework-3.14.0 RESTful Web APIs for Django\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "py39-django-rest-framework", Version: "3.14.0", Description: "RESTful Web APIs for Django"},
			},
		},
		{
			name:    "no comment (name-version only)",
			input:   "zsh-5.9\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "zsh", Version: "5.9", Description: ""},
			},
		},
		{
			name:    "very long output many packages",
			input:   "a-1.0 desc1\nb-2.0 desc2\nc-3.0 desc3\nd-4.0 desc4\ne-5.0 desc5\n",
			wantLen: 5,
		},
		{
			name:    "single character name",
			input:   "R-4.3.2                        Statistical Computing\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "R", Version: "4.3.2", Description: "Statistical Computing"},
			},
		},
		{
			name:    "version with complex suffix",
			input:   "libressl-3.8.2_1               TLS library\n",
			wantLen: 1,
			wantPkgs: []snack.Package{
				{Name: "libressl", Version: "3.8.2_1", Description: "TLS library"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseSearch(tt.input)
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
			}
		})
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

func TestParseInfoEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		want    *snack.Package
	}{
		{
			name:    "empty input",
			input:   "",
			wantNil: true,
		},
		{
			name:    "no name field returns nil",
			input:   "Version : 1.0\nComment : test\n",
			wantNil: true,
		},
		{
			name:  "name only (missing other fields)",
			input: "Name : bash\n",
			want:  &snack.Package{Name: "bash", Installed: true},
		},
		{
			name:  "extra unknown fields are ignored",
			input: "Name : vim\nVersion : 9.0\nMaintainer : someone@example.com\nWWW : https://vim.org\nComment : Vi IMproved\n",
			want:  &snack.Package{Name: "vim", Version: "9.0", Description: "Vi IMproved", Installed: true},
		},
		{
			name:  "colon in value",
			input: "Name : nginx\nComment : HTTP server: fast and reliable\n",
			want:  &snack.Package{Name: "nginx", Description: "HTTP server: fast and reliable", Installed: true},
		},
		{
			name:    "lines without colons are skipped",
			input:   "This is random text\nNo colons here\n",
			wantNil: true,
		},
		{
			name:  "whitespace around values",
			input: "Name           :    curl    \nVersion        :    8.5.0    \n",
			want:  &snack.Package{Name: "curl", Version: "8.5.0", Installed: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInfo(tt.input)
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
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.Installed != tt.want.Installed {
				t.Errorf("Installed = %v, want %v", got.Installed, tt.want.Installed)
			}
		})
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

func TestParseUpgradesEdgeCases(t *testing.T) {
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
			name:    "no upgrade lines",
			input:   "Updating FreeBSD repository catalogue...\nAll packages are up to date.\n",
			wantLen: 0,
		},
		{
			name: "mix of Upgrading Installing Reinstalling",
			input: `Upgrading nginx: 1.24.0 -> 1.26.0
Installing newpkg: 0 -> 1.0.0
Reinstalling bash: 5.2 -> 5.2
`,
			wantLen: 3,
			wantPkgs: []snack.Package{
				{Name: "nginx", Version: "1.26.0", Installed: true},
				{Name: "newpkg", Version: "1.0.0", Installed: true},
				{Name: "bash", Version: "5.2", Installed: true},
			},
		},
		{
			name:    "line with -> but no recognized prefix is skipped",
			input:   "Something: 1.0 -> 2.0\n",
			wantLen: 0,
		},
		{
			name:    "upgrading line without colon is skipped",
			input:   "Upgrading nginx 1.24.0 -> 1.26.0\n",
			wantLen: 0,
		},
		{
			name:    "upgrading line with -> but not enough parts after colon",
			input:   "Upgrading nginx: -> \n",
			wantLen: 0,
		},
		{
			name:    "upgrading line with wrong arrow",
			input:   "Upgrading nginx: 1.24.0 => 1.26.0\n",
			wantLen: 0,
		},
		{
			name:  "single upgrade line",
			input: "Upgrading zsh: 5.8 -> 5.9\n",
			wantPkgs: []snack.Package{
				{Name: "zsh", Version: "5.9", Installed: true},
			},
			wantLen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := parseUpgrades(tt.input)
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

func TestParseFileListEdgeCases(t *testing.T) {
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
			input:   "nginx-1.24.0:\n",
			wantLen: 0,
		},
		{
			name:    "paths with spaces",
			input:   "pkg-1.0:\n\t/usr/local/share/my package/file name.txt\n\t/usr/local/share/another dir/test\n",
			wantLen: 2,
			want: []string{
				"/usr/local/share/my package/file name.txt",
				"/usr/local/share/another dir/test",
			},
		},
		{
			name:    "single file",
			input:   "bash-5.2:\n\t/usr/local/bin/bash\n",
			wantLen: 1,
			want:    []string{"/usr/local/bin/bash"},
		},
		{
			name:    "no header just file paths",
			input:   "/usr/local/bin/curl\n/usr/local/lib/libcurl.so\n",
			wantLen: 2,
		},
		{
			name:    "blank lines between files",
			input:   "pkg-1.0:\n\t/usr/local/bin/a\n\n\t/usr/local/bin/b\n",
			wantLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := parseFileList(tt.input)
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

func TestParseOwner(t *testing.T) {
	input := "/usr/local/sbin/nginx was installed by package nginx-1.24.0\n"
	name := parseOwner(input)
	if name != "nginx" {
		t.Errorf("expected 'nginx', got %q", name)
	}
}

func TestParseOwnerEdgeCases(t *testing.T) {
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
			name:  "standard format",
			input: "/usr/local/bin/curl was installed by package curl-8.5.0\n",
			want:  "curl",
		},
		{
			name:  "package name with hyphens",
			input: "/usr/local/lib/libpython3.so was installed by package py39-python-3.9.18\n",
			want:  "py39-python",
		},
		{
			name:  "no match returns trimmed input",
			input: "some random output\n",
			want:  "some random output",
		},
		{
			name:  "whitespace around",
			input: "  /usr/local/bin/bash was installed by package bash-5.2.21  \n",
			want:  "bash",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOwner(tt.input)
			if got != tt.want {
				t.Errorf("parseOwner(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
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
			input:       "py39-django-rest-3.14.0",
			wantName:    "py39-django-rest",
			wantVersion: "3.14.0",
		},
		{
			name:        "leading hyphen",
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
			name:        "version with underscore suffix",
			input:       "libressl-3.8.2_1",
			wantName:    "libressl",
			wantVersion: "3.8.2_1",
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

func TestInterfaceCompliance(t *testing.T) {
	var _ snack.Manager = (*Pkg)(nil)
	var _ snack.VersionQuerier = (*Pkg)(nil)
	var _ snack.Cleaner = (*Pkg)(nil)
	var _ snack.FileOwner = (*Pkg)(nil)
}

func TestPackageUpgraderInterface(t *testing.T) {
	var _ snack.PackageUpgrader = (*Pkg)(nil)
}

func TestName(t *testing.T) {
	p := New()
	if p.Name() != "pkg" {
		t.Errorf("Name() = %q, want %q", p.Name(), "pkg")
	}
}

func TestCapabilities(t *testing.T) {
	caps := snack.GetCapabilities(New())

	// Should be true
	if !caps.VersionQuery {
		t.Error("expected VersionQuery=true")
	}
	if !caps.Clean {
		t.Error("expected Clean=true")
	}
	if !caps.FileOwnership {
		t.Error("expected FileOwnership=true")
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
	if caps.DryRun {
		t.Error("expected DryRun=false")
	}
	if !caps.PackageUpgrade {
		t.Error("expected PackageUpgrade=true")
	}
}
