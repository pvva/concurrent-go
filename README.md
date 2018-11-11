This repo contains few things that proved their usefulness but are absent in standard Go library.

## AtomicBoolean
Lets you atomically set and check for a boolean value.
```go
    notDone := concurrent.AtomicBooleanOf(true)

    ...

    for notDone.Get() {
        ...
    }

    ...

    notDone.Set(false)
```

Most of the time it's more convenient to have atomic boolean, then to execute atomic operations on int and compare the value with zero.


## TaskPool
Allows to create a pool, that is able to serve tasks concurrently via goroutines.
Each task is just a function: `type Task func()`
When creating a pool you should specify it's size. **Pool does not resize automatically.**

```go
    pool := concurrent.NewTaskPool(200)

    ...

    immediate := pool.Submit(func() {
        // do your stuff here
    })
    if !immediate {
        // submit was waiting in queue
    }

    ...

    // next call will wait if pool is full
    pool.SubmitSync(func() {
        // do your stuff here
    })

    ...

    pool.WaitForAll()
    pool.Shutdown()
```

Reasoning for task pool is simple: if you start spawning goroutines for every task you have, you may quickly exhaust available CPU and run into starvation.
In order to avoid this situation, maximum amount of concurrent tasks should be limited.

You can provide error handler function for task pool.
```go
    pool := concurrent.NewTaskPool(20, func(err interface{}) {
    	// handle error
    })
```
If you do not provide error handler, then in case of an error in your function, given to `Submit()`,
pool will propagate `panic` to your code, which might crash if there is no panic handling (see more at
[golang spec](https://golang.org/ref/spec#Handling_panics))


## Semaphore
Classic semaphore. Releasing empty semaphore doesn't block (it actually doesn't do anything).
```
    sem := concurrent.NewSemaphore(5)
    sem.Acquire(5)

    ...

    // this will block until some locks in semaphore are released
    sem.Acquire()

    ...

    sem.Release()

    ...

    sem.Release(5)

    ...

    // release entire semaphore
    sem.Release(sem.Cap())
```


## Blocking Queue
Reasonably unlimited queue which will block at attempt to get the element from it if it is empty.
```
    queue := concurrent.NewBlockingQueue()

    // if queue is drained, then success is false and no data is added to queue
    success := queue.Put(5)

    ...

    // possibly blocking get
    // If queue is drained, the call will return immediately and `drained` sign will be true, otherwise the call will block until there is data in queue
    data, drained := queue.Get()
    if data != nil {
        // do something with data
    }

    ...

    // non blocking get (polling)
    data, drained := queue.Get(true)
    if data != nil {
        // do something with data
    }

    ...

    // drain the queue, reseting all goroutines, waiting in Get() call
    queue.Drain()

    // reactivate drained queue
    queue.Reset()
```

## Limited Counter
Simple counter, that resets itself when reaching the limit. Can be safely used by concurrent routines.
```
    lc := concurrent.NewLimitedCounter(20)
    ...
    nextValue := lc.Next()
```
