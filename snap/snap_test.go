package snap

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseSnapList(t *testing.T) {
	input := `Name      Version    Rev    Tracking       Publisher   Notes
core22    20240111   1122   latest/stable  canonical✓  base
firefox   131.0      4647   latest/stable  mozilla✓    -
`
	pkgs := parseSnapList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "core22" || pkgs[0].Version != "20240111" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[1].Name != "firefox" || pkgs[1].Version != "131.0" {
		t.Errorf("unexpected second package: %+v", pkgs[1])
	}
	if !pkgs[0].Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseSnapListEmpty(t *testing.T) {
	input := `Name  Version  Rev  Tracking  Publisher  Notes
`
	pkgs := parseSnapList(input)
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseSnapFind(t *testing.T) {
	input := `Name      Version   Publisher    Notes  Summary
firefox   131.0     mozilla✓     -      Mozilla Firefox web browser
chromium  129.0     nickvdp      -      Chromium web browser
`
	pkgs := parseSnapFind(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "firefox" || pkgs[0].Version != "131.0" {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[0].Description != "Mozilla Firefox web browser" {
		t.Errorf("unexpected description: %q", pkgs[0].Description)
	}
}

func TestParseSnapInfo(t *testing.T) {
	input := `name:      firefox
summary:   Mozilla Firefox web browser
publisher: Mozilla✓ (mozilla✓)
snap-id:   3wdHCAVyZEmYsCMFDE9qt92UV8rC8Wdk
installed: 131.0 (4647) 283MB -
`
	pkg := parseSnapInfo(input)
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.Name != "firefox" {
		t.Errorf("expected name 'firefox', got %q", pkg.Name)
	}
	if pkg.Version != "131.0" {
		t.Errorf("expected version '131.0', got %q", pkg.Version)
	}
	if pkg.Description != "Mozilla Firefox web browser" {
		t.Errorf("unexpected description: %q", pkg.Description)
	}
	if !pkg.Installed {
		t.Error("expected Installed=true")
	}
}

func TestParseSnapInfoEmpty(t *testing.T) {
	pkg := parseSnapInfo("")
	if pkg != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseSnapInfoVersion(t *testing.T) {
	input := `name:      firefox
channels:
  latest/stable:    131.0       2024-10-01 (4647) 283MB -
  latest/candidate: 132.0b5     2024-10-05 (4650) 285MB -
  latest/beta:      132.0b5     2024-10-05 (4650) 285MB -
  latest/edge:      133.0a1     2024-10-06 (4655) 290MB -
`
	ver := parseSnapInfoVersion(input)
	if ver != "131.0" {
		t.Errorf("expected '131.0', got %q", ver)
	}
}

func TestParseSnapInfoVersionMissing(t *testing.T) {
	ver := parseSnapInfoVersion("name: test\n")
	if ver != "" {
		t.Errorf("expected empty, got %q", ver)
	}
}

func TestParseSnapRefreshList(t *testing.T) {
	input := `Name      Version   Rev   Publisher    Notes
firefox   132.0     4650  mozilla✓     -
`
	pkgs := parseSnapRefreshList(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "firefox" || pkgs[0].Version != "132.0" {
		t.Errorf("unexpected package: %+v", pkgs[0])
	}
}

func TestParseSnapRefreshListUpToDate(t *testing.T) {
	input := `Name  Version  Rev  Publisher  Notes
All snaps up to date.
`
	pkgs := parseSnapRefreshList(input)
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestSemverCmp(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.2.3", "1.2.4", -1},
		{"1.10.0", "1.9.0", 1},
		{"1.0", "1.0.0", 0},
		{"131.0", "132.0", -1},
	}
	for _, tt := range tests {
		got := semverCmp(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("semverCmp(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ snack.Manager = (*Snap)(nil)
	var _ snack.VersionQuerier = (*Snap)(nil)
}

func TestName(t *testing.T) {
	s := New()
	if s.Name() != "snap" {
		t.Errorf("Name() = %q, want %q", s.Name(), "snap")
	}
}
