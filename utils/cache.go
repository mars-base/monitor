package utils

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var c *cache.Cache

// CacheCreate create cache, expireSecs: expire time, cleanSecs: clean time
func CacheCreate(expireSecs int, cleanSecs int) {
	if c == nil {
		c = cache.New(time.Duration(expireSecs)*time.Second, time.Duration(cleanSecs)*time.Second)
	}
}

// set key
// Add an item to the cache, replacing any existing item.
// If the duration is 0 (DefaultExpiration), the cache's default expiration time is used.
// If it is -1 (NoExpiration), the item never expires.
func CacheSetKey(key string, val interface{}, d time.Duration) {
	if c != nil {
		c.Set(key, val, d)
	}
}

// get key
func CacheGetKey(key string) (interface{}, bool) {
	if c != nil {
		return c.Get(key)
	}
	return nil, false
}
