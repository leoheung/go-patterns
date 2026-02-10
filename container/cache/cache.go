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
	cachingDuration *time.Duration
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

func (c *Cache) Add(key string, data any, cachingDuration *time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var old *CacheItem
	var ok bool

	if old, ok = c.buffer[key]; ok {
		// 如果舊對象有過期任務，則取消它
		if old.cancelDelete != nil {
			old.cancelDelete.TryCancel()
		}
	}

	var cancel *pq.Cancelable
	var err error

	// 只有當 cachingDuration 不為 nil 時才安排過期任務
	if cachingDuration != nil {
		// 先創建 item 以便在閉包中捕獲指針
		item := &CacheItem{
			data:            data,
			cachingDuration: cachingDuration,
		}

		cancel, err = c.manager.PendNewTask(func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			// 雙重保險：即使 PTM 錯過了取消訊號，這裡也能防止誤刪
			if current, ok := c.buffer[key]; ok && current == item {
				delete(c.buffer, key)
			}
		}, time.Now().Add(*cachingDuration))

		if err != nil {
			// 如果安排失敗且舊任務存在，嘗試恢復舊任務（Best Effort）
			if old != nil && old.cancelDelete != nil {
				old.cancelDelete.TryRecover()
			}
			return fmt.Errorf("failed to arrange caching expiration: %s", err.Error())
		}
		item.cancelDelete = cancel
		c.buffer[key] = item
	} else {
		// 永久緩存
		c.buffer[key] = &CacheItem{
			data:            data,
			cachingDuration: nil,
			cancelDelete:    nil,
		}
	}

	return nil
}

func (c *Cache) Get(key string) any {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.buffer[key]
	if ok {
		// 如果是永久緩存 (cachingDuration == nil)，直接返回數據，不執行 Renew
		if item.cachingDuration == nil {
			return item.data
		}

		if item.cancelDelete != nil {
			if item.cancelDelete.TryCancel() {
				utils.DevLogSuccess(fmt.Sprintf("[成功]cancel %s 的expire", key))
			} else {
				// Cancel 失敗通常意味著已經被 Cancel 過，不影響我們繼續 Renew
				utils.DevLogInfo(fmt.Sprintf("[注意]cancel %s 失敗(可能已過期/取消), 繼續安排新expire", key))
			}
		}

		newCancel, err := c.manager.PendNewTask(func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			// 同樣加入身份檢查
			if current, ok := c.buffer[key]; ok && current == item {
				delete(c.buffer, key)
			}
		}, time.Now().Add(*item.cachingDuration))

		if err != nil {
			// 如果新任務安排失敗，嘗試恢復舊任務（Best Effort）
			if item.cancelDelete != nil {
				item.cancelDelete.TryRecover()
			}
			utils.DevLogError(fmt.Sprintf("[失敗]安排 %s 的新expire: %v", key, err))
		} else {
			utils.DevLogSuccess(fmt.Sprintf("[成功]安排 %s 的新expire", key))
			item.cancelDelete = newCancel
		}
		return item.data
	}
	return nil
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if data, ok := c.buffer[key]; ok {
		if data.cancelDelete != nil {
			data.cancelDelete.TryCancel()
		}
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
