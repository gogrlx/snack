package dnf

import (
	"strings"
	"testing"
)

func TestParseList(t *testing.T) {
	input := `Last metadata expiration check: 0:42:03 ago on Wed 26 Feb 2025 10:00:00 AM UTC.
Installed Packages
acl.x86_64                      2.3.1-4.el9                    @anaconda
bash.x86_64                     5.1.8-6.el9                    @anaconda
curl.x86_64                     7.76.1-23.el9                  @baseos
`
	pkgs := parseList(input)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	tests := []struct {
		name, ver, arch, repo string
	}{
		{"acl", "2.3.1-4.el9", "x86_64", "@anaconda"},
		{"bash", "5.1.8-6.el9", "x86_64", "@anaconda"},
		{"curl", "7.76.1-23.el9", "x86_64", "@baseos"},
	}
	for i, tt := range tests {
		if pkgs[i].Name != tt.name {
			t.Errorf("pkg[%d].Name = %q, want %q", i, pkgs[i].Name, tt.name)
		}
		if pkgs[i].Version != tt.ver {
			t.Errorf("pkg[%d].Version = %q, want %q", i, pkgs[i].Version, tt.ver)
		}
		if pkgs[i].Arch != tt.arch {
			t.Errorf("pkg[%d].Arch = %q, want %q", i, pkgs[i].Arch, tt.arch)
		}
		if pkgs[i].Repository != tt.repo {
			t.Errorf("pkg[%d].Repository = %q, want %q", i, pkgs[i].Repository, tt.repo)
		}
	}
}

func TestParseListUpgrades(t *testing.T) {
	input := `Available Upgrades
curl.x86_64                     7.76.1-26.el9                  baseos
vim-minimal.x86_64              2:9.0.1572-1.el9               appstream
`
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" || pkgs[0].Version != "7.76.1-26.el9" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
}

