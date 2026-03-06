package dpkg

import "testing"

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"curl", "curl"},
		{"curl:amd64", "curl"},
		{"bash:arm64", "bash"},
		{"python3:i386", "python3"},
		{"libc6:armhf", "libc6"},
		{"pkg:armel", "pkg"},
		{"pkg:mips", "pkg"},
		{"pkg:mipsel", "pkg"},
		{"pkg:mips64el", "pkg"},
		{"pkg:ppc64el", "pkg"},
		{"pkg:s390x", "pkg"},
		{"pkg:all", "pkg"},
		{"pkg:any", "pkg"},
		// Unknown arch suffix should be kept
		{"pkg:unknown", "pkg:unknown"},
		{"libstdc++6:amd64", "libstdc++6"},
		{"", ""},
		// Multiple colons — only last one checked
		{"a:b:amd64", "a:b"},
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
	tests := []struct {
		input    string
		wantName string
		wantArch string
	}{
		{"curl:amd64", "curl", "amd64"},
		{"bash:arm64", "bash", "arm64"},
		{"python3", "python3", ""},
		{"curl:i386", "curl", "i386"},
		{"pkg:armhf", "pkg", "armhf"},
		{"pkg:armel", "pkg", "armel"},
		{"pkg:mips", "pkg", "mips"},
		{"pkg:mipsel", "pkg", "mipsel"},
		{"pkg:mips64el", "pkg", "mips64el"},
		{"pkg:ppc64el", "pkg", "ppc64el"},
		{"pkg:s390x", "pkg", "s390x"},
		{"pkg:all", "pkg", "all"},
		{"pkg:any", "pkg", "any"},
		// Unknown arch — not split
		{"pkg:foobar", "pkg:foobar", ""},
		{"", "", ""},
		// Multiple colons
		{"a:b:arm64", "a:b", "arm64"},
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
