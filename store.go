package idmpt

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type IStore interface {
	// Получить данные по ключу
	Get(key uuid.UUID) (data []byte, err error)
	// Сохранить данные по ключу
	Set(key uuid.UUID, data []byte) error
	// Создать презапись сохранения данных
	Lock(key uuid.UUID) error
	// Удалить презапись сохранения данных
	Unlock(key uuid.UUID)
	Reset()
	Clear()
}

type store struct {
	m  map[byte]*bucket
	rw sync.RWMutex

	countBucket byte
	ttl         time.Duration
	ttlLock     time.Duration
	unique      bool

	minCount  atomic.Int32
	minBucket byte
}

type bucket struct {
	m  map[uuid.UUID]*storeItem
	rw sync.RWMutex

	count atomic.Int32
}

type storeItem struct {
	data   []byte
	ttl    int64
	isSave bool
}

func NewStore(countBucket byte, opts ...option) IStore {
	s := new(store)
	if countBucket == 0 {
		countBucket = 8
	}
	s.m = make(map[byte]*bucket, countBucket)
	for i := byte(0); i < 8; i++ {
		s.m[i] = &bucket{
			m: make(map[uuid.UUID]*storeItem, 128),
		}
	}
	s.countBucket = countBucket
	s.ttl = time.Hour
	s.ttlLock = time.Second * 10

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Получить данные по ключу
func (s *store) Get(key uuid.UUID) (data []byte, err error) {
	if item, ok := s.m[key[0]%s.countBucket].m[key]; ok {
		if item.isSave {
			if time.Now().UnixNano() < item.ttl {
				return item.data, nil
			}
			return nil, ItemExpiredError
		} else {
			return nil, LockExistsError
		}
	}
	return nil, ItemNotFound
}

// Сохранить данные по ключу
func (s *store) Set(key uuid.UUID, data []byte) error {
	bucket := s.m[key[0]%s.countBucket]
	now := time.Now()
	bucket.rw.Lock()
	defer bucket.rw.Unlock()

	if item, ok := bucket.m[key]; ok {
		if unix := now.UnixNano(); item.isSave && s.unique && unix < item.ttl {
			return ItemExistsError
		} else {
			item.data = data
			item.isSave = true
			item.ttl = now.Add(calculateTTL(s.ttl)).UnixNano()
			return nil
		}
	}

	return ItemNotFound
}

// Создать презапись сохранения данных
func (s *store) Lock(key uuid.UUID) error {
	bucket := s.m[key[0]%s.countBucket]
	now := time.Now()
	bucket.rw.Lock()
	defer bucket.rw.Unlock()

	if v, ok := bucket.m[key]; ok {
		if v.isSave {
			if now.UnixNano() < v.ttl {
				return LockExistsSaveError
			}
		} else {
			return LockExistsError
		}
	}

	item := new(storeItem)
	item.ttl = now.Add(calculateTTL(s.ttlLock)).UnixNano()

	bucket.m[key] = item
	return nil
}

// Удалить презапись сохранения данных
func (s *store) Unlock(key uuid.UUID) {
	bucket := s.m[key[0]%s.countBucket]
	bucket.rw.Lock()
	defer bucket.rw.Unlock()

	if v, ok := bucket.m[key]; ok {
		if !v.isSave {
			delete(bucket.m, key)
		}
	}
}

func (s *store) Reset() {
	for _, b := range s.m {
		b.m = make(map[uuid.UUID]*storeItem, 128)
	}
}

func (s *store) Clear() {
	for _, b := range s.m {
		b.m = nil
	}
	s.m = nil
}

// GC
func (s *store) expireDelete() {
	now := time.Now().UnixNano()
	for _, b := range s.m {
		b.rw.Lock()
		for key, item := range b.m {
			if item.ttl >= now {
				delete(b.m, key)
			}
		}
		b.rw.Unlock()
	}
}
