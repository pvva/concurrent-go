package concurrent

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestBlockingQueue(t *testing.T) {
	queue := NewBlockingQueue()
	var cnt int32

	for i := 0; i < 1000; i++ {
		go func() {
			for {
				data, drained := queue.Get()
				if drained {
					break
				}
				if _, ok := data.(int); ok {
					atomic.AddInt32(&cnt, 1)
				}
			}
		}()
	}

	values := int(1e6)

	for i := 0; i < values; i++ {
		queue.Put(i)
	}

	time.Sleep(time.Second * 10)
	if cnt != int32(values) {
		t.Fatal("Not all values passed through queue")
	}
}

func TestLimitedBlockingQueue(t *testing.T) {
	queue := NewLimitedBlockingQueue(100)
	var cnt int32

	for i := 0; i < 1000; i++ {
		go func() {
			for {
				data, drained := queue.Get()
				if drained {
					break
				}
				if _, ok := data.(int); ok {
					atomic.AddInt32(&cnt, 1)
				}
			}
		}()
	}

	values := int(1e6)

	for i := 0; i < values; i++ {
		queue.Put(i)
	}

	time.Sleep(time.Second * 10)
	if cnt != int32(values) {
		t.Fatal("Not all values passed through queue")
	}
}

func TestPutToDrainedQueue(t *testing.T) {
	bq := NewBlockingQueue()
	bq.Drain()
	if bq.Put(1) {
		t.Fatal("Drained blocking queue Put was successful")
	}

	lbq := NewLimitedBlockingQueue(5)
	lbq.Drain()
	if lbq.Put(1) {
		t.Fatal("Drained limited blocking queue Put was successful")
	}
}
