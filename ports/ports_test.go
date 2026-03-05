package ports

import (
	"testing"

	"github.com/gogrlx/snack"
)

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

func TestParseListEmpty(t *testing.T) {
	pkgs := parseList("")
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseListWhitespaceOnly(t *testing.T) {
	pkgs := parseList("   \n  \n\n")
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseListNoDescription(t *testing.T) {
	input := "vim-9.0.2100\n"
	pkgs := parseList(input)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "vim" || pkgs[0].Version != "9.0.2100" {
		t.Errorf("unexpected package: %+v", pkgs[0])
	}
	if pkgs[0].Description != "" {
		t.Errorf("expected empty description, got %q", pkgs[0].Description)
	}
}

func TestParseSearchResultsEmpty(t *testing.T) {
	pkgs := parseSearchResults("")
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseInfoOutputEmpty(t *testing.T) {
	pkg := parseInfoOutput("", "")
	if pkg != nil {
		t.Error("expected nil for empty input and empty pkg name")
	}
}

func TestParseInfoOutputFallbackToPkgArg(t *testing.T) {
	// No "Information for" header — should fall back to parsing the pkg argument.
	input := "Some random output\nwithout the expected header"
	pkg := parseInfoOutput(input, "curl-8.5.0")
	if pkg == nil {
		t.Fatal("expected non-nil package from fallback")
	}
	if pkg.Name != "curl" || pkg.Version != "8.5.0" {
		t.Errorf("unexpected fallback parse: %+v", pkg)
	}
}

func TestSplitNameVersionNoHyphen(t *testing.T) {
	name, ver := splitNameVersion("singleword")
	if name != "singleword" || ver != "" {
		t.Errorf("splitNameVersion(\"singleword\") = (%q, %q), want (\"singleword\", \"\")", name, ver)
	}
}

func TestSplitNameVersionLeadingHyphen(t *testing.T) {
	// A hyphen at position 0 should return the whole string as name.
	name, ver := splitNameVersion("-1.0")
	if name != "" || ver != "1.0" {
		// LastIndex("-1.0", "-") is 0, and idx <= 0 returns (s, "")
		// Actually idx=0 means the condition idx <= 0 is true
	}
	// Re-check: idx=0, condition is idx <= 0, so returns (s, "")
	name, ver = splitNameVersion("-1.0")
	if name != "-1.0" || ver != "" {
		t.Errorf("splitNameVersion(\"-1.0\") = (%q, %q), want (\"-1.0\", \"\")", name, ver)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ snack.Manager = (*Ports)(nil)
}

func TestName(t *testing.T) {
	p := New()
	if p.Name() != "ports" {
		t.Errorf("Name() = %q, want %q", p.Name(), "ports")
	}
}
