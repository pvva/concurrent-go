package concurrent

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestTaskPool(t *testing.T) {
	falseCounter := int32(0)
	allCounter := int32(0)

	taskLimit := 1000000

	pool := NewTaskPool(1000, func(i interface{}) {
		atomic.AddInt32(&falseCounter, 1)
	})

	for i := 0; i < taskLimit; i++ {
		pool.Submit(func() {
			randomTimeMultiplier := rand.Float32() * 10.0

			time.Sleep(time.Duration(randomTimeMultiplier) * time.Millisecond)
			atomic.AddInt32(&allCounter, 1)
		})
	}

	pool.WaitForAll()
	pool.Shutdown()

	if falseCounter > 0 {
		t.Errorf("False counter is %d", falseCounter)
	}
	if allCounter != int32(taskLimit) {
		t.Errorf("Not all tasks were executed")
	}
}

func TestTaskPoolRecursive(t *testing.T) {
	falseCounter := int32(0)
	allCounter := int32(0)

	taskLimit := 100

	pool := NewTaskPool(1, func(i interface{}) {
		atomic.AddInt32(&falseCounter, 1)
	})

	var executeTask func(pool *TaskPool)

	executeTask = func(pool *TaskPool) {
		for i := 0; i < 3; i++ {
			pool.Submit(func() {
				randomTimeMultiplier := rand.Float32() * 100.0
				time.Sleep(time.Duration(randomTimeMultiplier) * time.Millisecond)

				ac := int(atomic.AddInt32(&allCounter, 1))
				if ac <= taskLimit {
					pool.Submit(func() {
						executeTask(pool)
					})
				} else {
					atomic.AddInt32(&allCounter, -1)
				}
			})
		}
	}

	executeTask(pool)

	pool.WaitForAll()
	pool.Shutdown()

	if falseCounter > 0 {
		t.Errorf("False counter is %d", falseCounter)
	}
	if allCounter != int32(taskLimit) {
		t.Errorf("Not all tasks were executed")
	}
}
