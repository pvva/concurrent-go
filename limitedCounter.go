package concurrent

import (
	"sync/atomic"
)

type LimitedCounter struct {
	counter int32
	limit   int32
}

func NewLimitedCounter(limit int) *LimitedCounter {
	return &LimitedCounter{
		counter: -1,
		limit:   int32(limit),
	}
}

func (lc *LimitedCounter) Next() int {
	newCounter := atomic.AddInt32(&lc.counter, 1)

	if newCounter >= lc.limit {
		shifted := newCounter % lc.limit
		atomic.CompareAndSwapInt32(&lc.counter, newCounter, shifted)

		return int(shifted)
	}

	return int(newCounter)
}
