package shard

import "time"

type fOption func(s *option)

type option struct {
	ttl          time.Duration
	countShard   int
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
func SetShard(count int) fOption {
	return func(o *option) {
		o.countShard = count
	}
}

// Устанавливает начальный размер шардаов
func SetMinSizeShard(size int) fOption {
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
