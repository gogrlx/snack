# AGENTS.md - snack

Idiomatic Go wrappers for system package managers. Part of the [grlx](https://github.com/gogrlx/grlx) ecosystem.

## Quick Reference

```bash
# Build
go build ./...

# Unit tests (run anywhere)
go test -race ./...

# Lint
go vet ./...

# Integration tests (require real package manager)
go test -tags integration -v ./apt/     # On Debian/Ubuntu
go test -tags integration -v ./pacman/  # On Arch Linux
go test -tags integration -v ./apk/     # On Alpine
go test -tags integration -v ./dnf/     # On Fedora

# Container-based integration tests (require Docker)
go test -tags containertest -v -timeout 15m .
```

## Project Structure

```
snack/
├── snack.go           # Core Manager interface + capability interfaces
├── types.go           # Package, Options, Repository types + functional options
├── capabilities.go    # GetCapabilities() runtime capability detection
├── errors.go          # Sentinel errors (ErrNotFound, ErrPermissionDenied, etc.)
├── mutex.go           # Locker type for per-provider mutex
├── target.go          # Target helper functions
├── cmd/snack/         # CLI tool (cobra-based)
├── detect/            # Auto-detection of system package manager
│   ├── detect.go      # Default(), All(), ByName() - shared logic
│   ├── detect_linux.go    # Linux candidate ordering
│   ├── detect_freebsd.go  # FreeBSD candidates
│   └── detect_openbsd.go  # OpenBSD candidates
├── apt/               # Debian/Ubuntu (apt-get, apt-cache, dpkg-query)
├── apk/               # Alpine Linux
├── pacman/            # Arch Linux
├── dnf/               # Fedora/RHEL (supports dnf4 and dnf5)
├── rpm/               # RPM queries
├── dpkg/              # Low-level dpkg operations
├── flatpak/           # Flatpak
├── snap/              # Snapcraft
├── pkg/               # FreeBSD pkg(8)
├── ports/             # OpenBSD ports
└── aur/               # Arch User Repository (stub)
```

## Architecture Patterns

### Interface Hierarchy

All providers implement `snack.Manager` (the base interface). Extended capabilities are optional:

```go
// Base - every provider
snack.Manager           // Install, Remove, Purge, Upgrade, Update, List, Search, Info, IsInstalled, Version

// Optional - use type assertions
snack.VersionQuerier    // LatestVersion, ListUpgrades, UpgradeAvailable, VersionCmp
snack.Holder            // Hold, Unhold, ListHeld (version pinning)
snack.Cleaner           // Autoremove, Clean
snack.FileOwner         // FileList, Owner
snack.RepoManager       // ListRepos, AddRepo, RemoveRepo
snack.KeyManager        // AddKey, RemoveKey, ListKeys
snack.Grouper           // GroupList, GroupInfo, GroupInstall
snack.NameNormalizer    // NormalizeName, ParseArch
```

Check capabilities at runtime:

```go
caps := snack.GetCapabilities(mgr)
if caps.Hold {
    mgr.(snack.Holder).Hold(ctx, []string{"nginx"})
}
```

### Provider File Layout

Each provider follows this pattern:

```
apt/
├── apt.go                    # Public type + methods (delegates to platform-specific)
├── apt_linux.go              # Linux implementation (actual CLI calls)
├── apt_other.go              # Stub for non-Linux (returns ErrUnsupportedPlatform)
├── capabilities_linux.go     # Extended capability implementations
├── capabilities_other.go     # Capability stubs
├── normalize.go              # Name normalization helpers
├── parse.go                  # Output parsing functions
├── apt_test.go               # Unit tests (parsing, no system calls)
└── apt_integration_test.go   # Integration tests (//go:build integration)
```

### Build Tags

- **No tag**: Unit tests, run anywhere
- `integration`: Tests that require the actual package manager to be installed
- `containertest`: Tests using testcontainers (require Docker)

### Per-Provider Mutex

Each provider embeds `snack.Locker` and calls `Lock()`/`Unlock()` around mutating operations:

```go
type Apt struct {
    snack.Locker
}

func (a *Apt) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
    a.Lock()
    defer a.Unlock()
    return install(ctx, pkgs, opts...)
}
```

Different providers can run concurrently (apt + snap), but operations on the same provider serialize.

### Functional Options

Use functional options for operation configuration:

```go
mgr.Install(ctx, snack.Targets("nginx"),
    snack.WithSudo(),
    snack.WithAssumeYes(),
    snack.WithDryRun(),
)
```

Available options: `WithSudo()`, `WithAssumeYes()`, `WithDryRun()`, `WithVerbose()`, `WithRefresh()`, `WithReinstall()`, `WithRoot(path)`, `WithFromRepo(repo)`.

### Target Type

Packages are specified using `snack.Target`:

```go
type Target struct {
    Name     string  // Required
    Version  string  // Optional version constraint
    FromRepo string  // Optional repository constraint
    Source   string  // Local file path or URL
}

// Convenience constructor for name-only targets
targets := snack.Targets("nginx", "redis", "curl")
```

## Code Conventions

### Interface Compliance

Verify interface compliance at compile time:

```go
var (
    _ snack.Manager        = (*Apt)(nil)
    _ snack.VersionQuerier = (*Apt)(nil)
    _ snack.Holder         = (*Apt)(nil)
)
```

### Error Handling

Use sentinel errors from `errors.go`:

```go
snack.ErrNotInstalled       // Package is not installed
snack.ErrNotFound           // Package not found in repos
snack.ErrUnsupportedPlatform // Package manager unavailable
snack.ErrPermissionDenied   // Need sudo
snack.ErrManagerNotFound    // detect couldn't find a manager
snack.ErrDaemonNotRunning   // snapd not running, etc.
```

Wrap with context:

```go
return fmt.Errorf("apt-get install: %w", snack.ErrPermissionDenied)
```

### CLI Output Parsing

Parse functions live in `parse.go` and are thoroughly unit-tested:

```go
// apt/parse.go
func parseList(output string) []snack.Package { ... }
func parseSearch(output string) []snack.Package { ... }
func parseInfo(output string) (*snack.Package, error) { ... }

// apt/apt_test.go
func TestParseList(t *testing.T) {
    input := "bash\t5.2-1\tGNU Bourne Again SHell\n..."
    pkgs := parseList(input)
    // assertions
}
```

### Platform-Specific Stubs

Non-supported platforms return `ErrUnsupportedPlatform`:

```go
//go:build !linux

func install(_ context.Context, _ []snack.Target, _ ...snack.Option) error {
    return snack.ErrUnsupportedPlatform
}
```

## Testing

### Unit Tests

Test parsing logic without system calls:

```go
func TestParseList(t *testing.T) {
    input := "bash\t5.2-1\tGNU Bourne Again SHell\ncoreutils\t9.1-1\tGNU core utilities\n"
    pkgs := parseList(input)
    assert.Len(t, pkgs, 2)
    assert.Equal(t, "bash", pkgs[0].Name)
}
```

### Integration Tests

Use `//go:build integration` tag. Test against the real system:

```go
//go:build integration

func TestIntegration_Apt(t *testing.T) {
    mgr := apt.New()
    if !mgr.Available() {
        t.Skip("apt not available")
    }
    // Install, remove, query packages...
}
```

### Test Assertions

Use testify:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

require.NoError(t, err)           // Fatal on error
assert.True(t, installed)         // Continue on failure
assert.Equal(t, "apt", mgr.Name())
```

## CI/CD

GitHub Actions workflow at `.github/workflows/integration.yml`:

- **Unit tests**: Run on ubuntu-latest
- **Integration tests**: Run in native containers (debian, alpine, archlinux, fedora)
- **Coverage**: Uploaded to Codecov

Jobs test against:
- Debian (apt)
- Ubuntu (apt + snap)
- Fedora 39 (dnf4)
- Fedora latest (dnf5)
- Alpine (apk)
- Arch Linux (pacman)
- Ubuntu + Flatpak

## Adding a New Provider

1. Create directory: `newpkg/`
2. Create files following the pattern:
   - `newpkg.go` - public interface
   - `newpkg_linux.go` - Linux implementation
   - `newpkg_other.go` - stubs
   - `parse.go` - output parsing
   - `newpkg_test.go` - parsing unit tests
   - `newpkg_integration_test.go` - integration tests
3. Add compile-time interface checks
4. Register in `detect/detect_linux.go` (or appropriate platform file)
5. Add CI job in `.github/workflows/integration.yml`

## Common Gotchas

1. **Build tags**: Integration tests won't run without `-tags integration`
2. **Sudo**: Most operations need `snack.WithSudo()` when running as non-root
3. **Platform stubs**: Every function in `*_linux.go` needs a stub in `*_other.go`
4. **Mutex**: Always lock around mutating operations
5. **dnf4 vs dnf5**: The dnf package handles both versions with different parsing logic
6. **Architecture suffixes**: apt uses `pkg:amd64` format, use `NameNormalizer` to strip

## Dependencies

Key dependencies (from go.mod):

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/fang` - Cobra context wrapper
- `github.com/stretchr/testify` - Test assertions
- `github.com/testcontainers/testcontainers-go` - Container-based integration tests
