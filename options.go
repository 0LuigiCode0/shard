package shard

import (
	"time"

	"github.com/0LuigiCode0/shard/spinner"
)

type fOption func(s *option)

type option struct {
	ttl          time.Duration
	countShards  int
	minSizeShard int
	expireDelay  time.Duration
}

// Устанавливает время жизни записи
func SetTTL(ttl time.Duration) fOption {
	return func(o *option) {
		o.ttl = ttl
	}
}

// Устанавливает количество шардаов
func SetCountShards(count int) fOption {
	return func(o *option) {
		o.countShards = count
	}
}

// Устанавливает начальный размер шардаов
func SetStartSizeShard(size int) fOption {
	return func(o *option) {
		o.minSizeShard = size
	}
}

// Устанавливает задержку проверки просрочки
func SetExpireDelay(delay time.Duration) fOption {
	return func(o *option) {
		o.expireDelay = delay
	}
}

func SetLoopCount(count int) fOption {
	return func(_ *option) {
		spinner.SetLoopCount(count)
	}
}

func SetSpinCount(count int32) fOption {
	return func(_ *option) {
		spinner.SetSpinCount(count)
	}
}
