package shard

import (
	"hash/crc64"
	"unsafe"
)

type fGetIndex[k comparable] func(countShard uint64, key k) uint64

func getIndex[k comparable](countShard uint64, key k) uint64 {
	hash := crc64.New(crc64.MakeTable(crc64.ECMA))
	hash.Write(unsafe.Slice((*byte)(unsafe.Pointer(&key)), unsafe.Sizeof(key)))
	return hash.Sum64() % countShard
}

func getIndexStr[k any](countShard uint64, key k) uint64 {
	hash := crc64.New(crc64.MakeTable(crc64.ECMA))
	hash.Write([]byte(*(*string)(unsafe.Pointer(&key))))
	return hash.Sum64() % countShard
}
