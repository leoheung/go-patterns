# ShardedMap

分片並發 Map（泛型），每片獨立鎖。

`ShardedMap[K, V]` 將 keyspace 切成固定數量嘅分片（shard），每片各有自己嘅 `sync.RWMutex`。唔同 key 嘅並發操作通常命中唔同分片，互相唔會阻塞——比單一 mutex 包住成個 `map[K]V` 嘅鎖競爭低好多，尤其係高並發存取嘅時候。

分片選擇用咗針對類型優化嘅 hash：`string` key 用 FNV-1a 64-bit，整數 key 直接用自身值，其他類型 fallback 到基於記憶體內容嘅 hash。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/safemap"
```

## API 參考

### 建立

```go
// 建立一個有 shardCount 個分片嘅分片 Map。
// 如果 shardCount <= 0，默認用 32。
func NewShardedMap[K comparable, V any](shardCount int) *ShardedMap[K, V]
```

### 基本操作

```go
// Get 返回 key 對應嘅值同埋佢係咪存在（讀鎖）。
func (sm *ShardedMap[K, V]) Get(key K) (V, bool)

// Set 將值存入 key（寫鎖）。
func (sm *ShardedMap[K, V]) Set(key K, value V)

// Delete 移除 key（寫鎖）。
func (sm *ShardedMap[K, V]) Delete(key K)

// GetOrStore 如果 key 已存在就返回舊值；否則存入 value 並返回佢。
// loaded 表示值係咪原本已存在（語義同 sync.Map）。
func (sm *ShardedMap[K, V]) GetOrStore(key K, value V) (actual V, loaded bool)
```

### Compute

```go
// Compute 喺分片寫鎖下原子性咁更新 key 嘅值：
//
//   - 如果 key 存在而且 computePresent != nil：
//     儲存 computePresent(舊值)；
//   - 否則（key 唔存在，或者 computePresent == nil）：
//     呼叫 computeAbsent() 並儲存結果（如果 key 存在但
//     computePresent == nil，會覆蓋舊值）；
//   - 如果第二種情況下 computeAbsent == nil，返回 error
//     （"invalid parameters"）。
func (sm *ShardedMap[K, V]) Compute(key K, computeAbsent func() V, computePresent func(V) V) error
```

## 完整範例

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/container/safemap"
)

func main() {
	// 建立一個有 16 個分片嘅 Map
	sm := safemap.NewShardedMap[string, int](16)

	// Set / Get
	sm.Set("alice", 1)
	sm.Set("bob", 2)

	if v, ok := sm.Get("alice"); ok {
		fmt.Println("alice:", v) // 1
	}

	// GetOrStore：如果已存在就保留舊值
	actual, loaded := sm.GetOrStore("alice", 100)
	fmt.Println(actual, loaded) // 1 true
	actual, loaded = sm.GetOrStore("carol", 3)
	fmt.Println(actual, loaded) // 3 false

	// Compute：原子性遞增計數器
	_ = sm.Compute("counter", func() int {
		return 1 // 唔存在：由 1 開始
	}, func(old int) int {
		return old + 1 // 已存在：遞增
	})
	v, _ := sm.Get("counter")
	fmt.Println("counter:", v) // 1

	_ = sm.Compute("counter", func() int { return 1 }, func(old int) int { return old + 1 })
	v, _ = sm.Get("counter")
	fmt.Println("counter:", v) // 2

	// Delete
	sm.Delete("bob")
	_, ok := sm.Get("bob")
	fmt.Println("bob 存在:", ok) // false
}
```

## 特性

- **分片鎖**：固定數量嘅分片，每片各自有 `RWMutex`——並發存取下鎖競爭低
- **支援泛型**：`K comparable, V any`，任何 key/value 類型都用得
- **類型優化 Hash**：`string` 用 FNV-1a，整數用自身值，其他類型用基於記憶體嘅 fallback
- **sync.Map 風格 API**：`Get` / `Set` / `Delete` / `GetOrStore` 語義熟悉易用
- **原子性 Compute**：喺一把寫鎖內完成讀-改-寫，唔使自己加鎖
