package concurrent

import (
	"container/list"
	"sync"
)

type BlockingQueue struct {
	queue    *list.List
	lock     sync.Mutex
	condLock sync.Mutex
	cond     *sync.Cond
	drained  bool
}

func NewBlockingQueue() *BlockingQueue {
	bq := &BlockingQueue{
		queue: list.New(),
	}
	bq.cond = sync.NewCond(&bq.condLock)

	return bq
}

// Adds data to queue and awakes one of waiting for data goroutines if there are any, returns <true> if data was put to queue.
func (bq *BlockingQueue) Put(data interface{}) bool {
	bq.lock.Lock()
	if bq.drained {
		bq.lock.Unlock()

		return false
	}
	bq.queue.PushBack(data)
	bq.lock.Unlock()

	bq.cond.Signal()

	return true
}

// Get() will read next value from queue or block until there is data. Also returns "drained" sign.
// Given nonblocking <true> as a parameter, returns immediately with the value on top of queue or nil if it is empty.
func (bq *BlockingQueue) Get(nonblocking ...bool) (interface{}, bool) {
	bq.lock.Lock()
	if bq.drained {
		bq.lock.Unlock()

		return nil, true
	}
	if bq.queue.Len() > 0 {
		e := bq.queue.Front()
		bq.queue.Remove(e)
		bq.lock.Unlock()

		return e.Value, false
	}
	bq.lock.Unlock()

	if len(nonblocking) > 0 && nonblocking[0] {
		return nil, false
	}

	bq.cond.L.Lock()
	bq.cond.Wait()
	data, drained := bq.Get(nonblocking...)
	bq.cond.L.Unlock()

	return data, drained
}

// Clears queue and releases all blocking goroutines causing them to get <nil> as data.
// Drained queue will not receive any data.
func (bq *BlockingQueue) Drain() {
	bq.lock.Lock()
	bq.queue.Init()
	bq.drained = true
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
	bq.lock.Lock()
	l := bq.queue.Len()
	bq.lock.Unlock()

	return l
}
