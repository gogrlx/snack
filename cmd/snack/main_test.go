package main

import (
	"testing"

	"github.com/gogrlx/snack"
	"github.com/spf13/cobra"
)

func TestTargets(t *testing.T) {
	tests := []struct {
		name string
		args []string
		ver  string
		want []snack.Target
	}{
		{
			name: "no_args",
			args: nil,
			want: nil,
		},
		{
			name: "single_no_version",
			args: []string{"curl"},
			want: []snack.Target{{Name: "curl"}},
		},
		{
			name: "multiple_no_version",
			args: []string{"curl", "wget"},
			want: []snack.Target{{Name: "curl"}, {Name: "wget"}},
		},
		{
			name: "single_with_version",
			args: []string{"curl"},
			ver:  "7.88",
			want: []snack.Target{{Name: "curl", Version: "7.88"}},
		},
		{
			name: "multiple_with_version",
			args: []string{"curl", "wget"},
			ver:  "1.0",
			want: []snack.Target{{Name: "curl", Version: "1.0"}, {Name: "wget", Version: "1.0"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := targets(tt.args, tt.ver)
			if len(got) != len(tt.want) {
				t.Fatalf("targets() returned %d, want %d", len(got), len(tt.want))
			}
			for i, g := range got {
				if g.Name != tt.want[i].Name {
					t.Errorf("[%d] Name = %q, want %q", i, g.Name, tt.want[i].Name)
				}
				if g.Version != tt.want[i].Version {
					t.Errorf("[%d] Version = %q, want %q", i, g.Version, tt.want[i].Version)
				}
			}
		})
	}
}

func TestOpts(t *testing.T) {
	// Reset flags
	flagSudo = false
	flagYes = false
	flagDry = false

	o := opts()
	if len(o) != 0 {
		t.Errorf("expected 0 options with no flags, got %d", len(o))
	}

	flagSudo = true
	o = opts()
	if len(o) != 1 {
		t.Errorf("expected 1 option with sudo, got %d", len(o))
	}

	flagYes = true
	flagDry = true
	o = opts()
	if len(o) != 3 {
		t.Errorf("expected 3 options with all flags, got %d", len(o))
	}

	// Clean up
	flagSudo = false
	flagYes = false
	flagDry = false
}

func TestOptsApply(t *testing.T) {
	flagSudo = true
	flagYes = true
	flagDry = true
	defer func() {
		flagSudo = false
		flagYes = false
		flagDry = false
	}()

	applied := snack.ApplyOptions(opts()...)
	if !applied.Sudo {
		t.Error("expected Sudo=true")
	}
	if !applied.AssumeYes {
		t.Error("expected AssumeYes=true")
	}
	if !applied.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestGetManager(t *testing.T) {
	// Default detection
	flagMgr = ""
	m, err := getManager()
	if err != nil {
		t.Skipf("no manager available: %v", err)
	}
	if m.Name() == "" {
		t.Error("expected non-empty manager name")
	}

	// Explicit override - use the detected manager's name
	// since not all managers are available on all platforms
	flagMgr = m.Name()
	m2, err := getManager()
	if err != nil {
		t.Fatalf("getManager() with --manager=%s failed: %v", flagMgr, err)
	}
	if m2.Name() != flagMgr {
		t.Errorf("expected Name()=%s, got %q", flagMgr, m2.Name())
	}

	// Unknown manager
	flagMgr = "nonexistent-manager-xyz"
	_, err = getManager()
	if err == nil {
		t.Error("expected error for unknown manager")
	}

	flagMgr = ""
}

func TestVersionString(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestNewRootCmdRegistersPersistentFlags(t *testing.T) {
	cmd := newRootCmd()

	for _, flagName := range []string{"manager", "sudo", "yes", "dry-run"} {
		if cmd.PersistentFlags().Lookup(flagName) == nil {
			t.Fatalf("expected persistent flag %q to be registered", flagName)
		}
	}
}

func TestNewRootCmdRegistersExpectedCommands(t *testing.T) {
	cmd := newRootCmd()

	want := map[string]struct{}{
		"install": {},
		"remove":  {},
		"purge":   {},
		"upgrade": {},
		"update":  {},
		"list":    {},
		"search":  {},
		"info":    {},
		"which":   {},
		"hold":    {},
		"unhold":  {},
		"clean":   {},
		"repo":    {},
		"key":     {},
		"group":   {},
		"detect":  {},
		"version": {},
	}

	if len(cmd.Commands()) != len(want) {
		t.Fatalf("newRootCmd() registered %d commands, want %d", len(cmd.Commands()), len(want))
	}

	for _, subcmd := range cmd.Commands() {
		if _, ok := want[subcmd.Name()]; !ok {
			t.Fatalf("unexpected subcommand %q", subcmd.Name())
		}
		delete(want, subcmd.Name())
	}

	if len(want) != 0 {
		t.Fatalf("missing subcommands: %v", want)
	}
}

func TestNestedCommandRegistration(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		want []string
	}{
		{
			name: "repo",
			cmd:  repoCmd(),
			want: []string{"add", "list", "remove"},
		},
		{
			name: "key",
			cmd:  keyCmd(),
			want: []string{"add", "list", "remove"},
		},
		{
			name: "group",
			cmd:  groupCmd(),
			want: []string{"info", "install", "list"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.cmd.Commands()) != len(tt.want) {
				t.Fatalf("%s registered %d subcommands, want %d", tt.name, len(tt.cmd.Commands()), len(tt.want))
			}
			for _, name := range tt.want {
				if _, _, err := tt.cmd.Find([]string{name}); err != nil {
					t.Fatalf("expected %s to register subcommand %q: %v", tt.name, name, err)
				}
			}
		})
	}
}
