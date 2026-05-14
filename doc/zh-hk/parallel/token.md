# Token

## 概述

Token 包提供了兩種基於令牌的同步原語，用於管理並發訪問：

1. **BoolToken** - 線程安全的布爾值包裝器
2. **StaticTokens** - 固定數量的令牌，用於限流

## BoolToken

一個簡單的線程安全的布爾值，可以原子地獲取和設置。

### API 參考

#### `NewBoolToken(value bool) *BoolToken`

使用給定的初始值創建一個新的 BoolToken。

#### `(bt *BoolToken) Get() bool`

以線程安全的方式返回當前的布爾值。

#### `(bt *BoolToken) Set(value bool)`

以線程安全的方式設置布爾值。

### 示例

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/parallel/token"
)

func main() {
	bt := token.NewBoolToken(true)

	// 讀取值
	fmt.Println("初始值:", bt.Get()) // true

	// 更新值
	bt.Set(false)
	fmt.Println("設置後:", bt.Get()) // false
}
```

## StaticTokens

一個固定大小的令牌桶，用於控制並發訪問或限流。

### API 參考

#### `NewStaticTokens(numTokens int) (*StaticTokens, error)`

創建一個具有指定令牌數量的新 StaticTokens。如果 `numTokens <= 0`，則返回錯誤。

#### `(st *StaticTokens) GrantNextToken() bool`

嘗試授予下一個令牌。如果成功授予令牌則返回 `true`，如果沒有可用令牌則返回 `false`。此操作是原子性的且線程安全。

### 示例

```go
package main

import (
	"fmt"
	"sync"
	"github.com/leoheung/go-patterns/parallel/token"
)

func main() {
	st, err := token.NewStaticTokens(3)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if st.GrantNextToken() {
				fmt.Printf("Worker %d: 獲取到令牌\n", id)
				// 執行工作...
			} else {
				fmt.Printf("Worker %d: 沒有可用令牌\n", id)
			}
		}(i)
	}
	wg.Wait()
}
```

## 注意事項

- `BoolToken` 可用於簡單的標誌位同步
- `StaticTokens` 可用於實現信號量樣式的行爲或簡單的限流
- 兩種類型都可以安全地用於並發場景，無需額外同步