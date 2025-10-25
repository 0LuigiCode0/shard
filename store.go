package shard

import (
	"runtime"
	"sync"
	"time"

	"github.com/0LuigiCode0/shard/spinner"
)

// -------------------------------------------------------------------------- //
// MARK:Интерфейсы ключей
// -------------------------------------------------------------------------- //

type (
	keyStr interface{ ~string }
	keyNum interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
			~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
			~float32 | ~float64
	}
	keySeq[t keyNum] interface {
		~[1]t | ~[2]t | ~[3]t | ~[4]t | ~[5]t | ~[6]t | ~[7]t | ~[8]t |
			~[9]t | ~[10]t | ~[11]t | ~[12]t | ~[13]t | ~[14]t | ~[15]t | ~[16]t |
			~[17]t | ~[18]t | ~[19]t | ~[20]t | ~[21]t | ~[22]t | ~[23]t | ~[24]t |
			~[25]t | ~[26]t | ~[27]t | ~[28]t | ~[29]t | ~[30]t | ~[31]t | ~[32]t
	}
	_key[t keyNum] interface {
		keyNum | keySeq[t] | keyStr
	}
)

// -------------------------------------------------------------------------- //
// MARK:Типы хранилищ
// -------------------------------------------------------------------------- //

type item[v any] struct {
	data v
	ttl  int64
}

type shard[t keyNum, k _key[t], v any] struct {
	rw spinner.Spinner

	m     map[k]*item[v]
	count int
}

type shards[t keyNum, k _key[t], v any] []*shard[t, k, v]

type store[t keyNum, k _key[t], v any] struct {
	rw sync.RWMutex

	shards       shards[t, k, v]
	resizeShards shards[t, k, v]

	stop chan struct{}

	opt option
}

type (
	storeNum[k keyNum, v any]              struct{ store[k, k, v] }
	storeSeq[t keyNum, k keySeq[t], v any] struct{ store[t, k, v] }
	storeStr[k keyStr, v any]              struct{ store[byte, k, v] }
)

// -------------------------------------------------------------------------- //
// MARK:Get/Set
// -------------------------------------------------------------------------- //

func (s *storeNum[k, v]) Get(key k) (data v, err error)    { return s.get(s.getIndex, key) }
func (s *storeSeq[t, k, v]) Get(key k) (data v, err error) { return s.get(s.getIndex, key) }
func (s *storeStr[k, v]) Get(key k) (data v, err error)    { return s.get(s.getIndex, key) }

func (s *storeNum[k, v]) Set(key k, data v) error    { return s.set(s.getIndex, key, data) }
func (s *storeSeq[t, k, v]) Set(key k, data v) error { return s.set(s.getIndex, key, data) }
func (s *storeStr[k, v]) Set(key k, data v) error    { return s.set(s.getIndex, key, data) }

func (s *store[t, k, v]) set(f fGetIndex[t, k, v], key k, data v) error {
	if s.resizeShards != nil {
		s.resizeShards.set(f, key, s.opt.ttl, data)
	}
	return s.shards.set(f, key, s.opt.ttl, data)
}

func (s *store[t, k, v]) get(f fGetIndex[t, k, v], key k) (data v, err error) {
	if item, ok := s.shards.get(f, key); ok {
		if time.Now().UnixNano() < item.ttl {
			return item.data, nil
		}
		return def[v](), ErrItemExpired
	}
	return def[v](), ErrItemNotFound
}

func (shs *shards[t, k, v]) get(f fGetIndex[t, k, v], key k) (it *item[v], ok bool) {
	sh := (*shs)[f(len(*shs), key)]
	sh.rw.RLock()
	defer sh.rw.RUnlock()

	it, ok = sh.m[key]
	return
}

func (shs *shards[t, k, v]) set(f fGetIndex[t, k, v], key k, ttl time.Duration, data v) error {
	sh := (*shs)[f(len(*shs), key)]

	it := new(item[v])
	it.data = data
	it.ttl = time.Now().Add(ttl).UnixNano()

	sh.rw.Lock()
	defer sh.rw.Unlock()

	if _, ok := sh.m[key]; !ok {
		sh.count++
	}
	sh.m[key] = it

	return nil
}

