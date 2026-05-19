/*
ShardedMap 分片锁 Map，不同 Key 可能不冲突
每个分片独立加锁，互不干扰
• 分片数量可配置，默认32个
• 分片哈希函数使用 FNV-1a，速度快，分布均匀
• 分片数据结构为 map[string]any，支持任意类型值
*/

package safemap

import (
	"hash/fnv"
	"sync"
)

type ShardedMap struct {
	shards    []*shard
	shardCount int
}

type shard struct {
	mu   sync.RWMutex
	data map[string]any
}

// New 创建分片 Map
func NewShardedMap(shardCount int) *ShardedMap {
	if shardCount <= 0 {
		shardCount = 32 // 默认32个分片
	}
	
	sm := &ShardedMap{
		shards:     make([]*shard, shardCount),
		shardCount: shardCount,
	}
	
	for i := 0; i < shardCount; i++ {
		sm.shards[i] = &shard{
			data: make(map[string]any),
		}
	}
	
	return sm
}

// getShard 根据 key 获取对应的分片
func (sm *ShardedMap) getShard(key string) *shard {
	// FNV 哈希算法，速度快，分布均匀
	h := fnv.New32a()
	h.Write([]byte(key))
	return sm.shards[h.Sum32()%uint32(sm.shardCount)]
}

// Get 读取
func (sm *ShardedMap) Get(key string) (any, bool) {
	s := sm.getShard(key)
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// Set 写入
func (sm *ShardedMap) Set(key string, value any) {
	s := sm.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete 删除
func (sm *ShardedMap) Delete(key string) {
	s := sm.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// ComputeIfAbsent 如果不存在则计算（原子操作）
func (sm *ShardedMap) ComputeIfAbsent(key string, compute func() any) any {
	s := sm.getShard(key)
	
	// 先读锁检查
	s.mu.RLock()
	if val, ok := s.data[key]; ok {
		s.mu.RUnlock()
		return val
	}
	s.mu.RUnlock()
	
	// 不存在，加写锁计算
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 双重检查，防止重复计算
	if val, ok := s.data[key]; ok {
		return val
	}
	
	val := compute()
	s.data[key] = val
	return val
}