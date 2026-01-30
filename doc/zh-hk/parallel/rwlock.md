# 讀寫鎖

支援多個讀取者或單個寫入者的讀寫鎖實現。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/rwlock"
```

## API 參考

### 建立讀寫鎖

```go
// 建立新的讀寫鎖
rw := rwlock.NewRWLock()
```

### 讀取鎖

```go
// 取得讀取鎖
rw.RLock()
defer rw.RUnlock()
```

### 寫入鎖

```go
// 取得寫入鎖
rw.Lock()
defer rw.Unlock()
```

## 完整範例

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/leoheung/go-patterns/parallel/rwlock"
)

func main() {
    rw := rwlock.NewRWLock()
    data := make(map[string]string)
    var wg sync.WaitGroup
    
    // 寫入者
    for i := 0; i < 2; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            rw.Lock()
            defer rw.Unlock()
            
            key := fmt.Sprintf("key-%d", id)
            data[key] = fmt.Sprintf("value-%d", id)
            fmt.Printf("寫入者 %d 寫入 %s\n", id, key)
            time.Sleep(100 * time.Millisecond)
        }(i)
    }
    
    // 讀取者
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            rw.RLock()
            defer rw.RUnlock()
            
            fmt.Printf("讀取者 %d 讀取, 數據數量: %d\n", id, len(data))
            time.Sleep(50 * time.Millisecond)
        }(i)
    }
    
    wg.Wait()
}
```

## 輸出

```
寫入者 0 寫入 key-0
讀取者 0 讀取, 數據數量: 1
讀取者 1 讀取, 數據數量: 1
寫入者 1 寫入 key-1
讀取者 2 讀取, 數據數量: 2
...
```

## 特性

- **多個讀取者**: 並發讀取存取
- **獨佔寫入者**: 同一時間只有一個寫入者
- **無讀取者飢餓**: 公平排程
