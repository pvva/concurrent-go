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
