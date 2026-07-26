package snack_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gogrlx/snack"
	"github.com/stretchr/testify/assert"
)

func TestLocker_MutualExclusion(t *testing.T) {
	var l snack.Locker
	var counter atomic.Int64
	var wg sync.WaitGroup
	const goroutines = 50

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			l.Lock()
			defer l.Unlock()
			// If the lock works, only one goroutine touches
			// the counter at a time.
			cur := counter.Load()
			counter.Store(cur + 1)
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(goroutines), counter.Load())
}

func TestLocker_LockUnlock(t *testing.T) {
	var l snack.Locker
	var n int
	// Should not deadlock; verifies lock/unlock cycle works twice.
	l.Lock()
	n++
	l.Unlock()
	l.Lock()
	n++
	l.Unlock()
	assert.Equal(t, 2, n)
}
