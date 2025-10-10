package shard

import (
	"sync"
	"time"
)

type testMap struct {
	m  map[int]*it
	rw sync.RWMutex
}

type it struct {
	data string
	ttl  int64
}

func newTestMap(size int) *testMap {
	return &testMap{
		m: make(map[int]*it, size),
	}
}

func (tm *testMap) set(key int, data string) {
	now := time.Now()
	tm.rw.Lock()
	defer tm.rw.Unlock()

	if v, ok := tm.m[key]; ok {
		v.data = data
		v.ttl = now.Add(time.Second).UnixNano()
		return
	}
	tm.m[key] = &it{data: data, ttl: now.Add(time.Second).UnixNano()}
}

func (tm *testMap) get(key int) string {
	tm.rw.RLock()
	defer tm.rw.RUnlock()

	if v, ok := tm.m[key]; ok {
		return v.data
	}
	return ""
}
