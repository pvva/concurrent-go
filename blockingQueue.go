package concurrent

import (
	"sync"
	"sync/atomic"
)

type BlockingQueue struct {
	queue   []interface{}
	len     int32
	lock    sync.Mutex
	cond    *sync.Cond
	drained bool
}

func NewBlockingQueue() *BlockingQueue {
	bq := &BlockingQueue{
		queue: []interface{}{},
	}
	bq.cond = sync.NewCond(&bq.lock)

	return bq
}

// Adds data to queue and awakes one of waiting for data goroutines if there are any, returns <true> if data was put to queue.
func (bq *BlockingQueue) Put(data interface{}) bool {
	bq.lock.Lock()
	if bq.drained {
		bq.lock.Unlock()

		return false
	}
	bq.queue = append(bq.queue, data)
	atomic.AddInt32(&bq.len, 1)

	bq.cond.Signal()
	bq.lock.Unlock()

	return true
}

func (bq *BlockingQueue) internalGetWithUnlock() (interface{}, bool) {
	if bq.drained {
		bq.lock.Unlock()

		return nil, true
	}
	if len(bq.queue) > 0 {
		e := bq.queue[0]
		bq.queue = bq.queue[1:]
		atomic.AddInt32(&bq.len, -1)
		bq.lock.Unlock()

		return e, false
	}
	bq.lock.Unlock()

	return nil, false
}

// Get() will read next value from queue or block until there is data. Also returns "drained" sign.
// Given nonblocking <true> as a parameter, returns immediately with the value on top of queue or nil if it is empty.
func (bq *BlockingQueue) Get(nonblocking ...bool) (interface{}, bool) {
	bq.lock.Lock()

	for {
		data, drained := bq.internalGetWithUnlock()
		if drained {
			return nil, true
		}
		if data != nil {
			return data, false
		} else if len(nonblocking) > 0 && nonblocking[0] {
			return nil, false
		}

		bq.lock.Lock()
		bq.cond.Wait()
	}
}

// Clears queue and releases all blocking goroutines causing them to get <nil> as data.
// Drained queue will not receive any data.
func (bq *BlockingQueue) Drain() {
	bq.lock.Lock()
	bq.drained = true
	bq.queue = bq.queue[:0]
	bq.lock.Unlock()

	bq.cond.Broadcast()
}

// Activates queue to start receiving data again.
func (bq *BlockingQueue) Reset() {
	bq.lock.Lock()
	bq.drained = false
	bq.lock.Unlock()
}

// Queue size.
func (bq *BlockingQueue) Len() int {
	return int(atomic.LoadInt32(&bq.len))
}