// -------------------------------------------------------------------------- //
// MARK:Resize
// -------------------------------------------------------------------------- //

func (s *storeNum[k, v]) Resize(countShards int)    { s.resize(s.getIndex, countShards) }
func (s *storeSeq[t, k, v]) Resize(countShards int) { s.resize(s.getIndex, countShards) }
func (s *storeStr[k, v]) Resize(countShards int)    { s.resize(s.getIndex, countShards) }

func (s *store[t, k, v]) resize(getIndex fGetIndex[t, k, v], countShards int) {
	var allCount int
	for _, sh := range s.shards {
		allCount += sh.count
	}

	s.resizeShards = s.makeShards(countShards, allCount/countShards)
	for _, oldSh := range s.shards {
		oldSh.fillShards(getIndex, s.resizeShards)
	}

	s.opt.countShards = countShards
	s.swapShards()
}

func (sh *shard[t, k, v]) fillShards(getIndex fGetIndex[t, k, v], resizeShards shards[t, k, v]) {
	sh.rw.RLock()
	defer sh.rw.RUnlock()

	for key, it := range sh.m {
		newSh := resizeShards[getIndex(len(resizeShards), key)]
		newSh.itemSet(key, it)
	}
}

func (sh *shard[t, k, v]) itemSet(key k, it *item[v]) {
	sh.rw.Lock()
	defer sh.rw.Unlock()

	if _, ok := sh.m[key]; !ok {
		sh.m[key] = it
		sh.count++
	}
}

func (s *store[t, k, v]) swapShards() {
	s.rw.Lock()
	defer s.rw.Unlock()

	s.shards = s.resizeShards
	s.resizeShards = nil
}

// -------------------------------------------------------------------------- //
// MARK:ExpireDelete
// -------------------------------------------------------------------------- //

func (s *store[t, k, v]) expireDelete() {
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

func (sh *shard[t, k, v]) expireDelete(now int64) {
	sh.rw.Lock()
	defer sh.rw.Unlock()

	for key, item := range sh.m {
		if item.ttl < now {
			delete(sh.m, key)
			sh.count--
		}
	}
}

// -------------------------------------------------------------------------- //
// MARK:GetIndex
// -------------------------------------------------------------------------- //

type fGetIndex[t keyNum, k _key[t], v any] func(countShard int, key k) int

func getIndex(countShard, key int) int { return int(key) % countShard }

func (s *storeNum[k, v]) getIndex(countShard int, key k) int {
	return getIndex(countShard, int(key))
}

func (s *storeSeq[t, k, v]) getIndex(countShard int, key k) int {
	var sum int
	for i := 0; i < len(key); i++ {
		sum += int(key[i])
	}
	return getIndex(countShard, sum)
}

func (s *storeStr[k, v]) getIndex(countShard int, key k) int {
	var sum int
	for i := 0; i < len(key); i++ {
		sum += int(key[i])
	}
	return getIndex(countShard, sum)
}

// -------------------------------------------------------------------------- //
// MARK:Other
// -------------------------------------------------------------------------- //

func (s *store[t, k, v]) Clear() {
	s.rw.Lock()
	defer s.rw.Unlock()

	for _, sh := range s.shards {
		sh.m = make(map[k]*item[v], s.opt.minSizeShard)
		sh.count = 0
	}
	runtime.GC()
}

func (s *store[t, k, v]) Stop() {
	close(s.stop)
}

func (s *store[t, k, v]) makeShards(countShards, minSizeShard int) []*shard[t, k, v] {
	shards := make([]*shard[t, k, v], countShards)
	for i := 0; i < countShards; i++ {
		shards[i] = &shard[t, k, v]{
			m: make(map[k]*item[v], minSizeShard),
		}
	}
	return shards
}

// func (s *store[t, k, v]) Print() {
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
