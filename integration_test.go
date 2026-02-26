//go:build containertest

// Package snack_test provides testcontainers-based integration tests that
// run each package manager in its native container. Use:
//
//	go test -tags containertest -v -count=1 -timeout 10m .
//
// Requires Docker.
package snack_test

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const goVersion = "1.26.0"

type distroTest struct {
	name     string
	image    string
	setup    string // shell commands to install deps (Go installed separately)
	packages string // space-separated test directories
}

// installGo returns a shell snippet that installs Go from the official tarball.
func installGo(prereqs string) string {
	return fmt.Sprintf(
		"%s && wget -qO- https://go.dev/dl/go%s.linux-amd64.tar.gz | tar -C /usr/local -xzf -",
		prereqs, goVersion,
	)
}

var distros = []distroTest{
	{
		name:     "debian-apt",
		image:    "debian:bookworm",
		setup:    installGo("apt-get update && apt-get install -y sudo tree curl wget"),
		packages: "./apt/ ./dpkg/ ./detect/",
	},
	{
		name:     "alpine-apk",
		image:    "alpine:latest",
		setup:    installGo("apk add --no-cache sudo tree bash wget libc6-compat"),
		packages: "./apk/ ./detect/",
	},
	{
		name:     "archlinux-pacman",
		image:    "archlinux:latest",
		setup:    installGo("pacman -Syu --noconfirm && pacman -S --noconfirm sudo tree wget"),
		packages: "./pacman/ ./detect/",
	},
	{
		name:     "fedora-dnf4",
		image:    "fedora:39",
		setup:    installGo("dnf install -y tree sudo wget"),
		packages: "./dnf/ ./rpm/ ./detect/",
	},
	{
		name:     "fedora-dnf5",
		image:    "fedora:latest",
		setup:    installGo("dnf install -y tree sudo wget"),
		packages: "./dnf/ ./rpm/ ./detect/",
	},
}

func TestContainers(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}

	for _, d := range distros {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			req := testcontainers.ContainerRequest{
				Image: d.image,
				Cmd:   []string{"sleep", "infinity"},
				WaitingFor: wait.ForExec([]string{"echo", "ready"}).
					WithStartupTimeout(60 * time.Second),
			}

			container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			if err != nil {
				t.Fatalf("failed to start %s container: %v", d.name, err)
			}
			defer func() {
				_ = container.Terminate(ctx)
			}()

			// Copy source into container
			containerID := container.GetContainerID()

			// Execute setup
			t.Logf("Setting up %s...", d.name)
			exitCode, output, err := container.Exec(ctx, []string{"sh", "-c", d.setup})
			if err != nil || exitCode != 0 {
				t.Fatalf("setup failed (exit %d): %v\n%s", exitCode, err, output)
			}

			// Copy the module into the container
			copyCmd := exec.CommandContext(ctx, "docker", "cp", ".", containerID+":/src")
			if out, err := copyCmd.CombinedOutput(); err != nil {
				t.Fatalf("docker cp failed: %v\n%s", err, out)
			}

			// Run integration tests
			testCmd := fmt.Sprintf(
				"cd /src && export PATH=/usr/local/go/bin:$PATH && export GOPATH=/tmp/go && go test -v -tags integration -count=1 %s",
				d.packages,
			)
			t.Logf("Running tests in %s: %s", d.name, testCmd)

			exitCode, reader, err := container.Exec(ctx, []string{"sh", "-c", testCmd})
			if err != nil {
				t.Fatalf("test exec failed: %v", err)
			}

			// Read output
			buf := make([]byte, 64*1024)
			var output2 strings.Builder
			for {
				n, readErr := reader.Read(buf)
				if n > 0 {
					output2.Write(buf[:n])
				}
				if readErr != nil {
					break
				}
			}
			t.Log(output2.String())

			if exitCode != 0 {
				t.Fatalf("%s integration tests failed (exit %d)", d.name, exitCode)
			}
		})
	}
}
