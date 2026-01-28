package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/leoheung/go-patterns/container/pq"
)

type Cache struct {
	buffer  map[string]*CacheItem
	manager *pq.PriorityScheduledTaskManager
	mu      sync.RWMutex
}

type CacheItem struct {
	data            any
	cachingDuration time.Duration
}

func NewCache() (*Cache, error) {
	m, err := pq.NewPriorityScheduledTaskManager()
	if err != nil {
		return nil, err
	}

	cache := &Cache{
		buffer:  make(map[string]*CacheItem),
		manager: m,
		mu:      sync.RWMutex{},
	}
	return cache, nil
}

func (c *Cache) Add(key string, data any, cachingDuration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.buffer[key]; ok {
		return fmt.Errorf("failed to push the data with key: %s: the key already exists, please use another key", key)
	}

	err := c.manager.PendNewTask(func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.buffer, key)

	}, time.Now().Add(cachingDuration))

	if err != nil {
		return fmt.Errorf("failed to arrange caching expiration: %s", err.Error())
	}

	c.buffer[key] = &CacheItem{
		data:            data,
		cachingDuration: cachingDuration,
	}

	return nil
}

func (c *Cache) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.buffer[key]
	if ok {
		return item.data
	} else {
		return nil
	}
}


func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.buffer, key)
}