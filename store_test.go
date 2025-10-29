package shard

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/0LuigiCode0/shard/spinner"
	"github.com/google/uuid"
)

// -------------------------------------------------------------------------- //
// MARK: Test
// -------------------------------------------------------------------------- //

func TestStoreNum(t *testing.T) {
	type x int

	s := NewStore[string](Option[x]().SetTTL(time.Second * 5).SetCountShards(3).SetExpireDelay(time.Second))

	t.Log(s.Set(4, "hello"))
	t.Log(s.Set(2, "world"))
	t.Log(s.Get(4))
	t.Log(s.Get(2))
	s.Stop()
	fmt.Println()
}

func TestStoreSeq(t *testing.T) {
	u1 := uuid.New()
	u2 := uuid.New()

	s := NewStore[string](Option[uuid.UUID]().SetStartSizeShard(256))

	t.Log(s.Set(u1, "hello"))
	t.Log(s.Set(u2, "world"))
	t.Log(s.Get(u1))
	t.Log(s.Get(u2))
	s.Stop()
}

func TestStoreStr(t *testing.T) {
	type myString string
	s := NewStore[string](Option[myString]().SetTTL(time.Second * 2))
	t.Log(s.Set("one", "hello"))
	t.Log(s.Set("two", "world"))
	t.Log(s.Get("two"))
	t.Log(s.Get("one"))
	s.Stop()
}

func TestResize(t *testing.T) {
	u1 := uuid.New()
	u2 := uuid.New()

	s := NewStore[string](Option[uuid.UUID]().SetCountShards(8))
	s.Stop()

	t.Log(s.Set(u1, "hello"))
	t.Log(s.Set(u2, "world"))
	// s.Print()

	s.Resize(13)
	// s.Print()

	t.Log(s.Get(u1))
	t.Log(s.Get(u2))

	s.Resize(3)
	// s.Print()

	t.Log(s.Get(u1))
	t.Log(s.Get(u2))
}

func TestResizeGO(t *testing.T) {
	// runtime.GOMAXPROCS(6)
	wg := sync.WaitGroup{}

	s := NewStore[string](Option[int32]().SetCountShards(20).SetStartSizeShard(1000))
	s.Stop()

	s.Set(100001, "world")

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 1000000 {
			s.Set(rand.Int31n(100000), "hello")
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 1000000 {
			s.Get(rand.Int31n(100000))
		}
	}()

	time.Sleep(time.Millisecond * 100)
	// s.Print()
	s.Resize(9)
	// s.Print()
	t.Log(s.Get(100001))
	wg.Wait()
	// s.Print()
}

// -------------------------------------------------------------------------- //
// MARK:Bench
// -------------------------------------------------------------------------- //

const cpu = 24

func BenchmarkShardInt(b *testing.B) {
	runtime.GOMAXPROCS(cpu)

	wg := sync.WaitGroup{}
	s := NewStore[string](Option[int32]().SetCountShards(1000).SetStartSizeShard(1000))
	s.Stop()

	b.ResetTimer()
	b.StopTimer()
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Set(rand.Int31n(1000000), "hello")
			}
		})
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Get(rand.Int31n(1000000))
			}
		})
	}()
	wg.Wait()
}

func BenchmarkMapIntSpinner(b *testing.B) {
	runtime.GOMAXPROCS(cpu)

	wg := sync.WaitGroup{}
	spin := spinner.NewSpinner(cpu, 0)
	s := newTestMap[int32](1000*1000, &spin)

	b.ResetTimer()
	b.StopTimer()
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Set(rand.Int31n(1000000), "hello")
			}
		})
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Get(rand.Int31n(1000000))
			}
		})
	}()
	wg.Wait()
}

func BenchmarkMapIntMutex(b *testing.B) {
	runtime.GOMAXPROCS(cpu)

	wg := sync.WaitGroup{}
	s := newTestMap[int32](1000*1000, &sync.RWMutex{})

	b.ResetTimer()
	b.StopTimer()
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Set(rand.Int31n(1000000), "hello")
			}
		})
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Get(rand.Int31n(1000000))
			}
		})
	}()
	wg.Wait()
}

func BenchmarkShardUUID(b *testing.B) {
	runtime.GOMAXPROCS(cpu)

	wg := sync.WaitGroup{}
	s := NewStore[string](Option[uuid.UUID]().SetCountShards(1000).SetStartSizeShard(1000))
	s.Stop()

	b.ResetTimer()
	b.StopTimer()
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Set(uuid.New(), "hello")
			}
		})
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Get(uuid.New())
			}
		})
	}()
	wg.Wait()
}

func BenchmarkMapUUIDSpinner(b *testing.B) {
	runtime.GOMAXPROCS(cpu)

	wg := sync.WaitGroup{}
	spin := spinner.NewSpinner(cpu, 3)
	s := newTestMap[uuid.UUID](1000*1000, &spin)

	b.ResetTimer()
	b.StopTimer()
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Set(uuid.New(), "hello")
			}
		})
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Get(uuid.New())
			}
		})
	}()
	wg.Wait()
}

func BenchmarkMapUUIDMutex(b *testing.B) {
	runtime.GOMAXPROCS(cpu)

	wg := sync.WaitGroup{}
	s := newTestMap[uuid.UUID](1000*1000, &sync.RWMutex{})

	b.ResetTimer()
	b.StopTimer()
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Set(uuid.New(), "hello")
			}
		})
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Get(uuid.New())
			}
		})
	}()
	wg.Wait()
}
