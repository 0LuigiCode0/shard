package shard

import (
	"runtime"
	"sync"
	"time"

	"github.com/0LuigiCode0/shard/spinner"
)

type item[v any] struct {
	data v
	ttl  int64
}

type shard[k comparable, v any] struct {
	spin spinner.Spinner

	m     map[k]*item[v]
	count uint
}

type shards[k comparable, v any] []*shard[k, v]

type store[k comparable, v any] struct {
	mx sync.RWMutex

	shards       shards[k, v]
	resizeShards shards[k, v]

	stop chan struct{}

	opt *option[k]
}

// -------------------------------------------------------------------------- //
// MARK:Get/Set
// -------------------------------------------------------------------------- //

func (s *store[k, v]) Set(key k, data v) error {
	if s.resizeShards != nil {
		s.resizeShards.set(s.opt.fGetIndex, key, s.opt.ttl, data)
	}
	return s.shards.set(s.opt.fGetIndex, key, s.opt.ttl, data)
}

func (s *store[k, v]) Get(key k) (data v, err error) {
	if item, ok := s.shards.get(s.opt.fGetIndex, key); ok {
		if time.Now().UnixNano() < item.ttl {
			return item.data, nil
		}
		return def[v](), ErrItemExpired
	}
	return def[v](), ErrItemNotFound
}

func (shs shards[k, v]) get(fGetIndex fGetIndex[k], key k) (it *item[v], ok bool) {
	sh := shs[fGetIndex(uint64(len(shs)), key)]
	sh.spin.RLock()
	defer sh.spin.RUnlock()

	it, ok = sh.m[key]
	return
}

func (shs shards[k, v]) set(fGetIndex fGetIndex[k], key k, ttl time.Duration, data v) error {
	if len(shs) == 0 {
		return ErrNilSlice
	}
	sh := shs[fGetIndex(uint64(len(shs)), key)]

	it := new(item[v])
	it.data = data
	it.ttl = time.Now().Add(ttl).UnixNano()

	sh.spin.Lock()
	defer sh.spin.Unlock()

	if _, ok := sh.m[key]; !ok {
		sh.count++
	}
	sh.m[key] = it

	return nil
}

// -------------------------------------------------------------------------- //
// MARK:Resize
// -------------------------------------------------------------------------- //

func (s *store[k, v]) Resize(countShards uint) {
	var allCount uint
	for _, sh := range s.shards {
		allCount += sh.count
	}

	s.resizeShards = s.makeShards(countShards, allCount/countShards)
	for _, oldSh := range s.shards {
		oldSh.fillShards(s.opt.fGetIndex, s.resizeShards)
	}

	s.opt.countShards = countShards
	s.swapShards()
}

func (sh *shard[k, v]) fillShards(f fGetIndex[k], resizeShards shards[k, v]) {
	sh.spin.RLock()
	defer sh.spin.RUnlock()

	for key, it := range sh.m {
		newSh := resizeShards[f(uint64(len(resizeShards)), key)]
		newSh.itemSet(key, it)
	}
}

func (sh *shard[k, v]) itemSet(key k, it *item[v]) {
	sh.spin.Lock()
	defer sh.spin.Unlock()

	if _, ok := sh.m[key]; !ok {
		sh.m[key] = it
		sh.count++
	}
}

func (s *store[k, v]) swapShards() {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.shards = s.resizeShards
	s.resizeShards = nil
}

// -------------------------------------------------------------------------- //
// MARK:ExpireDelete
// -------------------------------------------------------------------------- //

func (s *store[k, v]) expireDelete() {
	tick := time.NewTicker(s.opt.expireDelay)
	defer tick.Stop()

	for {
		select {
		case <-s.stop:
			return
		case now := <-tick.C:
			for _, b := range s.shards {
				b.expireDelete(now.UnixNano())
			}
		}
	}
}

func (sh *shard[k, v]) expireDelete(now int64) {
	sh.spin.Lock()
	defer sh.spin.Unlock()

	for key, item := range sh.m {
		if item.ttl < now {
			delete(sh.m, key)
			sh.count--
		}
	}
}

// -------------------------------------------------------------------------- //
// MARK:Other
// -------------------------------------------------------------------------- //

func (s *store[k, v]) GetCount() []uint {
	s.mx.RLock()
	defer s.mx.RUnlock()

	out := make([]uint, 0, len(s.shards))
	for _, sh := range s.shards {
		out = append(out, sh.count)
	}

	return out
}

func (s *store[k, v]) Clear() {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, sh := range s.shards {
		sh.m = make(map[k]*item[v], s.opt.minSizeShard)
		sh.count = 0
	}
	runtime.GC()
}

func (s *store[k, v]) Stop() {
	close(s.stop)
}

func (s *store[k, v]) makeShards(countShards, minSizeShard uint) []*shard[k, v] {
	shards := make([]*shard[k, v], countShards)
	for i := uint(0); i < countShards; i++ {
		shards[i] = &shard[k, v]{
			spin: spinner.NewSpinner(s.opt.loopCount, s.opt.spinCount),
			m:    make(map[k]*item[v], minSizeShard),
		}
	}
	return shards
}

// func (s *store[k, v]) Print() {
// 	s.rw.RLock()
// 	defer s.rw.RUnlock()
// 	fmt.Println("shards ", len(s.shards))
// 	for i, sh := range s.shards {
// 		fmt.Print("shard ", i, "_", sh.count, " / ")
// 	}
// 	fmt.Println("\nresize shards ", len(s.resizeShards))
// 	for i, sh := range s.resizeShards {
// 		fmt.Print("shard ", i, "_", sh.count, "/ ")
// 	}
// }
