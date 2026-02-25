//go:build linux

package dpkg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogrlx/snack"
)

func fileList(ctx context.Context, pkg string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "dpkg", "-L", pkg)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		errMsg := stderr.String()
		if strings.Contains(errMsg, "is not installed") || strings.Contains(errMsg, "not found") {
			return nil, fmt.Errorf("dpkg -L %s: %w", pkg, snack.ErrNotInstalled)
		}
		return nil, fmt.Errorf("dpkg -L %s: %w: %s", pkg, err, errMsg)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func owner(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "dpkg", "-S", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("dpkg -S %s: %w", path, snack.ErrNotFound)
	}
	line := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		return "", fmt.Errorf("dpkg -S %s: unexpected output", path)
	}
	pkgPart := line[:colonIdx]
	if commaIdx := strings.Index(pkgPart, ","); commaIdx >= 0 {
		pkgPart = strings.TrimSpace(pkgPart[:commaIdx])
	}
	return strings.TrimSpace(pkgPart), nil
}
