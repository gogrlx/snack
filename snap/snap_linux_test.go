//go:build linux

package snap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAvailableRequiresSnapVersion(t *testing.T) {
	writeFakeSnap(t, `if [ "$1" = "version" ]; then
	exit 0
fi
exit 2
`)

	if !available() {
		t.Fatal("expected fake snap with version support to be available")
	}
}

func TestListUpgradesTreatsAllUpToDateErrorAsEmpty(t *testing.T) {
	writeFakeSnap(t, `if [ "$1" = "refresh" ] && [ "$2" = "--list" ]; then
	printf 'All snaps up to date.\n' >&2
	exit 1
fi
exit 2
`)

	upgrades, err := New().ListUpgrades(context.Background())
	if err != nil {
		t.Fatalf("ListUpgrades returned error: %v", err)
	}
	if len(upgrades) != 0 {
		t.Fatalf("expected no upgrades, got %d", len(upgrades))
	}
}

func TestCleanRemovesDisabledRevisions(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "snap.log")
	writeFakeSnap(t, `if [ "$1" = "list" ] && [ "$2" = "--all" ]; then
	printf 'Name Version Rev Tracking Publisher Notes\n'
	printf 'core22 20240111 1122 latest/stable canonical disabled\n'
	printf 'firefox 131.0 4647 latest/stable mozilla -\n'
	exit 0
fi
if [ "$1" = "remove" ]; then
	printf '%s %s %s\n' "$1" "$2" "$3" >> "$SNAP_TEST_LOG"
	exit 0
fi
exit 2
`)
	t.Setenv("SNAP_TEST_LOG", logPath)

	if err := New().Clean(context.Background()); err != nil {
		t.Fatalf("Clean returned error: %v", err)
	}

	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read snap command log: %v", err)
	}
	got := strings.TrimSpace(string(logBytes))
	if got != "remove core22 --revision=1122" {
		t.Fatalf("unexpected snap remove call %q", got)
	}
}

func writeFakeSnap(t *testing.T, body string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "snap")
	script := "#!/bin/sh\n" + body
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake snap: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
