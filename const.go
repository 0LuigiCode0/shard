package shard

import (
	"errors"
	"time"
)

//

var (
	// дефолтное количество	шардаов
	countShards = 128
	// дефолтный минимальный размер шарда
	sizeShard = 128
	// дефолтное ttl
	ttl = time.Minute
	// дефолтная задержка проверки просрочки
	expireDelay = time.Minute
)

//

var (
	ErrItemExists   = errors.New("item already exists")
	ErrItemNotFound = errors.New("item not found")
	ErrItemExpired  = errors.New("item already expired")
)
