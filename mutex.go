package snack

import "sync"

// Locker is an embeddable type that provides per-provider mutex locking.
// Each package manager backend should embed this in its struct to serialize
// mutating operations (Install, Remove, Purge, Upgrade, Update).
//
// Read-only operations (List, Search, Info, IsInstalled, Version) generally
// don't need the lock, but backends may choose to lock them if the underlying
// tool doesn't support concurrent reads.
//
// Different providers use independent locks, so an apt Install and a snap
// Install can run concurrently, but two apt Installs will serialize.
type Locker struct {
	mu sync.Mutex
}

// Lock acquires the provider lock. Call before mutating operations.
func (l *Locker) Lock() {
	l.mu.Lock()
}

// Unlock releases the provider lock. Defer after Lock.
func (l *Locker) Unlock() {
	l.mu.Unlock()
}
