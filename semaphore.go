package concurrent

import (
	"sync"
)

type Semaphore struct {
	capacity int64
	lock     sync.Mutex
	pool     chan struct{}
}

func NewSemaphore(size int64) *Semaphore {
	return &Semaphore{
		capacity: size,
		pool:     make(chan struct{}, size),
	}
}

func (sm *Semaphore) Acquire(n ...int64) {
	s := int64(1)
	if len(n) > 0 {
		s = n[0]
	}

	sm.lock.Lock()
	if s == 1 {
		if sm.Len() >= sm.Cap() {
			sm.lock.Unlock()
			sm.pool <- struct{}{}
		} else {
			sm.pool <- struct{}{}
			sm.lock.Unlock()
		}
	} else {
		f := sm.Cap() - sm.Len() - s
		if f >= 0 {
			f = s
			s = 0
		} else {
			f += s
			if f > 0 {
				s -= f
			}
		}
		for ; f > 0; f-- {
			sm.pool <- struct{}{}
		}
		sm.lock.Unlock()
		for ; s > 0; s-- {
			sm.pool <- struct{}{}
		}
	}
}

func (sm *Semaphore) Release(n ...int64) {
	s := int64(1)
	if len(n) > 0 {
		s = n[0]
	}

	sm.lock.Lock()
	if s == 1 {
		if sm.Len() >= 1 {
			<-sm.pool
		}
	} else {
		f := sm.Len()
		if f > s {
			f = s
		}
		for ; f > 0; f-- {
			<-sm.pool
		}
	}
	sm.lock.Unlock()
}

func (sm *Semaphore) Len() int64 {
	return int64(len(sm.pool))
}

func (sm *Semaphore) Cap() int64 {
	return sm.capacity
}
