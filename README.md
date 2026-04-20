# snack 🍿

Idiomatic Go wrappers for system package managers.

[![License 0BSD](https://img.shields.io/badge/License-0BSD-pink.svg)](https://opensource.org/licenses/0BSD)
[![GoDoc](https://img.shields.io/badge/GoDoc-reference-007d9c)](https://pkg.go.dev/github.com/gogrlx/snack)

**snack** provides thin, context-aware Go bindings for system package managers. Think [`taigrr/systemctl`](https://github.com/taigrr/systemctl) but for package management.

Part of the [grlx](https://github.com/gogrlx/grlx) ecosystem.

## Supported Package Managers

| Package | Manager | Platform | Status |
|---------|---------|----------|--------|
| `pacman` | pacman | Arch Linux | ✅ |
| `aur` | AUR (RPC + makepkg) | Arch Linux | ✅ |
| `apk` | apk-tools | Alpine Linux | ✅ |
| `apt` | APT (apt-get/apt-cache) | Debian/Ubuntu | ✅ |
| `dpkg` | dpkg | Debian/Ubuntu | ✅ |
| `dnf` | DNF | Fedora/RHEL | ✅ |
| `rpm` | RPM | Fedora/RHEL | ✅ |
| `flatpak` | Flatpak | Linux | ✅ |
| `snap` | snapd | Linux | ✅ |
| `brew` | Homebrew | macOS/Linux | ✅ |
| `pkg` | pkg(8) | FreeBSD | ✅ |
| `ports` | ports/packages | OpenBSD | ✅ |
| `winget` | Windows Package Manager | Windows | ✅ |
| `detect` | Auto-detection | All | ✅ |

### Capability Matrix

| Provider | VersionQuery | Hold | Clean | FileOwner | RepoMgmt | KeyMgmt | Groups | NameNorm | DryRun | PkgUpgrade |
|----------|:------------:|:----:|:-----:|:---------:|:--------:|:-------:|:------:|:--------:|:------:|:----------:|
| apt      | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - | ✅ | ✅ | ✅ |
| pacman   | ✅ | - | ✅ | ✅ | - | - | ✅ | ✅ | ✅ | ✅ |
| aur      | ✅ | - | ✅ | - | - | - | - | - | - | ✅ |
| apk      | ✅ | - | ✅ | ✅ | - | - | - | ✅ | ✅ | ✅ |
| dnf      | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| flatpak  | ✅ | - | ✅ | - | ✅ | - | - | ✅ | - | ✅ |
| snap     | ✅ | - | ✅ | - | - | - | - | ✅ | - | ✅ |
| brew     | ✅ | - | ✅ | ✅ | - | - | - | ✅ | - | ✅ |
| pkg      | ✅ | - | ✅ | ✅ | - | - | - | ✅ | - | ✅ |
| ports    | ✅ | - | ✅ | ✅ | - | - | - | ✅ | - | ✅ |
| winget   | ✅ | - | - | - | ✅ | - | - | ✅ | - | ✅ |

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

    // Install a package
    _, err := mgr.Install(ctx, snack.Targets("nginx"), snack.WithSudo(), snack.WithAssumeYes())
    if err != nil {
        log.Fatal(err)
    }

    // Check if installed
    installed, err := mgr.IsInstalled(ctx, "nginx")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("nginx installed:", installed)
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

aptMgr, err := detect.ByName(" APT ")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Requested:", aptMgr.Name())
```

## Interfaces

snack uses a layered interface design. Every provider implements `Manager` (the base). Extended capabilities are optional — use type assertions to check support:

```go
// Base — every provider
snack.Manager           // Install, Remove, Purge, Upgrade, Update, List, Search, Info, IsInstalled, Version

// Optional capabilities — type-assert to check
snack.VersionQuerier    // LatestVersion, ListUpgrades, UpgradeAvailable, VersionCmp
snack.Holder            // Hold, Unhold, ListHeld (version pinning)
snack.Cleaner           // Autoremove, Clean (orphan/cache cleanup)
snack.FileOwner         // FileList, Owner (file-to-package queries)
snack.RepoManager       // ListRepos, AddRepo, RemoveRepo
snack.KeyManager        // AddKey, RemoveKey, ListKeys (GPG keys)
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

## Design

- **Thin CLI wrappers** — each sub-package wraps a package manager's CLI tools. No FFI, no library bindings.
- **Common interface** — all managers implement `snack.Manager`, making them interchangeable.
- **Capability interfaces** — extended features via type assertion, so providers aren't forced to stub unsupported operations.
- **Per-provider mutex** — each provider serializes mutating operations independently; apt + snap can run in parallel.
- **Context-aware** — all operations accept `context.Context` for cancellation and timeouts.
- **Platform-safe** — build tags ensure packages compile everywhere but only run where appropriate.
- **No root assumption** — use `snack.WithSudo()` when elevated privileges are needed.

## Implementation Priority

1. pacman + AUR (Arch Linux)
2. apk (Alpine Linux)
3. apt + dpkg (Debian/Ubuntu)
4. dnf + rpm (Fedora/RHEL)
5. flatpak + snap (cross-distro)
6. pkg + ports (BSD)
7. winget (Windows)

## CLI

A companion CLI tool is planned for direct terminal usage:

```bash
snack install nginx
snack remove nginx
snack search redis
snack list
snack upgrade
```

## License

0BSD — see [LICENSE](LICENSE).
