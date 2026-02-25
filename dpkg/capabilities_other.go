//go:build !linux

package dpkg

import (
	"context"

	"github.com/gogrlx/snack"
)

func fileList(_ context.Context, _ string) ([]string, error) {
	return nil, snack.ErrUnsupportedPlatform
}

func owner(_ context.Context, _ string) (string, error) {
	return "", snack.ErrUnsupportedPlatform
}

// normalizeName and parseArch are in normalize.go (no build constraints).
