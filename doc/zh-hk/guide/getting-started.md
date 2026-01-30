# 快速開始

## 介紹

`go-patterns` 是一系列以 Go 語言實現的並發模式與數據結構，旨在幫助開發者更好地理解與運用 Go 的並發特性。

## 安裝

```bash
go get github.com/leoxiang66/go-patterns
```

## 快速上手

### 使用 List

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/container/list"
)

func main() {
    // 建立新列表
    l := list.New[int]()
    
    // 新增元素
    l.Append(1, 2, 3)
    l.Push(4)
    
    // 取得元素
    elem := l.Get(0)
    fmt.Println(elem) // 輸出: 1
    
    // 迭代
    for i := 0; i < l.Len(); i++ {
        fmt.Println(l.Get(i))
    }
}
```

### 使用 Pipeline

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/parallel/pipeline"
)

func main() {
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit)
    
    // 建立 Pipeline 階段
    square := func(x int) int { return x * x }
    double := func(x int) int { return x * 2 }
    
    stage1 := pipeline.AddOnPipe(quit, square, input)
    stage2 := pipeline.AddOnPipe(quit, double, stage1)
    
    // 發送數據
    go func() {
        for i := 1; i <= 5; i++ {
            input <- i
        }
        close(input)
    }()
    
    // 接收結果
    for result := range stage2 {
        fmt.Println(result)
    }
}
```

### 使用 Semaphore

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/parallel/semaphore"
)

func main() {
    // 建立容量為 3 的 Semaphore
    sem := semaphore.NewSemaphore(3)
    
    // 取得及釋放
    sem.Acquire()
    defer sem.Release()
    
    // 執行操作
    fmt.Println("執行操作中...")
}
```

## 專案結構

```
go-patterns/
├── container/      # 數據結構
│   ├── list/       # 通用動態陣列
│   ├── msgQueue/   # 基於 Channel 的消息隊列
│   ├── pq/         # 優先隊列
│   └── cache/      # 快取實現
├── parallel/       # 並發模式
│   ├── barrier/    # 同步屏障
│   ├── limiter/    # 速率限制器
│   ├── mutex/      # 互斥鎖
│   ├── pipeline/   # Pipeline 模式
│   ├── pool/       # Worker Pool
│   ├── pubsub/     # 發布/訂閱模式
│   ├── rwlock/     # 讀寫鎖
│   └── semaphore/  # 信號量
├── utils/          # 工具函數
├── cryptography/   # 加密工具
└── net/            # 網絡工具
```

## 下一步

- 探索 [Container](/zh-hk/container/) 數據結構
- 了解 [Parallel](/zh-hk/parallel/) 並發模式
- 查看 [Utils](/zh-hk/utils/) 輔助函數
