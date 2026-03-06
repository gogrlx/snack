package winget

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseTableList(t *testing.T) {
	input := `Name                              Id                              Version        Available  Source
---------------------------------------------------------------------------------------------------------
Visual Studio Code                Microsoft.VisualStudioCode       1.85.0         1.86.0     winget
Git                               Git.Git                          2.43.0                    winget
Google Chrome                     Google.Chrome                    120.0.6099.130            winget
`
	pkgs := parseTable(input, true)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "Microsoft.VisualStudioCode" {
		t.Errorf("expected ID as name, got %q", pkgs[0].Name)
	}
	if pkgs[0].Description != "Visual Studio Code" {
		t.Errorf("expected display name as description, got %q", pkgs[0].Description)
	}
	if pkgs[0].Version != "1.85.0" {
		t.Errorf("expected version '1.85.0', got %q", pkgs[0].Version)
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
	if pkgs[0].Repository != "winget" {
		t.Errorf("expected repository 'winget', got %q", pkgs[0].Repository)
	}
	if pkgs[1].Name != "Git.Git" {
		t.Errorf("expected 'Git.Git', got %q", pkgs[1].Name)
	}
}

func TestParseTableSearch(t *testing.T) {
	input := `Name                Id                              Version  Match            Source
-------------------------------------------------------------------------------------------------
Visual Studio Code  Microsoft.VisualStudioCode       1.86.0   Moniker: vscode  winget
VSCodium            VSCodium.VSCodium                1.85.2                    winget
`
	pkgs := parseTable(input, false)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "Microsoft.VisualStudioCode" {
		t.Errorf("expected 'Microsoft.VisualStudioCode', got %q", pkgs[0].Name)
	}
	if pkgs[0].Installed {
		t.Error("expected Installed=false for search")
	}
}

func TestParseTableEmpty(t *testing.T) {
	input := `Name  Id  Version  Source
------------------------------
`
	pkgs := parseTable(input, true)
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseTableNoSeparator(t *testing.T) {
	pkgs := parseTable("No installed package found matching input criteria.", true)
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseTableWithFooter(t *testing.T) {
	input := `Name            Id              Version   Available  Source
--------------------------------------------------------------
Git             Git.Git         2.43.0    2.44.0     winget
3 upgrades available.
`
	pkgs := parseTable(input, false)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "Git.Git" {
		t.Errorf("expected 'Git.Git', got %q", pkgs[0].Name)
	}
}

func TestParseTableUpgrade(t *testing.T) {
	input := `Name                              Id                              Version        Available  Source
---------------------------------------------------------------------------------------------------------
Visual Studio Code                Microsoft.VisualStudioCode       1.85.0         1.86.0     winget
Node.js                           OpenJS.NodeJS                    20.10.0        21.5.0     winget
2 upgrades available.
`
	pkgs := parseTable(input, false)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "Microsoft.VisualStudioCode" {
		t.Errorf("expected 'Microsoft.VisualStudioCode', got %q", pkgs[0].Name)
	}
	if pkgs[1].Name != "OpenJS.NodeJS" {
		t.Errorf("expected 'OpenJS.NodeJS', got %q", pkgs[1].Name)
	}
}

func TestParseShow(t *testing.T) {
	input := `Found Visual Studio Code [Microsoft.VisualStudioCode]
Version: 1.86.0
Publisher: Microsoft Corporation
Publisher URL: https://code.visualstudio.com
Author: Microsoft Corporation
Moniker: vscode
Description: Code editing. Redefined.
License: MIT
Installer Type: inno
`
	pkg := parseShow(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "Microsoft.VisualStudioCode" {
		t.Errorf("expected 'Microsoft.VisualStudioCode', got %q", pkg.Name)
	}
	if pkg.Version != "1.86.0" {
		t.Errorf("expected version '1.86.0', got %q", pkg.Version)
	}
	if pkg.Description != "Visual Studio Code" {
		t.Errorf("expected 'Visual Studio Code', got %q", pkg.Description)
	}
}

func TestParseShowNotFound(t *testing.T) {
	pkg := parseShow("No package found matching input criteria.")
	if pkg != nil {
		t.Error("expected nil for not-found output")
	}
}

func TestParseShowEmpty(t *testing.T) {
	pkg := parseShow("")
	if pkg != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseShowNoID(t *testing.T) {
	input := `Version: 1.0.0
Publisher: Someone
`
	pkg := parseShow(input)
	if pkg != nil {
		t.Error("expected nil when no Found header with ID")
	}
}

func TestParseSourceList(t *testing.T) {
	input := `Name    Argument
--------------------------------------------------------------
winget  https://cdn.winget.microsoft.com/cache
msstore https://storeedgefd.dsx.mp.microsoft.com/v9.0
`
	repos := parseSourceList(input)
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].Name != "winget" {
		t.Errorf("expected 'winget', got %q", repos[0].Name)
	}
	if repos[0].URL != "https://cdn.winget.microsoft.com/cache" {
		t.Errorf("unexpected URL: %q", repos[0].URL)
	}
	if !repos[0].Enabled {
		t.Error("expected Enabled=true")
	}
	if repos[1].Name != "msstore" {
		t.Errorf("expected 'msstore', got %q", repos[1].Name)
	}
}

func TestParseSourceListEmpty(t *testing.T) {
	repos := parseSourceList("")
	if len(repos) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(repos))
	}
}

func TestIsSeparatorLine(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"---", true},
		{"--------------------------------------------------------------", true},
		{"--- --- ---", true},
		{"Name    Id    Version", false},
		{"", false},
		{"   ", false},
		{"--a--", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isSeparatorLine(tt.input)
			if got != tt.want {
				t.Errorf("isSeparatorLine(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectColumns(t *testing.T) {
	header := "Name                              Id                              Version        Available  Source"
	cols := detectColumns(header)
	if len(cols) != 5 {
		t.Fatalf("expected 5 columns, got %d: %v", len(cols), cols)
	}
	if cols[0] != 0 {
		t.Errorf("expected first column at 0, got %d", cols[0])
	}
}

func TestExtractFields(t *testing.T) {
	line := "Visual Studio Code                Microsoft.VisualStudioCode       1.85.0         1.86.0     winget"
	cols := []int{0, 34, 67, 82, 93}
	fields := extractFields(line, cols)
	if len(fields) != 5 {
		t.Fatalf("expected 5 fields, got %d", len(fields))
	}
	if fields[0] != "Visual Studio Code" {
		t.Errorf("field[0] = %q, want 'Visual Studio Code'", fields[0])
	}
	if fields[1] != "Microsoft.VisualStudioCode" {
		t.Errorf("field[1] = %q, want 'Microsoft.VisualStudioCode'", fields[1])
	}
	if fields[2] != "1.85.0" {
		t.Errorf("field[2] = %q, want '1.85.0'", fields[2])
	}
	if fields[3] != "1.86.0" {
		t.Errorf("field[3] = %q, want '1.86.0'", fields[3])
	}
	if fields[4] != "winget" {
		t.Errorf("field[4] = %q, want 'winget'", fields[4])
	}
}

func TestExtractFieldsShortLine(t *testing.T) {
	line := "Git"
	cols := []int{0, 34, 67}
	fields := extractFields(line, cols)
	if fields[0] != "Git" {
		t.Errorf("field[0] = %q, want 'Git'", fields[0])
	}
	if fields[1] != "" {
		t.Errorf("field[1] = %q, want ''", fields[1])
	}
}

func TestIsFooterLine(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"3 upgrades available.", true},
		{"1 package(s) found.", true},
		{"No installed package found.", true},
		{"The following packages have upgrades:", true},
		{"Git  Git.Git  2.43.0", false},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isFooterLine(tt.input)
			if got != tt.want {
				t.Errorf("isFooterLine(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
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
		{"real versions", "1.85.0", "1.86.0", -1},
		{"single component", "5", "3", 1},
		{"empty vs empty", "", "", 0},
		{"empty vs version", "", "1.0", -1},
		{"version vs empty", "1.0", "", 1},
		{"four components", "120.0.6099.130", "120.0.6099.131", -1},
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
	var _ snack.Manager = (*Winget)(nil)
	var _ snack.VersionQuerier = (*Winget)(nil)
	var _ snack.RepoManager = (*Winget)(nil)
	var _ snack.NameNormalizer = (*Winget)(nil)
	var _ snack.PackageUpgrader = (*Winget)(nil)
}

func TestCapabilities(t *testing.T) {
	caps := snack.GetCapabilities(New())
	if !caps.VersionQuery {
		t.Error("expected VersionQuery=true")
	}
	if !caps.RepoManagement {
		t.Error("expected RepoManagement=true")
	}
	if !caps.NameNormalize {
		t.Error("expected NameNormalize=true")
	}
	if !caps.PackageUpgrade {
		t.Error("expected PackageUpgrade=true")
	}
	// Should be false
	if caps.Clean {
		t.Error("expected Clean=false")
	}
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
}

func TestName(t *testing.T) {
	w := New()
	if w.Name() != "winget" {
		t.Errorf("Name() = %q, want %q", w.Name(), "winget")
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Microsoft.VisualStudioCode", "Microsoft.VisualStudioCode"},
		{"  Git.Git  ", "Git.Git"},
		{"", ""},
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

func TestParseArch(t *testing.T) {
	name, arch := parseArch("Microsoft.VisualStudioCode")
	if name != "Microsoft.VisualStudioCode" || arch != "" {
		t.Errorf("parseArch returned (%q, %q), want (%q, %q)",
			name, arch, "Microsoft.VisualStudioCode", "")
	}
}

func TestStripVT(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no escapes", "hello", "hello"},
		{"simple CSI", "\x1b[2Khello", "hello"},
		{"color code", "\x1b[32mgreen\x1b[0m", "green"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripVT(tt.input)
			if got != tt.want {
				t.Errorf("stripVT(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsProgressLine(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"████████  50%", true},
		{"100%", true},
		{"Git  Git.Git  2.43.0", false},
		{"", false},
		{"Installing...", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isProgressLine(tt.input)
			if got != tt.want {
				t.Errorf("isProgressLine(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
