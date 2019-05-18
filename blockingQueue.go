package concurrent

import (
	"sync"
	"sync/atomic"
)

type Queue interface {
	Put(data interface{}) bool
	Get(nonblocking ...bool) (interface{}, bool)
	Drain()
	Reset()
	Len() int
}

type blockingQueue struct {
	queue   []interface{}
	len     int32
	lock    sync.Mutex
	cond    *sync.Cond
	drained bool
}

func NewBlockingQueue() Queue {
	bq := &blockingQueue{
		queue: []interface{}{},
	}
	bq.cond = sync.NewCond(&bq.lock)

	return bq
}

// Adds data to queue and awakes one of waiting for data goroutines if there are any, returns <true> if data was put to
// queue.
func (bq *blockingQueue) Put(data interface{}) bool {
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

func (bq *blockingQueue) internalGetWithUnlock() (interface{}, bool) {
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
func (bq *blockingQueue) Get(nonblocking ...bool) (interface{}, bool) {
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
func (bq *blockingQueue) Drain() {
	bq.lock.Lock()
	bq.drained = true
	bq.queue = bq.queue[:0]
	bq.lock.Unlock()

	bq.cond.Broadcast()
}

// Activates queue to start receiving data again.
func (bq *blockingQueue) Reset() {
	bq.lock.Lock()
	bq.drained = false
	bq.lock.Unlock()
}

// Queue size.
func (bq *blockingQueue) Len() int {
	return int(atomic.LoadInt32(&bq.len))
}

type limitedBlockingQueue struct {
	queue   chan interface{}
	limit   int
	len     int32
	drained bool
	lock    sync.Mutex
}

func NewLimitedBlockingQueue(limit int) Queue {
	return &limitedBlockingQueue{
		queue:   make(chan interface{}, limit),
		drained: false,
		limit:   limit,
	}
}

// Adds data to queue and awakes one of waiting for data goroutines if there are any, returns <true> if data was put to
// queue. Blocks if queue is full.
func (lbq *limitedBlockingQueue) Put(data interface{}) bool {
	ret := true
	defer func() {
		// recover in case of send to closed channel
		if recover() != nil {
			ret = false
		}
	}()
	lbq.queue <- data
	atomic.AddInt32(&lbq.len, 1)

	return ret
}

func (lbq *limitedBlockingQueue) Get(nonblocking ...bool) (interface{}, bool) {
	lbq.lock.Lock()
	defer lbq.lock.Unlock()

	if lbq.drained {
		return nil, true
	}

	if len(nonblocking) > 0 && nonblocking[0] {
		select {
		case data := <-lbq.queue:
			atomic.AddInt32(&lbq.len, -1)
			return data, false
		default:
			return nil, false
		}
	}
	data := <-lbq.queue
	atomic.AddInt32(&lbq.len, -1)

	return data, false
}

func (lbq *limitedBlockingQueue) Drain() {
	close(lbq.queue)
	lbq.lock.Lock()
	lbq.drained = true
	atomic.StoreInt32(&lbq.len, 0)
	lbq.lock.Unlock()
}

func (lbq *limitedBlockingQueue) Reset() {
	lbq.lock.Lock()
	lbq.queue = make(chan interface{}, lbq.limit)
	lbq.drained = false
	atomic.StoreInt32(&lbq.len, 0)
	lbq.lock.Unlock()
}

func (lbq *limitedBlockingQueue) Len() int {
	return int(atomic.LoadInt32(&lbq.len))
}