func TestParseSearch(t *testing.T) {
	input := `Last metadata expiration check: 0:10:00 ago.
=== Name Exactly Matched: nginx ===
nginx.x86_64 : A high performance web server and reverse proxy server
=== Name & Summary Matched: nginx ===
nginx-mod-http-perl.x86_64 : Nginx HTTP perl module
`
	pkgs := parseSearch(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "nginx" || pkgs[0].Arch != "x86_64" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[0].Description != "A high performance web server and reverse proxy server" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
}

func TestParseInfo(t *testing.T) {
	input := `Last metadata expiration check: 0:10:00 ago.
Available Packages
Name         : nginx
Version      : 1.20.1
Release      : 14.el9_2.1
Architecture : x86_64
Size         : 45 k
Source        : nginx-1.20.1-14.el9_2.1.src.rpm
Repository   : appstream
Summary      : A high performance web server
License      : BSD
Description  : Nginx is a web server.
`
	p := parseInfo(input)
	if p == nil {
		t.Fatal("expected package, got nil")
	}
	if p.Name != "nginx" {
		t.Errorf("Name = %q, want nginx", p.Name)
	}
	if p.Version != "1.20.1-14.el9_2.1" {
		t.Errorf("Version = %q, want 1.20.1-14.el9_2.1", p.Version)
	}
	if p.Arch != "x86_64" {
		t.Errorf("Arch = %q, want x86_64", p.Arch)
	}
	if p.Repository != "appstream" {
		t.Errorf("Repository = %q, want appstream", p.Repository)
	}
}

func TestParseVersionLock(t *testing.T) {
	input := `Last metadata expiration check: 0:05:00 ago.
nginx-0:1.20.1-14.el9_2.1.*
curl-0:7.76.1-23.el9.*
`
	pkgs := parseVersionLock(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "nginx" {
		t.Errorf("pkg[0].Name = %q, want nginx", pkgs[0].Name)
	}
	if pkgs[1].Name != "curl" {
		t.Errorf("pkg[1].Name = %q, want curl", pkgs[1].Name)
	}
}

func TestParseRepoList(t *testing.T) {
	input := `repo id                              repo name                                        status
appstream                            CentOS Stream 9 - AppStream                      enabled
baseos                               CentOS Stream 9 - BaseOS                         enabled
crb                                  CentOS Stream 9 - CRB                            disabled
`
	repos := parseRepoList(input)
	if len(repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(repos))
	}
	if repos[0].ID != "appstream" || !repos[0].Enabled {
		t.Errorf("unexpected repo[0]: %+v", repos[0])
	}
	if repos[2].ID != "crb" || repos[2].Enabled {
		t.Errorf("unexpected repo[2]: %+v", repos[2])
	}
}

func TestParseGroupList(t *testing.T) {
	input := `Available Groups:
   Container Management
   Development Tools
   Headless Management
Installed Groups:
   Minimal Install
`
	groups := parseGroupList(input)
	if len(groups) != 4 {
		t.Fatalf("expected 4 groups, got %d", len(groups))
	}
	if groups[0] != "Container Management" {
		t.Errorf("groups[0] = %q, want Container Management", groups[0])
	}
}

func TestParseGroupInfo(t *testing.T) {
	input := `Group: Development Tools
 Description: A basic development environment.
 Mandatory Packages:
   autoconf
   automake
   gcc
 Default Packages:
   byacc
   flex
 Optional Packages:
   ElectricFence
`
	pkgs := parseGroupInfo(input)
	if len(pkgs) != 6 {
		t.Fatalf("expected 6 packages, got %d", len(pkgs))
	}
	names := make(map[string]bool)
	for _, p := range pkgs {
		names[p.Name] = true
	}
	for _, want := range []string{"autoconf", "automake", "gcc", "byacc", "flex", "ElectricFence"} {
		if !names[want] {
			t.Errorf("missing package %q", want)
		}
	}
}

func TestParseGroupIsInstalled(t *testing.T) {
	input := `Available Groups:
   Container Management
   Development Tools
   Headless Management
Installed Groups:
   Minimal Install
   Server
`
	tests := []struct {
		group string
		want  bool
	}{
		{"Minimal Install", true},
		{"Server", true},
		{"Development Tools", false},
		{"Container Management", false},
		{"Nonexistent Group", false},
	}
	for _, tt := range tests {
		got := parseGroupIsInstalled(input, tt.group)
		if got != tt.want {
			t.Errorf("parseGroupIsInstalled(%q) = %v, want %v", tt.group, got, tt.want)
		}
	}

	// Verify that empty lines within the Installed section don't stop parsing.
	inputWithBlankLines := `Installed Groups:
   First Group

   Second Group
`
	if !parseGroupIsInstalled(inputWithBlankLines, "Second Group") {
		t.Error("parseGroupIsInstalled: should find group after blank line in installed section")
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"nginx.x86_64", "nginx"},
		{"curl.aarch64", "curl"},
		{"bash.noarch", "bash"},
		{"python3", "python3"},
		{"glibc.i686", "glibc"},
	}
	for _, tt := range tests {
		got := normalizeName(tt.input)
		if got != tt.want {
			t.Errorf("normalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseArch(t *testing.T) {
	tests := []struct {
		input, wantName, wantArch string
	}{
		{"nginx.x86_64", "nginx", "x86_64"},
		{"curl.aarch64", "curl", "aarch64"},
		{"bash", "bash", ""},
		{"python3.11.noarch", "python3.11", "noarch"},
	}
	for _, tt := range tests {
		name, arch := parseArch(tt.input)
		if name != tt.wantName || arch != tt.wantArch {
			t.Errorf("parseArch(%q) = (%q, %q), want (%q, %q)", tt.input, name, arch, tt.wantName, tt.wantArch)
		}
	}
}

// --- dnf5 parser tests ---

func TestStripPreamble(t *testing.T) {
	input := "Updating and loading repositories:\n Fedora 43 - x86_64              100% |  10.2 MiB/s |  20.5 MiB | 00m02s\nRepositories loaded.\nInstalled packages\nbash.x86_64   5.3.0-2.fc43   abc123\n"
	got := stripPreamble(input)
	if strings.Contains(got, "Updating and loading") {
		t.Error("preamble not stripped")
	}
	if strings.Contains(got, "Repositories loaded") {
		t.Error("preamble tail not stripped")
	}
	if !strings.Contains(got, "bash.x86_64") {
		t.Error("content was incorrectly stripped")
	}
}

func TestParseListDNF5(t *testing.T) {
	input := `Installed packages
alternatives.x86_64   1.33-3.fc43   a899a9b296804e8ab27411270a04f5e9
bash.x86_64           5.3.0-2.fc43  3b3d0b7480cd48d19a2c4259e547f2da
`
	pkgs := parseListDNF5(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "alternatives" || pkgs[0].Version != "1.33-3.fc43" || pkgs[0].Arch != "x86_64" {
		t.Errorf("unexpected pkg[0]: %+v", pkgs[0])
	}
	if pkgs[1].Name != "bash" || pkgs[1].Version != "5.3.0-2.fc43" {
		t.Errorf("unexpected pkg[1]: %+v", pkgs[1])
	}
}

func TestParseListDNF5WithPreamble(t *testing.T) {
	input := "Updating and loading repositories:\nRepositories loaded.\nInstalled packages\nbash.x86_64   5.3.0-2.fc43  abc\n"
	pkgs := parseListDNF5(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" {
		t.Errorf("expected bash, got %q", pkgs[0].Name)
	}
}

func TestParseSearchDNF5(t *testing.T) {
	input := `Matched fields: name
 tree.x86_64  File system tree viewer
 treescan.noarch  Scan directory trees, list directories/files, stat, sync, grep
Matched fields: summary
 baobab.x86_64  A graphical directory tree analyzer
`
	pkgs := parseSearchDNF5(input)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "tree" || pkgs[0].Arch != "x86_64" {
		t.Errorf("unexpected pkg[0]: %+v", pkgs[0])
	}
	if pkgs[0].Description != "File system tree viewer" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
	if pkgs[2].Name != "baobab" {
		t.Errorf("unexpected pkg[2]: %+v", pkgs[2])
	}
}

func TestParseInfoDNF5(t *testing.T) {
	input := `Available packages
Name           : tree
Epoch          : 0
Version        : 2.2.1
Release        : 2.fc43
Architecture   : x86_64
Download size  : 61.3 KiB
Installed size : 112.2 KiB
Source         : tree-pkg-2.2.1-2.fc43.src.rpm
Repository     : fedora
Summary        : File system tree viewer
`
	p := parseInfoDNF5(input)
	if p == nil {
		t.Fatal("expected package, got nil")
	}
	if p.Name != "tree" {
		t.Errorf("Name = %q, want tree", p.Name)
	}
	if p.Version != "2.2.1-2.fc43" {
		t.Errorf("Version = %q, want 2.2.1-2.fc43", p.Version)
	}
	if p.Arch != "x86_64" {
		t.Errorf("Arch = %q, want x86_64", p.Arch)
	}
	if p.Repository != "fedora" {
		t.Errorf("Repository = %q, want fedora", p.Repository)
	}
	if p.Description != "File system tree viewer" {
		t.Errorf("Description = %q", p.Description)
	}
}

func TestParseGroupListDNF5(t *testing.T) {
	input := `ID                          Name                                        Installed
neuron-modelling-simulators Neuron Modelling Simulators                        no
kde-desktop                 KDE                                                no
`
	groups := parseGroupListDNF5(input)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0] != "Neuron Modelling Simulators" {
		t.Errorf("groups[0] = %q", groups[0])
	}
	if groups[1] != "KDE" {
		t.Errorf("groups[1] = %q", groups[1])
	}
}

func TestParseGroupIsInstalledDNF5(t *testing.T) {
	input := `ID                          Name                                        Installed
neuron-modelling-simulators Neuron Modelling Simulators                        no
kde-desktop                 KDE                                                yes
`
	tests := []struct {
		group string
		want  bool
	}{
		{"KDE", true},
		{"Neuron Modelling Simulators", false},
		{"Nonexistent Group", false},
	}
	for _, tt := range tests {
		got := parseGroupIsInstalledDNF5(input, tt.group)
		if got != tt.want {
			t.Errorf("parseGroupIsInstalledDNF5(%q) = %v, want %v", tt.group, got, tt.want)
		}
	}
}

func TestParseGroupInfoDNF5(t *testing.T) {
	input := `Id                   : kde-desktop
Name                 : KDE
Description          : The KDE Plasma Workspaces...
Installed            : no
Mandatory packages   : plasma-desktop
                     : plasma-workspace
Default packages     : NetworkManager-config-connectivity-fedora
`
	pkgs := parseGroupInfoDNF5(input)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	names := map[string]bool{}
	for _, p := range pkgs {
		names[p.Name] = true
	}
	for _, want := range []string{"plasma-desktop", "plasma-workspace", "NetworkManager-config-connectivity-fedora"} {
		if !names[want] {
			t.Errorf("missing package %q", want)
		}
	}
}

func TestParseVersionLockDNF5(t *testing.T) {
	input := `# Added by 'versionlock add' command on 2026-02-26 03:14:29
Package name: tree
evr = 2.2.1-2.fc43
# Added by 'versionlock add' command on 2026-02-26 03:14:45
Package name: curl
evr = 8.11.1-3.fc43
`
	pkgs := parseVersionLockDNF5(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "tree" {
		t.Errorf("pkg[0].Name = %q, want tree", pkgs[0].Name)
	}
	if pkgs[1].Name != "curl" {
		t.Errorf("pkg[1].Name = %q, want curl", pkgs[1].Name)
	}
}

func TestParseRepoListDNF5(t *testing.T) {
	input := `repo id                         repo name                                           status
fedora                          Fedora 43 - x86_64                                 enabled
updates                         Fedora 43 - x86_64 - Updates                       enabled
updates-testing                 Fedora 43 - x86_64 - Test Updates                 disabled
`
	repos := parseRepoListDNF5(input)
	if len(repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(repos))
	}
	if repos[0].ID != "fedora" || !repos[0].Enabled {
		t.Errorf("unexpected repo[0]: %+v", repos[0])
	}
	if repos[2].ID != "updates-testing" || repos[2].Enabled {
		t.Errorf("unexpected repo[2]: %+v", repos[2])
	}
}

// --- Edge case tests ---

func TestParseListEmpty(t *testing.T) {
	pkgs := parseList("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty input, got %d", len(pkgs))
	}
}

func TestParseListSinglePackage(t *testing.T) {
	input := `Installed Packages
curl.x86_64                     7.76.1-23.el9                  @baseos
`
	pkgs := parseList(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" {
		t.Errorf("Name = %q, want curl", pkgs[0].Name)
	}
}

func TestParseListMalformedLines(t *testing.T) {
	input := `Installed Packages
curl.x86_64                     7.76.1-23.el9                  @baseos
thislinehasnospaces
   only-one-field
bash.x86_64                     5.1.8-6.el9                    @anaconda
`
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (skip malformed), got %d", len(pkgs))
	}
}

func TestParseListNoHeader(t *testing.T) {
	// Lines that look like packages without the "Installed Packages" header
	input := `curl.x86_64                     7.76.1-23.el9                  @baseos
bash.x86_64                     5.1.8-6.el9                    @anaconda
`
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
}

func TestParseListTwoColumns(t *testing.T) {
	// Only name.arch and version, no repo column
	input := `Installed Packages
curl.x86_64                     7.76.1-23.el9
`
	pkgs := parseList(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Repository != "" {
		t.Errorf("Repository = %q, want empty", pkgs[0].Repository)
	}
}

func TestParseSearchEmpty(t *testing.T) {
	pkgs := parseSearch("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty input, got %d", len(pkgs))
	}
}

func TestParseSearchSingleResult(t *testing.T) {
	input := `=== Name Exactly Matched: curl ===
curl.x86_64 : A utility for getting files from remote servers
`
	pkgs := parseSearch(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" {
		t.Errorf("Name = %q, want curl", pkgs[0].Name)
	}
}

func TestParseSearchMalformedLines(t *testing.T) {
	input := `=== Name Matched ===
curl.x86_64 : A utility
no-separator-here
another.line.without : proper : colons
bash.noarch : Shell
`
	pkgs := parseSearch(input)
	// "curl.x86_64 : A utility" and "another.line.without : proper : colons" and "bash.noarch : Shell"
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
}

func TestParseInfoEmpty(t *testing.T) {
	p := parseInfo("")
	if p != nil {
		t.Errorf("expected nil from empty input, got %+v", p)
	}
}

func TestParseInfoNoName(t *testing.T) {
	input := `Version      : 1.0
Architecture : x86_64
`
	p := parseInfo(input)
	if p != nil {
		t.Errorf("expected nil when no Name field, got %+v", p)
	}
}

func TestParseInfoReleaseBeforeVersion(t *testing.T) {
	// Release without prior Version should not panic
	input := `Name    : test
Release : 1.el9
Version : 2.0
`
	p := parseInfo(input)
	if p == nil {
		t.Fatal("expected non-nil package")
	}
	// Release came before Version was set, so it won't append properly,
	// but Version should at least be set
	if p.Name != "test" {
		t.Errorf("Name = %q, want test", p.Name)
	}
}

func TestParseInfoFromRepo(t *testing.T) {
	input := `Name         : bash
Version      : 5.1.8
Release      : 6.el9
From repo    : baseos
Summary      : The GNU Bourne Again shell
`
	p := parseInfo(input)
	if p == nil {
		t.Fatal("expected non-nil package")
	}
	if p.Repository != "baseos" {
		t.Errorf("Repository = %q, want baseos", p.Repository)
	}
}

func TestParseVersionLockEmpty(t *testing.T) {
	pkgs := parseVersionLock("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty input, got %d", len(pkgs))
	}
}

