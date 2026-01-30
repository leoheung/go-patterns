# Pipeline

用於數據處理的 Pipeline 模式。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/pipeline"
```

## API 參考

### AddOnPipe

將數據從 X 轉換為 Y 的通用 Pipeline 節點。

```go
// q: 退出 Channel
// f: 轉換函數
// in: 輸入 Channel
out := pipeline.AddOnPipe(q, f, in)
```

### FanIn

將多個輸入 Channel 合併為單個輸出 Channel。

```go
out := pipeline.FanIn(q, input1, input2, input3)
```

### FanOut

將數據從單個輸入 Channel 分發到多個輸出 Channel。

```go
outs := pipeline.FanOut(q, in, 3) // 3 個輸出 Channel
```

### Broadcast

將數據廣播給多個訂閱者。

```go
broadcast := pipeline.NewBroadcast(q, in)
subscriber1 := broadcast.Subscribe()
subscriber2 := broadcast.Subscribe()
go broadcast.Run()
```

### Take

從輸入 Channel 取得前 n 個元素。

```go
out := pipeline.Take(q, 5, in) // 取得前 5 個元素
```

## 範例: 簡單 Pipeline

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/parallel/pipeline"
)

func main() {
    // 建立 Channel
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit)

    // 建立 Pipeline: Square -> Double
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
        fmt.Println(result) // 輸出: 2, 8, 18, 32, 50
    }
}
```

## 範例: FanOut 及 FanIn

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/parallel/pipeline"
)

func main() {
    // 建立 Channel
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit)

    // FanOut: 將數據分發給 3 個 Worker
    workers := pipeline.FanOut(quit, input, 3)

    // 並行處理數據
    process := func(x int) int { return x * 2 }
    var processed []chan int
    for _, worker := range workers {
        processed = append(processed, pipeline.AddOnPipe(quit, process, worker))
    }

    // FanIn: 合併所有 Worker 的結果
    output := pipeline.FanIn(quit, processed...)

    // 發送數據
    go func() {
        for i := 1; i <= 5; i++ {
            input <- i
        }
        close(input)
    }()

    // 接收結果
    for result := range output {
        fmt.Println(result)
    }
}
```

## 特性

- **可組合**: 鏈接多個 Pipeline 階段
- **類型安全**: 泛型函數
- **並發**: 利用 Go Channel 實現並發
- **可取消**: 退出 Channel 用於優雅關閉
