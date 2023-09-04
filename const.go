package idmpt

import (
	"errors"
)

// Статус записей в кеше
type state int

const (
	lock state = iota
	save
)

//

var (
	// дефолтный минимальный размер кеша
	sizeStore   = 256
	reasonError = "cache error"

	// Ключ для хедера
	HeaderIdempotentKey = "X-MTS-IDMPT-KEY"
)

//

var (
	LockExistsError     = errors.New("lock already exists")
	LockNotFoundError   = errors.New("lock not found")
	LockExistsSaveError = errors.New("can't create lock, item already save")
	ItemExistsError     = errors.New("item already exists")
	ItemNotFound        = errors.New("item not found")
	ItemExpiredError    = errors.New("item already expired")
)

const (
	CacheErrorCode = iota + 100
)
