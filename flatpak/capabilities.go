package flatpak

import (
	"context"

	"github.com/gogrlx/snack"
)

// Compile-time interface checks.
var (
	_ snack.Cleaner     = (*Flatpak)(nil)
	_ snack.RepoManager = (*Flatpak)(nil)
)

// Autoremove removes unused runtimes and extensions.
func (f *Flatpak) Autoremove(ctx context.Context, opts ...snack.Option) error {
	f.Lock()
	defer f.Unlock()
	return autoremove(ctx, opts...)
}

// Clean is a no-op for flatpak (no direct cache clean equivalent).
func (f *Flatpak) Clean(ctx context.Context) error {
	return nil
}

// ListRepos returns all configured remotes.
func (f *Flatpak) ListRepos(ctx context.Context) ([]snack.Repository, error) {
	return listRepos(ctx)
}

// AddRepo adds a new remote.
func (f *Flatpak) AddRepo(ctx context.Context, repo snack.Repository) error {
	f.Lock()
	defer f.Unlock()
	return addRepo(ctx, repo)
}

// RemoveRepo removes a configured remote.
func (f *Flatpak) RemoveRepo(ctx context.Context, id string) error {
	f.Lock()
	defer f.Unlock()
	return removeRepo(ctx, id)
}
