package idmpt

import "time"

type Option func() option

type option func(s *store)

func SetTTL(ttl time.Duration) option {
	return func(s *store) {
		s.ttl = ttl
	}
}
func SetTTLLock(ttl time.Duration) option {
	return func(s *store) {
		s.ttlLock = ttl
	}
}
func EnabledUnique() option {
	return func(s *store) {
		s.unique = true
	}
}
