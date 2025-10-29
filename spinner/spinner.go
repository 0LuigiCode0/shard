package spinner

import (
	"runtime"
	"sync/atomic"
	_ "unsafe"
)

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type Spinner struct {
	_ noCopy

	locked  uint32
	rlocked int64

	LoopCount byte
	SpinCount byte
}

func spin(cycle byte)

func NewSpinner(loopCount, spinCount byte) Spinner {
	return Spinner{LoopCount: loopCount, SpinCount: spinCount}
}

// -------------------------------------------------------------------------- //
// MARK:Lock/Unlock
// -------------------------------------------------------------------------- //

func (s *Spinner) Lock() {
	for {
		for range s.LoopCount {
			if s.locked == 0 && atomic.CompareAndSwapUint32(&s.locked, 0, 1) {
				for s.rlocked != 0 {
					spin(s.SpinCount)
				}
				return
			}
			spin(s.SpinCount)
		}
		runtime.Gosched()
	}
}

func (s *Spinner) Unlock() {
	atomic.StoreUint32(&s.locked, 0)
}

func (s *Spinner) RLock() {
	for {
		for range s.LoopCount {
			if s.locked == 0 {
				atomic.AddInt64(&s.rlocked, 1)
				spin(s.SpinCount)
				if s.locked == 0 {
					return
				}
				atomic.AddInt64(&s.rlocked, -1)
			}
			spin(s.SpinCount)
		}
		runtime.Gosched()
	}
}

func (s *Spinner) RUnlock() {
	atomic.AddInt64(&s.rlocked, -1)
}
