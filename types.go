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

// ApplyOptions processes functional options into an Options struct.
func ApplyOptions(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
