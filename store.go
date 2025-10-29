package shard

type IStore[k comparable, v any] interface {
	// Получить данные по ключу
	Get(key k) (data v, err error)
	// Сохранить данные по ключу
	Set(key k, data v) error
	// Изменят количество шардов с пересчетом ключей
	Resize(countShards uint)
	// Чистит записи в шардах
	Clear()
	// Останавливает удаление по ttl
	Stop()
	// Возвращает количество записей в каждом шарде
	GetCount() []uint
	// Print()
}

func NewStore[v any, k comparable](opt *option[k]) IStore[k, v] {
	s := new(store[k, v])
	s.stop = make(chan struct{})

	if opt == nil {
		opt = Option[k]()
	}
	s.opt = opt

	s.shards = s.makeShards(s.opt.countShards, s.opt.minSizeShard)

	go s.expireDelete()
	return s
}
