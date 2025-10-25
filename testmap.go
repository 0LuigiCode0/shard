package shard

import (
	"time"
)

type locker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type testMap[t keyNum, k _key[t]] struct {
	m     map[k]*testitem
	rw    locker
	count int
}

type testitem struct {
	data string
	ttl  int64
}

func newTestMap[t keyNum, k _key[t]](size int, rw locker) *testMap[t, k] {
	return &testMap[t, k]{
		m:  make(map[k]*testitem, size),
		rw: rw,
	}
}

func (tm *testMap[t, k]) Set(key k, data string) {
	it := new(testitem)
	it.data = data
	it.ttl = time.Now().Add(ttl).UnixNano()

	tm.rw.Lock()
	defer tm.rw.Unlock()

	if _, ok := tm.m[key]; !ok {
		tm.count++
	}
	tm.m[key] = it
}

func (tm *testMap[t, k]) Get(key k) string {
	tm.rw.RLock()
	defer tm.rw.RUnlock()

	if v, ok := tm.m[key]; ok {
		return v.data
	}
	return ""
}
