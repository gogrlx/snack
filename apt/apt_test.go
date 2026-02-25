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

// Verify Apt implements snack.Manager at compile time.
var _ snack.Manager = (*Apt)(nil)
