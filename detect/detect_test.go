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

func TestDefaultCachesResult(t *testing.T) {
	Reset()
	m1, err1 := Default()
	m2, err2 := Default()
	if m1 != m2 {
		t.Error("expected Default() to return the same Manager on repeated calls")
	}
	if err1 != err2 {
		t.Error("expected Default() to return the same error on repeated calls")
	}
	// Exactly one of m1/err1 must be non-nil (either a manager was found or not).
	if (m1 == nil) == (err1 == nil) {
		t.Error("expected exactly one of Manager or error to be non-nil")
	}
}

func TestResetAllowsRedetection(t *testing.T) {
	_, _ = Default()
	Reset()
	// After reset, defaultOnce should be fresh; calling Default() again should work.
	_, _ = Default()
}
