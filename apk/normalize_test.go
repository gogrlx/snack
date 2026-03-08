package apk

import (
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"curl", "curl"},
		{"curl-x86_64", "curl"},
		{"openssl-aarch64", "openssl"},
		{"musl-armhf", "musl"},
		{"busybox-armv7", "busybox"},
		{"lib-ssl-dev-x86", "lib-ssl-dev"},
		{"zlib-ppc64le", "zlib"},
		{"kernel-s390x", "kernel"},
		{"toolchain-riscv64", "toolchain"},
		{"app-loongarch64", "app"},
		// No arch suffix — unchanged
		{"python", "python"},
		{"go", "go"},
		{"", ""},
		// Suffix that isn't an arch — unchanged
		{"my-pkg-foo", "my-pkg-foo"},
		{"libfoo-1.0", "libfoo-1.0"},
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

func TestParseArchNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantArch string
	}{
		{"x86_64", "curl-x86_64", "curl", "x86_64"},
		{"x86", "musl-x86", "musl", "x86"},
		{"aarch64", "openssl-aarch64", "openssl", "aarch64"},
		{"armhf", "busybox-armhf", "busybox", "armhf"},
		{"armv7", "lib-armv7", "lib", "armv7"},
		{"ppc64le", "app-ppc64le", "app", "ppc64le"},
		{"s390x", "pkg-s390x", "pkg", "s390x"},
		{"riscv64", "tool-riscv64", "tool", "riscv64"},
		{"loongarch64", "gcc-loongarch64", "gcc", "loongarch64"},
		{"no arch", "curl", "curl", ""},
		{"unknown suffix", "pkg-foobar", "pkg-foobar", ""},
		{"empty", "", "", ""},
		{"hyphen but not arch", "lib-ssl-dev", "lib-ssl-dev", ""},
		{"multi hyphen with arch", "lib-ssl-dev-x86_64", "lib-ssl-dev", "x86_64"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotArch := parseArchNormalize(tt.input)
			if gotName != tt.wantName || gotArch != tt.wantArch {
				t.Errorf("parseArchNormalize(%q) = (%q, %q), want (%q, %q)",
					tt.input, gotName, gotArch, tt.wantName, tt.wantArch)
			}
		})
	}
}
