package shard

import (
	"reflect"
	"runtime"
	"time"
)

type option[k comparable] struct {
	ttl          time.Duration
	countShards  uint
	minSizeShard uint
	expireDelay  time.Duration

	loopCount byte
	spinCount byte

	fGetIndex fGetIndex[k]
}

func Option[k comparable]() *option[k] {
	cpus := runtime.NumCPU()
	if cpus < 4 {
		cpus = 4
	}

	fGetIndex := getIndex[k]
	if reflect.TypeFor[k]().Kind() == reflect.String {
		fGetIndex = getIndexStr[k]
	}

	return &option[k]{
		countShards:  countShards,
		minSizeShard: sizeShard,
		ttl:          ttl,
		expireDelay:  expireDelay,
		loopCount:    byte(cpus),
		spinCount:    2,
		fGetIndex:    fGetIndex,
	}
}

// Устанавливает время жизни записи
func (o *option[k]) SetTTL(ttl time.Duration) *option[k] {
	o.ttl = ttl
	return o
}

// Устанавливает количество шардаов
func (o *option[k]) SetCountShards(count uint) *option[k] {
	o.countShards = count
	return o
}

// Устанавливает начальный размер шардаов
func (o *option[k]) SetStartSizeShard(size uint) *option[k] {
	o.minSizeShard = size
	return o
}

// Устанавливает задержку проверки просрочки
func (o *option[k]) SetExpireDelay(delay time.Duration) *option[k] {
	o.expireDelay = delay
	return o
}

func (o *option[k]) SetLoopCount(count byte) *option[k] {
	o.loopCount = count
	return o
}

func (o *option[k]) SetSpinCount(count byte) *option[k] {
	o.spinCount = count
	return o
}

func (o *option[k]) SetFuncGetIndex(fGetIndex fGetIndex[k]) *option[k] {
	o.fGetIndex = fGetIndex
	return o
}
