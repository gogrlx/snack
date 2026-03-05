//go:build linux

package apt

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestFormatTargets(t *testing.T) {
	tests := []struct {
		name    string
		targets []snack.Target
		want    []string
	}{
		{
			name:    "empty",
			targets: nil,
			want:    []string{},
		},
		{
			name:    "name_only",
			targets: []snack.Target{{Name: "curl"}},
			want:    []string{"curl"},
		},
		{
			name:    "with_version",
			targets: []snack.Target{{Name: "curl", Version: "7.88"}},
			want:    []string{"curl=7.88"},
		},
		{
			name: "mixed",
			targets: []snack.Target{
				{Name: "curl"},
				{Name: "bash", Version: "5.2-1"},
				{Name: "vim"},
			},
			want: []string{"curl", "bash=5.2-1", "vim"},
		},
		{
			name:    "version_with_epoch",
			targets: []snack.Target{{Name: "systemd", Version: "1:252-2"}},
			want:    []string{"systemd=1:252-2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTargets(tt.targets)
			if len(got) != len(tt.want) {
				t.Fatalf("formatTargets() returned %d args, want %d", len(got), len(tt.want))
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("formatTargets()[%d] = %q, want %q", i, g, tt.want[i])
				}
			}
		})
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		pkgs []snack.Target
		opts []snack.Option
		want []string
	}{
		{
			name: "basic_install",
			cmd:  "install",
			pkgs: []snack.Target{{Name: "curl"}},
			want: []string{"apt-get", "install", "curl"},
		},
		{
			name: "install_with_yes",
			cmd:  "install",
			pkgs: []snack.Target{{Name: "curl"}},
			opts: []snack.Option{snack.WithAssumeYes()},
			want: []string{"apt-get", "install", "-y", "curl"},
		},
		{
			name: "install_with_dry_run",
			cmd:  "install",
			pkgs: []snack.Target{{Name: "curl"}},
			opts: []snack.Option{snack.WithDryRun()},
			want: []string{"apt-get", "install", "--dry-run", "curl"},
		},
		{
			name: "install_with_sudo",
			cmd:  "install",
			pkgs: []snack.Target{{Name: "curl"}},
			opts: []snack.Option{snack.WithSudo()},
			want: []string{"sudo", "apt-get", "install", "curl"},
		},
		{
			name: "install_with_reinstall",
			cmd:  "install",
			pkgs: []snack.Target{{Name: "curl"}},
			opts: []snack.Option{snack.WithReinstall()},
			want: []string{"apt-get", "install", "--reinstall", "curl"},
		},
		{
			name: "remove_no_reinstall_flag",
			cmd:  "remove",
			pkgs: []snack.Target{{Name: "curl"}},
			opts: []snack.Option{snack.WithReinstall()},
			want: []string{"apt-get", "remove", "curl"}, // --reinstall only on install
		},
		{
			name: "install_from_repo",
			cmd:  "install",
			pkgs: []snack.Target{{Name: "curl"}},
			opts: []snack.Option{snack.WithFromRepo("stable")},
			want: []string{"apt-get", "install", "-t", "stable", "curl"},
		},
		{
			name: "all_options",
			cmd:  "install",
			pkgs: []snack.Target{{Name: "curl", Version: "7.88"}},
			opts: []snack.Option{snack.WithSudo(), snack.WithAssumeYes(), snack.WithDryRun(), snack.WithFromRepo("sid"), snack.WithReinstall()},
			want: []string{"sudo", "apt-get", "install", "-y", "--dry-run", "-t", "sid", "--reinstall", "curl=7.88"},
		},
		{
			name: "multiple_packages",
			cmd:  "install",
			pkgs: []snack.Target{{Name: "curl"}, {Name: "wget"}, {Name: "bash", Version: "5.2"}},
			want: []string{"apt-get", "install", "curl", "wget", "bash=5.2"},
		},
		{
			name: "upgrade_no_packages",
			cmd:  "upgrade",
			pkgs: nil,
			opts: []snack.Option{snack.WithAssumeYes()},
			want: []string{"apt-get", "upgrade", "-y"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildArgs(tt.cmd, tt.pkgs, tt.opts...)
			if len(got) != len(tt.want) {
				t.Fatalf("buildArgs() = %v (len %d), want %v (len %d)", got, len(got), tt.want, len(tt.want))
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("buildArgs()[%d] = %q, want %q", i, g, tt.want[i])
				}
			}
		})
	}
}

func TestExtractURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic_deb",
			input: "deb http://archive.ubuntu.com/ubuntu/ jammy main",
			want:  "http://archive.ubuntu.com/ubuntu/",
		},
		{
			name:  "deb_src",
			input: "deb-src http://archive.ubuntu.com/ubuntu/ jammy main",
			want:  "http://archive.ubuntu.com/ubuntu/",
		},
		{
			name:  "with_options",
			input: "deb [arch=amd64] https://apt.example.com/repo stable main",
			want:  "https://apt.example.com/repo",
		},
		{
			name:  "with_signed_by",
			input: "deb [arch=amd64 signed-by=/etc/apt/keyrings/key.gpg] https://repo.example.com/deb stable main",
			want:  "https://repo.example.com/deb",
		},
		{
			name:  "just_type",
			input: "deb",
			want:  "",
		},
		{
			name:  "empty_options_bracket",
			input: "deb [] http://example.com/repo stable",
			want:  "http://example.com/repo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractURL(tt.input)
			if got != tt.want {
				t.Errorf("extractURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
