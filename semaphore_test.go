package concurrent

import (
	"sync"
	"testing"
	"time"
)

func TestSequential(t *testing.T) {
	sem := NewSemaphore(5)

	var lat time.Time

	go func() {
		sem.Acquire()
		sem.Acquire()
		sem.Acquire()
		sem.Acquire()

		println(sem.Len())

		sem.Acquire()

		println("should hang now")
		sem.Acquire()
		lat = time.Now()
		println("after release")
	}()

	time.Sleep(time.Second * 3)
	lbr := time.Now()
	sem.Release()

	time.Sleep(time.Second)

	if lat.Before(lbr) {
		t.Error("Failed to wait at acquire: ", lat, lbr)
	}

	println("Released")
	if sem.Len() != 5 {
		t.Error("Failed to acquire after release")
	}
	println(sem.Len())

	sem.Release()
	sem.Release()
	sem.Release()
	sem.Release()
	sem.Release()
	println(sem.Len())
	if sem.Len() != 0 {
		t.Error("Failed to release all locks")
	}

	println("Idle release")
	sem.Release()
	if sem.Len() != 0 {
		t.Error("Failed to execute idle release")
	}
}

func TestParallel(t *testing.T) {
	var starter sync.WaitGroup
	starter.Add(1)

	sm := NewSemaphore(20)

	for i := 0; i < 5; i++ {
		go func(idx int) {
			starter.Wait()

			sm.Acquire(5)
			println("gr ", idx, " acquired 5 locks: ", sm.Len())
		}(i)
	}

	starter.Done()
	time.Sleep(time.Second * 3)
	if sm.Len() != 20 {
		t.Error("Failed to acquire all locks concurrently")
	}
	println("after wait: ", sm.Len())
	sm.Release(5)
	time.Sleep(time.Second)
	if sm.Len() != 20 {
		t.Error("Failed to acquire locks after release")
	}
	println("releasing all")
	sm.Release(20)

	println("done: ", sm.Len())
	if sm.Len() != 0 {
		t.Error("Failed to release all locks")
	}
}

func BenchmarkSemaphoreAcquire(b *testing.B) {
	step := 1
	sm := NewSemaphore(int64(b.N * step))

	for n := 0; n < b.N; n++ {
		sm.Acquire(int64(step))
	}
}

func BenchmarkSemaphoreRelease(b *testing.B) {
	step := 1
	sm := NewSemaphore(int64(b.N * step))
	sm.Acquire(int64(b.N * step))

	for n := 0; n < b.N; n++ {
		sm.Release(int64(step))
	}
}

func BenchmarkMutex(b *testing.B) {
	locks := make([]sync.Mutex, b.N)

	for n := 0; n < b.N; n++ {
		locks[n].Lock()
	}
}
