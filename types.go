package snack

// Package represents a system package.
type Package struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Arch        string `json:"arch,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Installed   bool   `json:"installed"`
}

// Options holds configuration for package manager operations.
type Options struct {
	Sudo      bool
	AssumeYes bool
	DryRun    bool
	Root      string // alternate root filesystem
	Verbose   bool
	Refresh   bool   // refresh package index before operation
	FromRepo  string // constrain operation to a specific repository
	Reinstall bool   // reinstall already-installed packages
}

// Option is a functional option for package manager operations.
type Option func(*Options)

// WithSudo runs the command with sudo.
func WithSudo() Option {
	return func(o *Options) { o.Sudo = true }
}

// WithAssumeYes automatically confirms prompts.
func WithAssumeYes() Option {
	return func(o *Options) { o.AssumeYes = true }
}

// WithDryRun simulates the operation without making changes.
func WithDryRun() Option {
	return func(o *Options) { o.DryRun = true }
}

// WithRoot sets an alternate root filesystem path.
func WithRoot(root string) Option {
	return func(o *Options) { o.Root = root }
}

// WithVerbose enables verbose output.
func WithVerbose() Option {
	return func(o *Options) { o.Verbose = true }
}

// WithRefresh refreshes the package database before the operation.
// Equivalent to pacman -Sy, apt-get update, apk update, etc.
func WithRefresh() Option {
	return func(o *Options) { o.Refresh = true }
}

// WithFromRepo constrains the operation to a specific repository.
// e.g. "unstable" for apt, "community" for pacman.
func WithFromRepo(repo string) Option {
	return func(o *Options) { o.FromRepo = repo }
}

// WithReinstall forces reinstallation of already-installed packages.
func WithReinstall() Option {
	return func(o *Options) { o.Reinstall = true }
}

// Repository represents a configured package repository.
type Repository struct {
	ID       string `json:"id"`                  // unique identifier
	Name     string `json:"name,omitempty"`      // human-readable name
	URL      string `json:"url"`                 // repository URL
	Enabled  bool   `json:"enabled"`             // whether the repo is active
	GPGCheck bool   `json:"gpg_check,omitempty"` // whether GPG verification is enabled
	GPGKey   string `json:"gpg_key,omitempty"`   // GPG key URL or ID
	Type     string `json:"type,omitempty"`      // e.g. "deb", "rpm-md", "pkg"
	Arch     string `json:"arch,omitempty"`      // architecture filter
}

// ApplyOptions processes functional options into an Options struct.
func ApplyOptions(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
