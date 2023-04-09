package pizza

import "time"

type CacheValue[V any] struct {
	val       V
	createdAt time.Time
}

type Cache[T any] struct {
	ttl     time.Duration
	store   map[string]CacheValue[T]
	refresh func(key string) (T, error)
}

func NewCache[T any](ttl time.Duration, refreshFunc func(key string) (T, error)) Cache[T] {
	return Cache[T]{
		ttl:     ttl,
		store:   make(map[string]CacheValue[T]),
		refresh: refreshFunc,
	}
}

func (c *Cache[T]) Get(key string) (T, error) {
	v, ok := c.store[key]
	if !ok || v.createdAt.Add(c.ttl).Before(time.Now()) {
		newVal, err := c.refresh(key)
		if err != nil {
			return *(new(T)), err
		}
		v = CacheValue[T]{newVal, time.Now()}
		c.store[key] = v
		return newVal, nil
	} else {
		return v.val, nil
	}
}
