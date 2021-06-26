package cache

import (
	"sync"
	"time"
)

/**
 * This Cache behaves like Redis, where each individual key have its time to live
 */

type RedisCacheItem struct {
	value interface{}

	// ttl in seconds
	ttl int64
}
type RedisCache struct {
	Lock      sync.RWMutex
	checkTime time.Duration
	Item      map[string]RedisCacheItem
}

func NewRedisCache(d time.Duration) *RedisCache {
	rc := &RedisCache{
		Item:      make(map[string]RedisCacheItem),
		checkTime: d,
	}

	go func() {
		for {
			time.Sleep(d)
			rc.Lock.Lock()
			for k, v := range rc.Item {
				now := time.Now()
				ttl := time.Unix(v.ttl, 0)
				elapsed := now.Sub(ttl)
				if elapsed > 0 {
					delete(rc.Item, k)
				}
			}
			rc.Lock.Unlock()
		}
	}()

	return rc
}

func (rc *RedisCache) Put(key string, item interface{}, ttl int64) {

	rci := RedisCacheItem{
		value: item,
		ttl:   ttl,
	}
	rc.Lock.Lock()
	rc.Item[key] = rci
	rc.Lock.Unlock()
}

func (rc *RedisCache) Get(key string) (RedisCacheItem, bool) {
	rc.Lock.Lock()
	v, ok := rc.Item[key]
	rc.Lock.Unlock()
	return v, ok
}

func (rc *RedisCache) Remove(key string) {
	rc.Lock.Lock()
	delete(rc.Item, key)
	rc.Lock.Unlock()
}
