# snack 🍿

Idiomatic Go wrappers for system package managers.

[![License 0BSD](https://img.shields.io/badge/License-0BSD-pink.svg)](https://opensource.org/licenses/0BSD)
[![GoDoc](https://img.shields.io/badge/GoDoc-reference-007d9c)](https://pkg.go.dev/github.com/gogrlx/snack)

**snack** provides thin, context-aware Go bindings for system package managers. Think [`taigrr/systemctl`](https://github.com/taigrr/systemctl) but for package management.

Part of the [grlx](https://github.com/gogrlx/grlx) ecosystem.

## Supported Package Managers

| Package | Manager | Platform | Status |
|---------|---------|----------|--------|
| `pacman` | pacman | Arch Linux | 🚧 |
| `aur` | AUR (makepkg) | Arch Linux | 🚧 |
| `apk` | apk-tools | Alpine Linux | 🚧 |
| `apt` | APT (apt-get/apt-cache) | Debian/Ubuntu | 🚧 |
| `dpkg` | dpkg | Debian/Ubuntu | 🚧 |
| `dnf` | DNF | Fedora/RHEL | 🚧 |
| `rpm` | RPM | Fedora/RHEL | 🚧 |
| `flatpak` | Flatpak | Cross-distro | 🚧 |
| `snap` | snapd | Cross-distro | 🚧 |
| `pkg` | pkg(8) | FreeBSD | 🚧 |
| `ports` | ports/packages | OpenBSD | 🚧 |
| `detect` | Auto-detection | All | 🚧 |

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
    err := mgr.Install(ctx, []string{"nginx"}, snack.WithSudo(), snack.WithAssumeYes())
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
```

## Design

- **Thin CLI wrappers** — each sub-package wraps a package manager's CLI tools. No FFI, no library bindings.
- **Common interface** — all managers implement `snack.Manager`, making them interchangeable.
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
