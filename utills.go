package shard

func def[v any]() v {
	out := new(v)
	return *out
}
