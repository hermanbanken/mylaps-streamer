package main

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Cache is a in-memory thread-safe cache
type Cache struct {
	lookup sync.Map
}

// CacheItem contains the locking mechanism for cache to prevent thundering herds
type CacheItem struct {
	value     interface{}
	computed  bool
	createdAt time.Time
	expiresAt time.Time
	lock      *sync.Mutex
}

type getter func() interface{}

// GetOrElse gets from the cache, or creates the entry once
func (c *Cache) GetOrElse(key string, getterFn getter) interface{} {
	return c.GetOrElseWithin(key, getterFn, 5*time.Second)
}

// GetOrElseWithin gets from the cache, or creates the entry once
func (c *Cache) GetOrElseWithin(key string, getterFn getter, expireIn time.Duration) interface{} {
	var lock = &sync.Mutex{}
	lock.Lock()
	defer lock.Unlock()

	fresh := CacheItem{nil, false, time.Time{}, time.Now().Add(expireIn), lock}
	i, loaded := c.lookup.LoadOrStore(key, &fresh)

	cacheItem, ok := i.(*CacheItem)
	if !ok {
		log.Warn("Failed to load CacheItem")
		return nil
	}

	if loaded {
		cacheItem.lock.Lock()
		defer cacheItem.lock.Unlock()
		if cacheItem.expiresAt.Before(time.Now()) {
			c.lookup.Delete(key)
			cacheItem.value = c.GetOrElseWithin(key, getterFn, expireIn)
		}
		log.Infof("Returning cached %v", cacheItem)
		return cacheItem.value
	}

	value := getterFn()
	cacheItem.createdAt = time.Now()
	cacheItem.value = value
	cacheItem.computed = true
	log.Infof("Cached %v", cacheItem)
	return value
}
