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
}

func spin(cycle int32)

// -------------------------------------------------------------------------- //
// MARK:Lock/Unlock
// -------------------------------------------------------------------------- //

var (
	loopCount int
	spinCount int32 = 2
)

func init() {
	cpus := runtime.NumCPU()
	if cpus < 4 {
		cpus = 4
	}
	loopCount = cpus
}

func SetLoopCount(count int)   { loopCount = count }
func SetSpinCount(count int32) { spinCount = count }

func (s *Spinner) Lock() {
	for {
		for range loopCount {
			if s.locked == 0 && atomic.CompareAndSwapUint32(&s.locked, 0, 1) {
				for s.rlocked != 0 {
					spin(spinCount)
				}
				return
			}
			spin(spinCount)
		}
		runtime.Gosched()
	}
}

func (s *Spinner) Unlock() {
	atomic.StoreUint32(&s.locked, 0)
}

func (s *Spinner) RLock() {
	for {
		for range loopCount {
			if s.locked == 0 {
				atomic.AddInt64(&s.rlocked, 1)
				spin(spinCount)
				if s.locked == 0 {
					return
				}
				atomic.AddInt64(&s.rlocked, -1)
			}
			spin(spinCount)
		}
		runtime.Gosched()
	}
}

func (s *Spinner) RUnlock() {
	atomic.AddInt64(&s.rlocked, -1)
}
