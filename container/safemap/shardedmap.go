package safemap

import (
	"sync"
	"unsafe"
)

type shard[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

type ShardedMap[K comparable, V any] struct {
	shards []*shard[K, V]
}

func NewShardedMap[K comparable, V any](shardCount int) *ShardedMap[K, V] {
	if shardCount <= 0 {
		shardCount = 32
	}
	sm := &ShardedMap[K, V]{
		shards: make([]*shard[K, V], shardCount),
	}
	for i := range sm.shards {
		sm.shards[i] = &shard[K, V]{data: make(map[K]V)}
	}
	return sm
}

func (sm *ShardedMap[K, V]) getShard(key K) *shard[K, V] {
	return sm.shards[hash(key)%uint64(len(sm.shards))]
}

// hash 根据类型选择最优哈希算法
func hash[K comparable](key K) uint64 {
	switch v := any(key).(type) {
	case string:
		return hashString(v)
	case int:
		return uint64(v)
	case int8:
		return uint64(v)
	case int16:
		return uint64(v)
	case int32:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case uintptr:
		return uint64(v)
	default:
		return hashAny(key)
	}
}

// FNV-1a 64bit 字符串哈希
func hashString(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

// 其他类型的 fallback 哈希
// 利用 unsafe 读取内存内容作为哈希值
func hashAny[K comparable](key K) uint64 {
	size := unsafe.Sizeof(key)
	switch size {
	case 0:
		return 0
	case 1:
		return uint64(*(*uint8)(unsafe.Pointer(&key)))
	case 2:
		return uint64(*(*uint16)(unsafe.Pointer(&key)))
	case 4:
		return uint64(*(*uint32)(unsafe.Pointer(&key)))
	case 8:
		return *(*uint64)(unsafe.Pointer(&key))
	default:
		return *(*uint64)(unsafe.Pointer(&key))
	}
}

func (sm *ShardedMap[K, V]) Get(key K) (V, bool) {
	s := sm.getShard(key)
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (sm *ShardedMap[K, V]) Set(key K, value V) {
	s := sm.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (sm *ShardedMap[K, V]) Delete(key K) {
	s := sm.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (sm *ShardedMap[K, V]) GetOrStore(key K, value V) (actual V, loaded bool) {
	s := sm.getShard(key)
	s.mu.RLock()
	if val, ok := s.data[key]; ok {
		s.mu.RUnlock()
		return val, true
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if val, ok := s.data[key]; ok {
		return val, true
	}
	s.data[key] = value
	return value, false
}

func (sm *ShardedMap[K, V]) ComputeIfAbsent(key K, compute func() V) V {
	s := sm.getShard(key)
	s.mu.RLock()
	if val, ok := s.data[key]; ok {
		s.mu.RUnlock()
		return val
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if val, ok := s.data[key]; ok {
		return val
	}
	val := compute()
	s.data[key] = val
	return val
}
