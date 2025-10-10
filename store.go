package shard

import (
	"runtime"
	"sync"
	"time"
)

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

type (
	storeNum[k keyNum, v any]              struct{ store[k, k, v] }
	storeSeq[t keyNum, k keySeq[t], v any] struct{ store[t, k, v] }
	storeStr[k keyStr, v any]              struct{ store[byte, k, v] }
)

type store[t keyNum, k _key[t], v any] struct {
	shards []*shard[t, k, v]
	rw     sync.RWMutex
	stop   chan struct{}

	opt option
}

type shard[t keyNum, k _key[t], v any] struct {
	m  map[k]*item[v]
	rw sync.RWMutex
}

type item[v any] struct {
	data v
	ttl  int64
}

func (s *storeNum[k, v]) Get(key k) (data v, err error) {
	return s.get(shardNum(s.opt.countShard, key), key)
}

func (s *storeSeq[t, k, v]) Get(key k) (data v, err error) {
	return s.get(shardSeq[t](s.opt.countShard, key), key)
}

func (s *storeStr[k, v]) Get(key k) (data v, err error) {
	return s.get(shardStr(s.opt.countShard, key), key)
}

func (s *storeNum[k, v]) Set(key k, data v) error {
	return s.set(shardNum(s.opt.countShard, key), key, data)
}

func (s *storeSeq[t, k, v]) Set(key k, data v) error {
	return s.set(shardSeq[t](s.opt.countShard, key), key, data)
}

func (s *storeStr[k, v]) Set(key k, data v) error {
	return s.set(shardStr(s.opt.countShard, key), key, data)
}

// Получить данные по ключу
func (s *store[t, k, v]) get(shardNum int, key k) (data v, err error) {
	sh := s.shards[shardNum]
	now := time.Now()
	sh.rw.RLock()
	defer sh.rw.RUnlock()

	if item, ok := sh.m[key]; ok {
		if now.UnixNano() < item.ttl {
			return item.data, nil
		}
		return def[v](), ErrItemExpired

	}
	return def[v](), ErrItemNotFound
}

// // Сохранить данные по ключу
func (s *store[t, k, v]) set(shardNum int, key k, data v) error {
	sh := s.shards[shardNum]
	now := time.Now()
	sh.rw.Lock()
	defer sh.rw.Unlock()

	if item, ok := sh.m[key]; ok {
		if now.UnixNano() < item.ttl {
			item.data = data
			item.ttl = now.Add(s.opt.ttl).UnixNano()
			return nil
		} else {
			return ErrItemExists
		}
	}
	sh.m[key] = &item[v]{
		data: data,
		ttl:  now.Add(s.opt.ttl).UnixNano(),
	}

	return nil
}

// TO-DO: придумать как увиличивать количество шардаов с перестановкой записей и минимальной блокировкой
// func (s *store[t, k, v]) Grow(count int) {
// 	s.rw.Lock()
// 	defer s.rw.Unlock()

// 	for _, b := range s.sh {
// 	}
// }

func (s *store[t, k, v]) Clear() {
	s.rw.Lock()
	defer s.rw.Unlock()

	for _, b := range s.shards {
		b.m = make(map[k]*item[v], s.opt.minSizeShard)
	}
	runtime.GC()
}

func (s *store[t, k, v]) Stop() {
	close(s.stop)
}

// GC
func (s *store[t, k, v]) expireDelete() {
	for {
		select {
		case <-s.stop:
			return
		case now := <-time.After(s.opt.expireDelay):
			for _, b := range s.shards {
				b.expireDelete(now.UnixNano())
			}
		}
	}
}

func (b *shard[t, k, v]) expireDelete(now int64) {
	b.rw.Lock()
	defer b.rw.Unlock()
	for key, item := range b.m {
		if item.ttl < now {
			delete(b.m, key)
		}
	}
}

// Расчет номера шарда исходя их ключа и его типа

func shardNum[k keyNum](countShard int, in k) int {
	return int(in) % countShard
}

func shardSeq[t keyNum, k keySeq[t]](countShard int, in k) int {
	var out int
	for i := 0; i < len(in); i++ {
		out += int(in[i])
	}
	return shardNum(countShard, out)
}

func shardStr[k keyStr](countShard int, in k) int {
	var out int
	for i := 0; i < len(in); i++ {
		out += int(in[i])
	}
	return shardNum(countShard, out)
}
