# shard

Библиотека пригодится для in memory cache. В отличии с стандартным подходом из мапы обернутой в мьютекс или sync.Map, я предлогаю подход когда данные хранятся не в одной мапе, а поделены на шарды. Каждый шард это lock-free структура. Этот подход сильно ускоряет работу с данными при частой записи.

В качестве ключа может использоваться любой сравниваемый тип. Из ключа через хеш сумму высчитывается индекс шарда, вы можете переопределить эту функцию.

```golang
Примеры инициализации

// Простой тип ключа
NewStore[string, int32](nil) // k: int32 v: string

// Alias тип ключа
type myString string
NewStore[string, myString](nil) // k: myString v: string

// Своя хеш функция 
NewStore[string]( // k: struct{ A int } v: string
   Options[struct{ A int }]().
      SetFuncGetIndex(getIndex),
)
func getIndex(countShard uint64, k struct{ A int }) int { return k.A }
```

Записи с просроченным ttl удаляются автоматически. Периодичность проверки можно задать через опции.

Если в процессе выполнения программы нужно увеличить или уменьшить хранилище, то есть метод `Resize`, под капотом создается временный пул новых шардов с стартовым размером как `allCount / countShards`. Работает по аналогии с мапой поэтому задержки минимальны.

Для статистики метод `GetCount` вернет количество записей в каждом шарде.

Для тех кому нужен только блокировщик есть `github.com/0LuigiCode0/shard/spinner` это блокировки на атомиках с спинлоком на ассемблере

Сравнение бенчмарков

```golang
При 24 ядрах
cpu: AMD Ryzen 9 7900X 12-Core Processor
BenchmarkShardInt-24           10000000                34.86 ns/op
BenchmarkMapIntSpinner-24      10000000               326.9 ns/op
BenchmarkMapIntMutex-24        10000000               301.1 ns/op
BenchmarkShardUUID-24           10000000                64.91 ns/op
BenchmarkMapUUIDSpinner-24      10000000               539.9 ns/op
BenchmarkMapUUIDMutex-24        10000000               597.6 ns/op

При 4 ядрах
cpu: AMD Ryzen 9 7900X 12-Core Processor
BenchmarkShardInt-24           10000000                69.79 ns/op
BenchmarkMapIntSpinner-24      10000000               190.8 ns/op
BenchmarkMapIntMutex-24        10000000               375.0 ns/op
BenchmarkShardUUID-24           10000000               164.1 ns/op
BenchmarkMapUUIDSpinner-24      10000000               358.7 ns/op
BenchmarkMapUUIDMutex-24        10000000               664.2 ns/op
```
