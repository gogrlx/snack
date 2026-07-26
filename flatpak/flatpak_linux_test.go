//go:build linux

package flatpak

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogrlx/snack"
)

func TestAvailableWithFakeFlatpak(t *testing.T) {
	writeFakeFlatpak(t, "exit 0\n")

	if !available() {
		t.Fatal("expected fake flatpak to be available")
	}
}

func TestListReposUsesFlatpakRemotesOutput(t *testing.T) {
	writeFakeFlatpak(t, `if [ "$1" = "remotes" ]; then
	printf 'flathub\thttps://dl.flathub.org/repo/\t\n'
	printf 'nightly\thttps://nightly.gnome.org/repo/\tdisabled\n'
	exit 0
fi
exit 2
`)

	repos, err := New().ListRepos(context.Background())
	if err != nil {
		t.Fatalf("ListRepos returned error: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].ID != "flathub" || repos[0].URL != "https://dl.flathub.org/repo/" || !repos[0].Enabled {
		t.Errorf("unexpected first repo: %+v", repos[0])
	}
	if repos[1].ID != "nightly" || repos[1].Enabled {
		t.Errorf("unexpected second repo: %+v", repos[1])
	}
}

func TestRunMapsPermissionDenied(t *testing.T) {
	writeFakeFlatpak(t, `printf 'permission denied\n' >&2
exit 1
`)

	_, err := run(context.Background(), []string{"install", "flathub", "org.example.App"})
	if !errors.Is(err, snack.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
}

func writeFakeFlatpak(t *testing.T, body string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "flatpak")
	script := "#!/bin/sh\n" + body
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake flatpak: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
