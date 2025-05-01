package cache

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrorCacheMiss = errors.New("Cache miss")
)

type item struct {
	value     interface{}
	ttl       int64
	createdAt int64
}

type MemoryCache struct {
	cache map[interface{}]*item
	sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		cache: make(map[interface{}]*item),
	}
	go mc.setTimeout()
	return mc
}

func (mc *MemoryCache) setTimeout() {
	for {
		mc.Lock()
		for key, _ := range mc.cache {
			if time.Now().Unix()-mc.cache[key].ttl > mc.cache[key].createdAt {
				delete(mc.cache, key)
			}
		}
		mc.Unlock()
		<-time.After(time.Second)
	}
}
func (mc *MemoryCache) Set(key, value interface{}, ttl int64) error {
	defer mc.Unlock()
	mc.Lock()
	mc.cache[key] = &item{
		value:     value,
		ttl:       ttl,
		createdAt: time.Now().Unix(),
	}
	return nil
}
func (mc *MemoryCache) Get(key interface{}) (interface{}, error) {

	mc.RLock()
	val, ok := mc.cache[key]
	mc.RUnlock()
	if !ok {
		return nil, ErrorCacheMiss
	}
	return val, nil
}
