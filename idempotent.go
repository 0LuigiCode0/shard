// Реализует кеш для работы с идемпотентностью
//
//	Вся работа осуществляется через idmpt.Cache, при старте сервиса необходимо вызвать InitCache для инициализации кеша и запуска мусорщика (мусорщик работает с периодичностью параметра LockDelay тоесть время жизни блокировки)
//
// Логика работы:
//  1. Get - ищет в кеше запись по ключу, если находит то возвращает данные либо ошибку если это блокировка, иначе идет дальше
//  2. Lock - выполняется первым, создает запись в кеше с небольшим временем жизни. Нужна для отбрасывания одинаковых запросов (тк первый запрос заблокировал ключ раньше то последующие будут падать с ошибкой, пока не выполнится первый)
//  3. handle - функция идущая выше по стеку, возвращает ошибку и данные
//  4. Unlock - если handle вернул ошибку то убираем блокировку с ключа, чтобы новые запросы (например из ретрая) прошли без проблем
//  5. Save - сохранят данные по ключу, убирая блокировку, и устанавливает новое значение времени жизни
package idmpt

import (
	"context"
	"time"
)

// Инстанс кеша для идемпотентности
var Cache *idmptCache

type idmptCache struct {
	s         *store
	lockDelay time.Duration
	ttl       time.Duration
	ctx       context.Context
}

type Config struct {
	LockDelay time.Duration
	Ttl       time.Duration
	SizeStore int
}

func init() {
	Cache = &idmptCache{
		s:         &store{},
		lockDelay: time.Second,
		ttl:       time.Minute,
	}
}

// Инициализация кеша и установка своих параметров
func InitCache(ctx context.Context, conf *Config) {
	if conf.SizeStore > 0 {
		sizeStore = conf.SizeStore
	}
	if conf.LockDelay > 0 {
		Cache.lockDelay = conf.LockDelay
	}
	if conf.Ttl > 0 {
		Cache.ttl = conf.Ttl
	}
	Cache.ctx = ctx
	Cache.s.m = make(map[string]*storeItem, sizeStore)
	Cache.s.keysForDeleted = make(map[string]struct{}, sizeStore)

	go Cache.expireDelete()
}

//

// Получить данные по ключу
func (obs *idmptCache) Get(key string) (any, error) {
	return obs.s.get(key)
}

// Сохранить данные по ключу
func (obs *idmptCache) Save(data any, key string) error {
	return obs.s.save(data, key, obs.ttl)
}

// Создать презапись сохранения данных
func (obs *idmptCache) Lock(key string) error {
	return obs.s.lock(key, obs.lockDelay)
}

// Удалить презапись сохранения данных
func (obs *idmptCache) Unlock(key string) {
	obs.s.unlock(key)
	return
}

//

// GC
func (obs *idmptCache) expireDelete() {
	ticker := time.NewTicker(obs.lockDelay)
	defer ticker.Stop()
	for {
		select {
		case <-obs.ctx.Done():
			return
		case <-ticker.C:
			obs.s.expireDelete()
		}
	}
}

//

// Реализует работу с кешом