func TestParseVersionLockSingleEntry(t *testing.T) {
	input := `nginx-0:1.20.1-14.el9_2.1.*
`
	pkgs := parseVersionLock(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "nginx" {
		t.Errorf("Name = %q, want nginx", pkgs[0].Name)
	}
}

func TestParseRepoListEmpty(t *testing.T) {
	repos := parseRepoList("")
	if len(repos) != 0 {
		t.Errorf("expected 0 repos from empty input, got %d", len(repos))
	}
}

func TestParseRepoListSingleRepo(t *testing.T) {
	input := `repo id                              repo name                                        status
baseos                               CentOS Stream 9 - BaseOS                         enabled
`
	repos := parseRepoList(input)
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].ID != "baseos" || !repos[0].Enabled {
		t.Errorf("unexpected repo: %+v", repos[0])
	}
}

func TestParseGroupListEmpty(t *testing.T) {
	groups := parseGroupList("")
	if len(groups) != 0 {
		t.Errorf("expected 0 groups from empty input, got %d", len(groups))
	}
}

func TestParseGroupInfoEmpty(t *testing.T) {
	pkgs := parseGroupInfo("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty input, got %d", len(pkgs))
	}
}

func TestParseGroupInfoWithMarks(t *testing.T) {
	input := `Group: Web Server
 Mandatory Packages:
   = httpd
   + mod_ssl
   - php
`
	pkgs := parseGroupInfo(input)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	names := map[string]bool{}
	for _, p := range pkgs {
		names[p.Name] = true
	}
	for _, want := range []string{"httpd", "mod_ssl", "php"} {
		if !names[want] {
			t.Errorf("missing package %q", want)
		}
	}
}

