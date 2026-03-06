package main

import (
	"testing"

	"github.com/gogrlx/snack"
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

	// Explicit override
	flagMgr = "apt"
	m, err = getManager()
	if err != nil {
		t.Fatalf("getManager() with --manager=apt failed: %v", err)
	}
	if m.Name() != "apt" {
		t.Errorf("expected Name()=apt, got %q", m.Name())
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
