package concurrent

import "sync/atomic"

type AtomicBoolean int32

func AtomicBooleanOf(initial bool) *AtomicBoolean {
	ab := new(AtomicBoolean)
	ab.Set(initial)

	return ab
}

func (ab *AtomicBoolean) Set(value bool) {
	v := int32(0)
	if value {
		v = 1
	}
	atomic.StoreInt32((*int32)(ab), v)
}

func (ab *AtomicBoolean) Get() bool {
	return atomic.LoadInt32((*int32)(ab)) == 1
}

func (ab *AtomicBoolean) CompareAndSet(expected, update bool) bool {
	vExpected := int32(0)
	if expected {
		vExpected = 1
	}
	vUpdate := int32(0)
	if update {
		vUpdate = 1
	}

	return atomic.CompareAndSwapInt32((*int32)(ab), vExpected, vUpdate)
}
