package concurrent

import (
	"runtime"
	"sync"
	"testing"
)

func TestLimitedCounter(t *testing.T) {
	limit := 20
	innerLimit := 100
	results := make([]int, limit)
	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(limit)
	lc := NewLimitedCounter(limit)

	for i := 0; i < int(limit); i++ {
		go func() {
			for j := 0; j < innerLimit; j++ {
				idx := lc.Next()
				lock.Lock()
				results[idx]++
				lock.Unlock()
				runtime.Gosched()
			}

			wg.Done()
		}()
	}

	wg.Wait()

	for i := 0; i < limit; i++ {
		if results[i] != innerLimit {
			t.Fatal("Limited counter failed to increment concurrently")
		}
	}
}

func BenchmarkLimitedCounter(b *testing.B) {
	lc := NewLimitedCounter(10)

	for n := 0; n < b.N; n++ {
		lc.Next()
	}
}
