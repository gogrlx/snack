//go:build linux

package pacman

import (
	"testing"

	"github.com/gogrlx/snack"
)

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		base     []string
		opts     snack.Options
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "basic",
			base:     []string{"-S", "vim"},
			opts:     snack.Options{},
			wantCmd:  "pacman",
			wantArgs: []string{"-S", "vim"},
		},
		{
			name:     "with sudo",
			base:     []string{"-S", "vim"},
			opts:     snack.Options{Sudo: true},
			wantCmd:  "sudo",
			wantArgs: []string{"pacman", "-S", "vim"},
		},
		{
			name:     "with root and noconfirm",
			base:     []string{"-S", "vim"},
			opts:     snack.Options{Root: "/mnt", AssumeYes: true},
			wantCmd:  "pacman",
			wantArgs: []string{"-r", "/mnt", "-S", "vim", "--noconfirm"},
		},
		{
			name:     "dry run",
			base:     []string{"-S", "vim"},
			opts:     snack.Options{DryRun: true},
			wantCmd:  "pacman",
			wantArgs: []string{"-S", "vim", "--print"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := buildArgs(tt.base, tt.opts)
			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}
