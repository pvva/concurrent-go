package concurrent

import (
	"sync"
	"sync/atomic"
)

type Task func()

type TaskPool struct {
	queue        chan Task
	queueSize    int32
	workersSize  int32
	accept       *AtomicBoolean
	done         chan struct{}
	finishWaiter sync.WaitGroup
	errorHandler func()
}

func NewTaskPool(workers int, errorHandlers ...func(interface{})) *TaskPool {
	pool := &TaskPool{
		queue:       make(chan Task, workers),
		accept:      AtomicBooleanOf(true),
		done:        make(chan struct{}),
		workersSize: int32(workers),
	}
	if len(errorHandlers) > 0 && errorHandlers[0] != nil {
		pool.errorHandler = func() {
			if err := recover(); err != nil {
				errorHandlers[0](err)
			}
		}
	} else {
		pool.errorHandler = func() {}
	}

	for i := 0; i < workers; i++ {
		go pool.doWork()
	}

	return pool
}

func (pool *TaskPool) queueTask(task Task) {
	if pool.accept.Get() {
		defer pool.errorHandler()

		atomic.AddInt32(&pool.queueSize, 1)
		pool.queue <- task
	}
}

// submits a task waiting if necessary
func (pool *TaskPool) SubmitSync(task Task) {
	if pool.accept.Get() {
		pool.registerWork()
		pool.queueTask(task)
	}
}

// returns true if submit was not waiting
func (pool *TaskPool) Submit(task Task) bool {
	immediate := false

	if pool.accept.Get() {
		pool.registerWork()

		immediate = atomic.LoadInt32(&pool.queueSize) < pool.workersSize
		if immediate {
			pool.queueTask(task)
		} else {
			go pool.queueTask(task)
		}
	}

	return immediate
}

func (pool *TaskPool) Shutdown() {
	pool.accept.Set(false)
	pool.done <- struct{}{}
	close(pool.queue)
	close(pool.done)
	atomic.StoreInt32(&pool.queueSize, 0)
}

func (pool *TaskPool) doWork() {
	for {
		select {
		case task := <-pool.queue:
			atomic.AddInt32(&pool.queueSize, -1)

			if task == nil {
				continue
			}

			pool.doTask(task)
		case <-pool.done:
			return
		}
	}
}

func (pool *TaskPool) doTask(task Task) {
	defer func() {
		pool.errorHandler()
		pool.doneWork()
	}()

	task()
}

func (pool *TaskPool) registerWork() {
	pool.finishWaiter.Add(1)
}

func (pool *TaskPool) doneWork() {
	pool.finishWaiter.Done()
}

func (pool *TaskPool) WaitForAll() {
	defer pool.errorHandler()
	pool.finishWaiter.Wait()
}
