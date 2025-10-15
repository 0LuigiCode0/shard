package shard

import (
	"math/rand"
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
	s := NewStoreSeq[byte, uuid.UUID, string](SetMinSizeShard(256))
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
	wg := sync.WaitGroup{}
	s := NewStoreNum[int, string](SetCountShards(20), SetTTL(time.Second*5))
	rand.New(rand.NewSource(42))

	t.Log(s.Set(4, "hello"))

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 3000 {
			s.Set(rand.Int(), "hello")
			time.Sleep(time.Millisecond)
		}
	}()
	time.Sleep(time.Second)
	s.print()
	s.Resize(10)
	s.print()
	t.Log(s.Get(4))

	s.Stop()
	wg.Wait()
}

func BenchmarkStoreNum(b *testing.B) {
	s := NewStoreNum[int, string](SetMinSizeShard(1000), SetTTL(time.Second*6), SetCountShards(10000), SetExpireDelay(time.Second*3))
	defer s.Stop()
	rand.New(rand.NewSource(42))

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s.Set(rand.Int(), "hello")
		}
	})
}

func BenchmarkMap(b *testing.B) {
	s := newTestMap(10000 * 1000)
	rand.New(rand.NewSource(42))

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s.set(rand.Int(), "hello")
		}
	})
}
