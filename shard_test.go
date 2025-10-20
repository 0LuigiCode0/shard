package shard

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStoreNum(t *testing.T) {
	type x int

	s := NewStoreNum[x, string](SetTTL(time.Second*5), SetCountShards(3), SetExpireDelay(time.Second))
	t.Log(s.Set(4, "hello"))
	t.Log(s.Set(2, "world"))
	t.Log(s.Get(4))
	t.Log(s.Get(2))
	s.Stop()
}

func TestStoreSeq(t *testing.T) {
	u1 := uuid.New()
	u2 := uuid.New()
	s := NewStoreSeq[byte, uuid.UUID, string](SetStartSizeShard(256))
	t.Log(s.Set(u1, "hello"))
	t.Log(s.Set(u2, "world"))
	t.Log(s.Get(u1))
	t.Log(s.Get(u2))
	s.Stop()
}

func TestStoreStr(t *testing.T) {
	s := NewStoreStr[string, string](SetTTL(time.Second * 2))
	t.Log(s.Set("one", "hello"))
	t.Log(s.Set("two", "world"))
	t.Log(s.Get("two"))
	t.Log(s.Get("one"))
	s.Stop()
}

func TestResize(t *testing.T) {
	uuid.SetRand(rand.New(rand.NewSource(time.Now().Unix())))
	u1 := uuid.New()
	u2 := uuid.New()

	s := NewStoreSeq[byte, uuid.UUID, string](SetCountShards(8))
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
	runtime.GOMAXPROCS(6)
	wg := sync.WaitGroup{}

	s := NewStoreNum[int32, string](SetCountShards(20), SetStartSizeShard(1000))
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

func BenchmarkStoreNum(b *testing.B) {
	runtime.GOMAXPROCS(4)
	wg := sync.WaitGroup{}

	s := NewStoreNum[int32, string](SetCountShards(1000), SetStartSizeShard(1000))
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

func BenchmarkMap(b *testing.B) {
	runtime.GOMAXPROCS(4)

	wg := sync.WaitGroup{}
	s := newTestMap(1000 * 1000)

	b.ResetTimer()
	b.StopTimer()
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Set(rand.Intn(1000000), "hello")
			}
		})
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				s.Get(rand.Intn(1000000))
			}
		})
	}()
	wg.Wait()
}