func TestParseGroupIsInstalledEmpty(t *testing.T) {
	if parseGroupIsInstalled("", "anything") {
		t.Error("expected false for empty input")
	}
}

func TestNormalizeNameEdgeCases(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"", ""},
		{"pkg.unknown.ext", "pkg.unknown.ext"},
		{"name.with.dots.x86_64", "name.with.dots"},
		{"python3.11", "python3.11"},
		{"glibc.s390x", "glibc"},
		{"kernel.src", "kernel"},
		{".x86_64", ""},
		{"pkg.ppc64le", "pkg"},
		{"pkg.armv7hl", "pkg"},
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

func TestParseArchEdgeCases(t *testing.T) {
	tests := []struct {
		input, wantName, wantArch string
	}{
		{"", "", ""},
		{"pkg.i386", "pkg", "i386"},
		{"pkg.ppc64le", "pkg", "ppc64le"},
		{"pkg.s390x", "pkg", "s390x"},
		{"pkg.armv7hl", "pkg", "armv7hl"},
		{"pkg.src", "pkg", "src"},
		{"pkg.unknown", "pkg.unknown", ""},
		{"name.with.many.dots.noarch", "name.with.many.dots", "noarch"},
		{".noarch", "", "noarch"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, arch := parseArch(tt.input)
			if name != tt.wantName || arch != tt.wantArch {
				t.Errorf("parseArch(%q) = (%q, %q), want (%q, %q)",
					tt.input, name, arch, tt.wantName, tt.wantArch)
			}
		})
	}
}

