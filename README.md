# snack 🍿

Idiomatic Go wrappers for system package managers.

[![License 0BSD](https://img.shields.io/badge/License-0BSD-pink.svg)](https://opensource.org/licenses/0BSD)
[![GoDoc](https://img.shields.io/badge/GoDoc-reference-007d9c)](https://pkg.go.dev/github.com/gogrlx/snack)

**snack** provides thin, context-aware Go bindings for system package managers. Think [`taigrr/systemctl`](https://github.com/taigrr/systemctl) but for package management.

Part of the [grlx](https://github.com/gogrlx/grlx) ecosystem.

## Supported Package Managers

| Package | Manager | Platform | Extras |
|---------|---------|----------|--------|
| `apt` | APT (apt-get/apt-cache) | Debian/Ubuntu | DryRun, FileOwner, Holder, KeyManager, NameNormalizer, RepoManager |
| `dpkg` | dpkg | Debian/Ubuntu | DryRun, FileOwner, NameNormalizer |
| `dnf` | DNF 4/5 | Fedora/RHEL | DryRun, FileOwner, Grouper, Holder, KeyManager, NameNormalizer, RepoManager |
| `rpm` | RPM | Fedora/RHEL | FileOwner, NameNormalizer |
| `pacman` | pacman | Arch Linux | DryRun, FileOwner, Grouper |
| `aur` | AUR (makepkg) | Arch Linux | — |
| `apk` | apk-tools | Alpine Linux | DryRun, FileOwner |
| `flatpak` | Flatpak | Cross-distro | RepoManager |
| `snap` | snapd | Cross-distro | — |
| `pkg` | pkg(8) | FreeBSD | FileOwner |
| `ports` | ports/packages | OpenBSD | FileOwner |
| `detect` | Auto-detection | All | — |

All providers implement `Manager`, `VersionQuerier`, `Cleaner`, and `PackageUpgrader`. The **Extras** column lists additional capabilities beyond that baseline.

## Install

```bash
go get github.com/gogrlx/snack
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gogrlx/snack"
    "github.com/gogrlx/snack/apt"
)

func main() {
    ctx := context.Background()
    mgr := apt.New()

    // Install packages
    result, err := mgr.Install(ctx, snack.Targets("nginx", "curl"), snack.WithSudo(), snack.WithAssumeYes())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Installed: %d, Unchanged: %d\n", len(result.Installed), len(result.Unchanged))

    // Check if installed
    installed, err := mgr.IsInstalled(ctx, "nginx")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("nginx installed:", installed)

    // Upgrade specific packages
    if up, ok := mgr.(snack.PackageUpgrader); ok {
        _, err := up.UpgradePackages(ctx, snack.Targets("nginx"), snack.WithSudo(), snack.WithAssumeYes())
        if err != nil {
            log.Fatal(err)
        }
    }
}
```

### Auto-detection

```go
import "github.com/gogrlx/snack/detect"

mgr, err := detect.Default()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Detected:", mgr.Name())

// All available managers
for _, m := range detect.All() {
    fmt.Println(m.Name())
}
```

## Interfaces

snack uses a layered interface design. Every provider implements `Manager` (the base). Extended capabilities are optional — use type assertions to check support:

```go
// Base — every provider
snack.Manager            // Install, Remove, Purge, Upgrade, Update, List, Search, Info, IsInstalled, Version

// Core optional — implemented by all providers
snack.VersionQuerier     // LatestVersion, ListUpgrades, UpgradeAvailable, VersionCmp
snack.Cleaner            // Autoremove, Clean (orphan/cache cleanup)
snack.PackageUpgrader    // UpgradePackages (upgrade specific packages)

// Provider-specific — type-assert to check
snack.Holder             // Hold, Unhold, ListHeld, IsHeld (version pinning)
snack.FileOwner          // FileList, Owner (file-to-package queries)
snack.RepoManager        // ListRepos, AddRepo, RemoveRepo
snack.KeyManager         // AddKey, RemoveKey, ListKeys (GPG keys)
snack.Grouper            // GroupList, GroupInfo, GroupInstall, GroupIsInstalled
snack.NameNormalizer     // NormalizeName, ParseArch
snack.DryRunner          // SupportsDryRun (honors WithDryRun option)
```

Check capabilities at runtime:

```go
caps := snack.GetCapabilities(mgr)
if caps.Hold {
    mgr.(snack.Holder).Hold(ctx, []string{"nginx"})
}
if caps.FileOwnership {
    owner, _ := mgr.(snack.FileOwner).Owner(ctx, "/usr/bin/curl")
    fmt.Println("Owned by:", owner)
}
```

## Options

All mutating operations accept functional options:

```go
snack.WithSudo()           // prepend sudo
snack.WithAssumeYes()      // auto-confirm prompts
snack.WithDryRun()         // simulate (if DryRunner)
snack.WithVerbose()        // verbose output
snack.WithRefresh()        // refresh index before operation
snack.WithReinstall()      // reinstall even if current
snack.WithRoot("/mnt")     // alternate root filesystem
snack.WithFromRepo("sid")  // install from specific repository
```

## CLI

A companion CLI is included at `cmd/snack`:

```bash
snack install nginx curl     # install packages
snack remove nginx           # remove packages
snack upgrade                # upgrade all packages
snack update                 # refresh package index
snack search redis           # search for packages
snack info nginx             # show package details
snack list                   # list installed packages
snack which /usr/bin/curl    # find owning package
snack hold nginx             # pin package version
snack unhold nginx           # unpin package version
snack clean                  # autoremove + clean cache
snack detect                 # show detected managers + capabilities
snack version                # show version
```

Global flags: `--manager <name>`, `--sudo`, `--yes`, `--dry-run`

## Design

- **Thin CLI wrappers** — each sub-package wraps a package manager's CLI tools. No FFI, no library bindings.
- **Common interface** — all managers implement `snack.Manager`, making them interchangeable.
- **Capability interfaces** — extended features via type assertion, so providers aren't forced to stub unsupported operations.
- **Per-provider mutex** — each provider serializes mutating operations independently; apt + snap can run in parallel.
- **Context-aware** — all operations accept `context.Context` for cancellation and timeouts.
- **Platform-safe** — build tags ensure packages compile everywhere but only run where appropriate.
- **No root assumption** — use `snack.WithSudo()` when elevated privileges are needed.

## License

0BSD — see [LICENSE](LICENSE).
