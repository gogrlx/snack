package dpkg

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestParseList(t *testing.T) {
	input := "bash\t5.2-1\tinstall ok installed\ncoreutils\t9.1-1\tdeinstall ok config-files\n"
	pkgs := parseList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || !pkgs[0].Installed {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[1].Installed {
		t.Errorf("expected second package not installed: %+v", pkgs[1])
	}
}

func TestParseDpkgList(t *testing.T) {
	input := `Desired=Unknown/Install/Remove/Purge/Hold
| Status=Not/Inst/Conf-files/Unpacked/halF-conf/Half-inst/trig-aWait/Trig-pend
|/ Err?=(none)/Reinst-required (Status,Err: uppercase=bad)
||/ Name           Version      Architecture Description
+++-==============-============-============-=================================
ii  bash           5.2-1        amd64        GNU Bourne Again SHell
rc  oldpkg         1.0-1        amd64        Some old package
`
	pkgs := parseDpkgList(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || !pkgs[0].Installed {
		t.Errorf("unexpected: %+v", pkgs[0])
	}
	if pkgs[1].Installed {
		t.Errorf("expected oldpkg not installed: %+v", pkgs[1])
	}
}

func TestParseInfo(t *testing.T) {
	input := `Package: bash
Status: install ok installed
Version: 5.2-1
Architecture: amd64
Description: GNU Bourne Again SHell
`
	p, err := parseInfo(input)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "bash" || p.Version != "5.2-1" || !p.Installed {
		t.Errorf("unexpected: %+v", p)
	}
}

func TestParseInfoEmpty(t *testing.T) {
	_, err := parseInfo("")
	if err != snack.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNew(t *testing.T) {
	d := New()
	if d.Name() != "dpkg" {
		t.Errorf("expected 'dpkg', got %q", d.Name())
	}
}

func TestUpgradeUnsupported(t *testing.T) {
	d := New()
	if err := d.Upgrade(nil); err != snack.ErrUnsupportedPlatform {
		t.Errorf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

func TestUpdateUnsupported(t *testing.T) {
	d := New()
	if err := d.Update(nil); err != snack.ErrUnsupportedPlatform {
		t.Errorf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

// Verify Dpkg implements snack.Manager at compile time.
var _ snack.Manager = (*Dpkg)(nil)
