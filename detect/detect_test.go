package detect

import (
	"errors"
	"testing"

	"github.com/gogrlx/snack"
)

func TestByNameUnknown(t *testing.T) {
	_, err := ByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown manager")
	}
	if !errors.Is(err, snack.ErrManagerNotFound) {
		t.Errorf("expected ErrManagerNotFound, got %v", err)
	}
}

func TestByNameKnown(t *testing.T) {
	// All known manager names should be resolvable by ByName, even if
	// unavailable on this system.
	knownNames := []string{"apt", "dnf", "pacman", "apk", "flatpak", "snap", "aur"}
	for _, name := range knownNames {
		t.Run(name, func(t *testing.T) {
			m, err := ByName(name)
			if err != nil {
				t.Fatalf("ByName(%q) returned error: %v", name, err)
			}
			if m.Name() != name {
				t.Errorf("ByName(%q).Name() = %q", name, m.Name())
			}
		})
	}
}

func TestByNameReturnsCorrectType(t *testing.T) {
	m, err := ByName("apt")
	if err != nil {
		t.Skip("apt not in allManagers on this platform")
	}
	if m.Name() != "apt" {
		t.Errorf("expected Name()=apt, got %q", m.Name())
	}
}

func TestAllReturnsSlice(t *testing.T) {
	managers := All()
	// On Linux with apt installed, we should get at least 1
	// But don't fail if none — could be a weird CI environment
	seen := make(map[string]bool)
	for _, m := range managers {
		name := m.Name()
		if seen[name] {
			t.Errorf("duplicate manager in All(): %s", name)
		}
		seen[name] = true
	}
}

func TestAllManagersAreAvailable(t *testing.T) {
	for _, m := range All() {
		if !m.Available() {
			t.Errorf("All() returned unavailable manager: %s", m.Name())
		}
	}
}

func TestDefaultDoesNotPanic(t *testing.T) {
	Reset()
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

func TestDefaultReturnsAvailableManager(t *testing.T) {
	Reset()
	m, err := Default()
	if err != nil {
		t.Skipf("no manager available on this system: %v", err)
	}
	if !m.Available() {
		t.Error("Default() returned unavailable manager")
	}
	if m.Name() == "" {
		t.Error("Default() returned manager with empty name")
	}
}

func TestResetAllowsRedetection(t *testing.T) {
	_, _ = Default()
	Reset()
	// After reset, calling Default() again should work.
	_, _ = Default()
}

func TestResetConcurrent(t *testing.T) {
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			Reset()
			_, _ = Default()
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestHasBinary(t *testing.T) {
	// sh should exist on any Unix system
	if !HasBinary("sh") {
		t.Error("expected HasBinary(sh) = true")
	}
	if HasBinary("this-binary-does-not-exist-anywhere-12345") {
		t.Error("expected HasBinary(nonexistent) = false")
	}
}

func TestCandidatesNotEmpty(t *testing.T) {
	c := candidates()
	if len(c) == 0 {
		t.Error("candidates() returned empty slice")
	}
}

func TestAllManagersNotEmpty(t *testing.T) {
	a := allManagers()
	if len(a) == 0 {
		t.Error("allManagers() returned empty slice")
	}
	// allManagers should be a superset of candidates
	if len(a) < len(candidates()) {
		t.Error("allManagers() should include at least all candidates")
	}
}
