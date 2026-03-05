package apt

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseList(t *testing.T) {
	input := "bash\t5.2-1\tGNU Bourne Again SHell\ncoreutils\t9.1-1\tGNU core utilities\n"
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || pkgs[0].Version != "5.2-1" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseSearch(t *testing.T) {
	input := "vim - Vi IMproved\nnano - small text editor\n"
	pkgs := parseSearch(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "vim" || pkgs[0].Description != "Vi IMproved" {
		t.Errorf("unexpected package: %+v", pkgs[0])
	}
}

func TestParseInfo(t *testing.T) {
	input := `Package: bash
Version: 5.2-1
Architecture: amd64
Description: GNU Bourne Again SHell
`
	p, err := parseInfo(input)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "bash" || p.Version != "5.2-1" || p.Arch != "amd64" {
		t.Errorf("unexpected package: %+v", p)
	}
}

func TestParseInfoEmpty(t *testing.T) {
	_, err := parseInfo("")
	if err != snack.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestParseListEmpty(t *testing.T) {
	pkgs := parseList("")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestNew(t *testing.T) {
	a := New()
	if a.Name() != "apt" {
		t.Errorf("expected 'apt', got %q", a.Name())
	}
}

func TestSupportsDryRun(t *testing.T) {
	a := New()
	if !a.SupportsDryRun() {
		t.Error("expected SupportsDryRun() = true")
	}
}

func TestCapabilities(t *testing.T) {
	caps := snack.GetCapabilities(New())
	checks := []struct {
		name string
		got  bool
		want bool
	}{
		{"VersionQuery", caps.VersionQuery, true},
		{"Hold", caps.Hold, true},
		{"Clean", caps.Clean, true},
		{"FileOwnership", caps.FileOwnership, true},
		{"RepoManagement", caps.RepoManagement, true},
		{"KeyManagement", caps.KeyManagement, true},
		{"Groups", caps.Groups, false},
		{"NameNormalize", caps.NameNormalize, true},
		{"DryRun", caps.DryRun, true},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if c.got != c.want {
				t.Errorf("Capabilities.%s = %v, want %v", c.name, c.got, c.want)
			}
		})
	}
}

// Compile-time interface checks.
var (
	_ snack.Manager         = (*Apt)(nil)
	_ snack.VersionQuerier  = (*Apt)(nil)
	_ snack.Holder          = (*Apt)(nil)
	_ snack.Cleaner         = (*Apt)(nil)
	_ snack.FileOwner       = (*Apt)(nil)
	_ snack.RepoManager     = (*Apt)(nil)
	_ snack.KeyManager      = (*Apt)(nil)
	_ snack.NameNormalizer  = (*Apt)(nil)
	_ snack.DryRunner       = (*Apt)(nil)
	_ snack.PackageUpgrader = (*Apt)(nil)
)
