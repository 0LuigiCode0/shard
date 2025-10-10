// Реализация подхода когда данные хранятся не в одной мапе, а поделены на шарды.
// Предлагается 3 основных типа ключей: числа, массивы, строки
package shard

type iStore[k, v any] interface {
	// Получить данные по ключу
	Get(key k) (data v, err error)
	// Сохранить данные по ключу
	Set(key k, data v) error
	// Grow(count int)
	// Чистит записи в шардах
	Clear()
	// Останавливает удаление по ttl
	Stop()
}

type (
	IStoreNum[k keyNum, v any]              interface{ iStore[k, v] }
	IStoreSeq[t keyNum, k keySeq[t], v any] interface{ iStore[k, v] }
	IStoreStr[k keyStr, v any]              interface{ iStore[k, v] }
)

// Хранилище где ключ это все числовые значения
func NewStoreNum[k keyNum, v any](opts ...fOption) IStoreNum[k, v] {
	return &storeNum[k, v]{*newStore[k, k, v](opts)}
}

// Хранилище где ключ это массивы с любыми числовые значениями и размером до 32
func NewStoreSeq[t keyNum, k keySeq[t], v any](opts ...fOption) IStoreSeq[t, k, v] {
	return &storeSeq[t, k, v]{*newStore[t, k, v](opts)}
}

// Хранилище где ключ это строка
func NewStoreStr[k keyStr, v any](opts ...fOption) IStoreStr[k, v] {
	return &storeStr[k, v]{*newStore[byte, k, v](opts)}
}

func newStore[t keyNum, k _key[t], v any](opts []fOption) *store[t, k, v] {
	s := new(store[t, k, v])
	s.opt.countShard = countShard
	s.opt.minSizeShard = sizeShard
	s.opt.ttl = ttl
	s.opt.expireDelay = expireDelay
	s.stop = make(chan struct{})

	for _, opt := range opts {
		opt(&s.opt)
	}
	s.shards = make([]*shard[t, k, v], s.opt.countShard)
	for i := 0; i < s.opt.countShard; i++ {
		s.shards[i] = &shard[t, k, v]{
			m: make(map[k]*item[v], s.opt.minSizeShard),
		}
	}

	go s.expireDelete()
	return s
}
