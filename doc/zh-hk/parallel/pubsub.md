# PubSub

發布-訂閱模式實現。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/pubsub"
```

## API 參考

### 建立 PubSub

```go
// 建立新的 PubSub 實例
ps := pubsub.NewPubSub()
```

### 訂閱

```go
// 訂閱主題
ch := ps.Subscribe("topic-name")
```

### 發布

```go
// 發布消息到主題
ps.Publish("topic-name", message)
```

### 取消訂閱

```go
// 取消訂閱主題
ps.Unsubscribe("topic-name", ch)
```

## 完整範例

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/parallel/pubsub"
)

func main() {
    ps := pubsub.NewPubSub()
    
    // 訂閱者 1
    ch1 := ps.Subscribe("news")
    go func() {
        for msg := range ch1 {
            fmt.Printf("訂閱者 1 收到: %v\n", msg)
        }
    }()
    
    // 訂閱者 2
    ch2 := ps.Subscribe("news")
    go func() {
        for msg := range ch2 {
            fmt.Printf("訂閱者 2 收到: %v\n", msg)
        }
    }()
    
    // 發布消息
    time.Sleep(100 * time.Millisecond)
    ps.Publish("news", "快訊: Go 1.22 發布了!")
    ps.Publish("news", "快訊: 新增了模式!")
    
    time.Sleep(100 * time.Millisecond)
}
```

## 輸出

```
訂閱者 1 收到: 快訊: Go 1.22 發布了!
訂閱者 2 收到: 快訊: Go 1.22 發布了!
訂閱者 1 收到: 快訊: 新增了模式!
訂閱者 2 收到: 快訊: 新增了模式!
```

## 特性

- **多個訂閱者**: 一對多消息分發
- **基於主題**: 按主題組織消息
- **異步傳遞**: 非阻塞發布
