package detect

import (
	"testing"
)

func TestByNameUnknown(t *testing.T) {
	_, err := ByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown manager")
	}
}

func TestAllReturnsSlice(t *testing.T) {
	// Just verify it doesn't panic; actual availability depends on system.
	_ = All()
}

func TestDefaultDoesNotPanic(t *testing.T) {
	// May return error if no managers available; that's fine.
	_, _ = Default()
}
