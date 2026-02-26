//go:build containertest

// Package snack_test provides testcontainers-based integration tests that
// exercise every Manager method and capability interface across distros.
//
//	go test -tags containertest -v -count=1 -timeout 15m .
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

// distroTest describes a container environment and the test packages available in it.
type distroTest struct {
	name     string
	image    string
	setup    string // shell commands to install deps
	packages string // space-separated Go test directories

	// Test fixture data — packages known to exist in this distro's repos.
	testPkg        string // a small package to install/remove (e.g. "tree")
	searchQuery    string // a query that returns results (e.g. "curl")
	infoPkg        string // a package to get info on (always available)
	knownFile      string // a file path owned by a known package
	knownFileOwner string // the package that owns knownFile (may include version)
	groupName      string // a package group name (empty if no Grouper)
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
		name:           "debian-apt",
		image:          "debian:bookworm",
		setup:          installGo("apt-get update && apt-get install -y sudo tree curl wget dpkg"),
		packages:       "./apt/ ./dpkg/ ./detect/",
		testPkg:        "tree",
		searchQuery:    "curl",
		infoPkg:        "bash",
		knownFile:      "/usr/bin/tree",
		knownFileOwner: "tree",
	},
	{
		name:           "alpine-apk",
		image:          "alpine:latest",
		setup:          installGo("apk add --no-cache sudo tree bash wget libc6-compat curl"),
		packages:       "./apk/ ./detect/",
		testPkg:        "tree",
		searchQuery:    "curl",
		infoPkg:        "bash",
		knownFile:      "/usr/bin/tree",
		knownFileOwner: "tree",
	},
	{
		name:           "archlinux-pacman",
		image:          "archlinux:latest",
		setup:          installGo("pacman -Syu --noconfirm && pacman -S --noconfirm sudo tree wget"),
		packages:       "./pacman/ ./detect/",
		testPkg:        "tree",
		searchQuery:    "curl",
		infoPkg:        "bash",
		knownFile:      "/usr/bin/tree",
		knownFileOwner: "tree",
		groupName:      "base-devel",
	},
	{
		name:           "fedora-dnf4",
		image:          "fedora:39",
		setup:          installGo("dnf install -y tree sudo wget"),
		packages:       "./dnf/ ./rpm/ ./detect/",
		testPkg:        "tree",
		searchQuery:    "curl",
		infoPkg:        "bash",
		knownFile:      "/usr/bin/tree",
		knownFileOwner: "tree",
		groupName:      "Development Tools",
	},
	{
		name:           "fedora-dnf5",
		image:          "fedora:latest",
		setup:          installGo("dnf install -y tree sudo wget"),
		packages:       "./dnf/ ./rpm/ ./detect/",
		testPkg:        "tree",
		searchQuery:    "curl",
		infoPkg:        "bash",
		knownFile:      "/usr/bin/tree",
		knownFileOwner: "tree",
		groupName:      "Development Tools",
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
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
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

			// Run integration tests with coverage
			testCmd := fmt.Sprintf(
				"cd /src && export PATH=/usr/local/go/bin:$PATH && export GOPATH=/tmp/go && "+
					"go test -v -tags integration -count=1 -coverprofile=/tmp/coverage.out -coverpkg=./... %s 2>&1",
				d.packages,
			)
			t.Logf("Running tests in %s: %s", d.name, testCmd)

			exitCode, reader, err := container.Exec(ctx, []string{"sh", "-c", testCmd})
			if err != nil {
				t.Fatalf("test exec failed: %v", err)
			}

			buf := make([]byte, 128*1024)
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

			// Extract coverage file
			cpOut := exec.CommandContext(ctx, "docker", "cp",
				containerID+":/tmp/coverage.out", fmt.Sprintf("coverage-%s.out", d.name))
			if out, err := cpOut.CombinedOutput(); err != nil {
				t.Logf("coverage extraction failed (non-fatal): %v\n%s", err, out)
			}
		})
	}
}