// --- dnf5 edge case tests ---

func TestStripPreambleEmpty(t *testing.T) {
	got := stripPreamble("")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestStripPreambleNoPreamble(t *testing.T) {
	input := "Installed packages\nbash.x86_64   5.3.0-2.fc43   abc\n"
	got := stripPreamble(input)
	if got != input {
		t.Errorf("expected unchanged output when no preamble present")
	}
}

func TestParseListDNF5Empty(t *testing.T) {
	pkgs := parseListDNF5("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty input, got %d", len(pkgs))
	}
}

func TestParseListDNF5SinglePackage(t *testing.T) {
	input := `Installed packages
curl.aarch64   7.76.1-23.el9   abc123
`
	pkgs := parseListDNF5(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" || pkgs[0].Arch != "aarch64" {
		t.Errorf("unexpected: %+v", pkgs[0])
	}
}

func TestParseSearchDNF5Empty(t *testing.T) {
	pkgs := parseSearchDNF5("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty input, got %d", len(pkgs))
	}
}

func TestParseInfoDNF5Empty(t *testing.T) {
	p := parseInfoDNF5("")
	if p != nil {
		t.Errorf("expected nil from empty input, got %+v", p)
	}
}

func TestParseInfoDNF5NoName(t *testing.T) {
	input := `Version : 1.0
Architecture : x86_64
`
	p := parseInfoDNF5(input)
	if p != nil {
		t.Errorf("expected nil when no Name field, got %+v", p)
	}
}

