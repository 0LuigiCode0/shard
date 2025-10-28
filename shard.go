package shard

type IStore[k comparable, v any] interface {
	// Получить данные по ключу
	Get(key k) (data v, err error)
	// Сохранить данные по ключу
	Set(key k, data v) error
	// Изменят количество шардов с пересчетом ключей
	Resize(countShards int)
	// Чистит записи в шардах
	Clear()
	// Останавливает удаление по ttl
	Stop()

	// Print()
}

func NewStore[v any, k comparable](fGetIndex fGetIndex[k], opts ...fOption) IStore[k, v] {
	s := new(store[k, v])
	s.opt.countShards = countShards
	s.opt.minSizeShard = sizeShard
	s.opt.ttl = ttl
	s.opt.expireDelay = expireDelay
	s.stop = make(chan struct{})

	s.fGetIndex = fGetIndex

	for _, opt := range opts {
		opt(&s.opt)
	}
	s.shards = s.makeShards(s.opt.countShards, s.opt.minSizeShard)

	go s.expireDelete()
	return s
}
