package dnf

import (
	"testing"

	"github.com/gogrlx/snack"
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

// Ensure interface checks from capabilities.go are satisfied.
var (
	_ snack.Manager        = (*DNF)(nil)
	_ snack.VersionQuerier = (*DNF)(nil)
	_ snack.Holder         = (*DNF)(nil)
	_ snack.Cleaner        = (*DNF)(nil)
	_ snack.FileOwner      = (*DNF)(nil)
	_ snack.RepoManager    = (*DNF)(nil)
	_ snack.KeyManager     = (*DNF)(nil)
	_ snack.Grouper        = (*DNF)(nil)
	_ snack.NameNormalizer = (*DNF)(nil)
)
