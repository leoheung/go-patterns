package cache

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/leoheung/go-patterns/container/pq"
	"github.com/leoheung/go-patterns/utils"
)

type Cache struct {
	buffer  map[string]*CacheItem
	manager *pq.PriorityScheduledTaskManager
	mu      sync.RWMutex
}

type CacheItem struct {
	data            any
	cachingDuration time.Duration
	cancelDelete    *pq.Cancelable
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

	var old *CacheItem
	var ok bool

	if old, ok = c.buffer[key]; ok {
		old.cancelDelete.TryCancel()
	}

	cancel, err := c.manager.PendNewTask(func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.buffer, key)

	}, time.Now().Add(cachingDuration))

	if err != nil {
		if old != nil{
			old.cancelDelete.TryRecover()
		}
		return fmt.Errorf("failed to arrange caching expiration: %s", err.Error())
	}

	c.buffer[key] = &CacheItem{
		data:            data,
		cachingDuration: cachingDuration,
		cancelDelete:    cancel,
	}

	return nil
}

func (c *Cache) Get(key string) any {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.buffer[key]
	if ok {
		if item.cancelDelete.TryCancel() {
			utils.DevLogSuccess(fmt.Sprintf("[成功]cancel %s 的expire", key))
		} else {
			utils.DevLogError(fmt.Sprintf("[失敗]cancel %s 的expire", key))
		}

		newCancel, err := c.manager.PendNewTask(func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			delete(c.buffer, key)

		}, time.Now().Add(item.cachingDuration))
		if err != nil {
			item.cancelDelete.TryRecover()
			utils.DevLogError(fmt.Sprintf("[失敗]安排 %s 的新expire", key))
		} else {
			utils.DevLogSuccess(fmt.Sprintf("[成功]安排 %s 的新expire", key))
			item.cancelDelete = newCancel
		}
		return item.data
	} else {
		return nil
	}
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if data, ok := c.buffer[key]; ok {
		data.cancelDelete.TryCancel()
		delete(c.buffer, key)
	}
}

func (c *Cache) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var ret strings.Builder
	fmt.Fprintf(&ret, "total %d cache items\n", len(c.buffer))
	for k := range c.buffer {
		ret.WriteString(k)
		ret.WriteString(",")
	}
	ret.WriteString("\n")

	ret.WriteString(c.manager.String())
	return ret.String()
}