func TestParseGroupListDNF5Empty(t *testing.T) {
	groups := parseGroupListDNF5("")
	if len(groups) != 0 {
		t.Errorf("expected 0 groups from empty input, got %d", len(groups))
	}
}

func TestParseGroupIsInstalledDNF5Empty(t *testing.T) {
	if parseGroupIsInstalledDNF5("", "anything") {
		t.Error("expected false for empty input")
	}
}

func TestParseVersionLockDNF5Empty(t *testing.T) {
	pkgs := parseVersionLockDNF5("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty input, got %d", len(pkgs))
	}
}

func TestParseVersionLockDNF5SingleEntry(t *testing.T) {
	input := `# Added by 'versionlock add' command on 2026-02-26 03:14:29
Package name: nginx
evr = 1.20.1-14.el9
`
	pkgs := parseVersionLockDNF5(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "nginx" {
		t.Errorf("Name = %q, want nginx", pkgs[0].Name)
	}
}

func TestParseRepoListDNF5Empty(t *testing.T) {
	repos := parseRepoListDNF5("")
	if len(repos) != 0 {
		t.Errorf("expected 0 repos from empty input, got %d", len(repos))
	}
}

func TestParseGroupInfoDNF5Empty(t *testing.T) {
	pkgs := parseGroupInfoDNF5("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty input, got %d", len(pkgs))
	}
}

func TestParseGroupInfoDNF5SinglePackage(t *testing.T) {
	input := `Id                   : test-group
Name                 : Test
Mandatory packages   : single-pkg
`
	pkgs := parseGroupInfoDNF5(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "single-pkg" {
		t.Errorf("Name = %q, want single-pkg", pkgs[0].Name)
	}
}
