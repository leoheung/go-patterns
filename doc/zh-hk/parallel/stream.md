# Stream

建立喺 Go Channel 之上、可組合而且類型安全嘅流式 Pipeline。可以將 `Transform`、`Filter`、`Parallel` 等階段鏈接埋一齊,再用 `Collect` 或 `ForEach` 收結果。每個 Pipeline 都帶住一個全局退出(Done)Channel,成條鏈可以優雅咁關閉。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/stream"
```

## API 參考

### NewPipeline

用現有嘅輸入 Channel 建立 Pipeline。quit channel 係全局終止信號(通常係 `ctx.Done()`),會自動傳畀所有下游階段。

```go
// quit_ch: 全局退出 / Done Channel
// input_ch: 輸入 Channel(由調用者寫入同關閉)
p := stream.NewPipeline(quit_ch, input_ch)
```

### NewPipelineFromSlice

用 slice 建立 Pipeline。元素會推入內部嘅緩衝 Channel,而且 Channel 會自動關閉。

```go
p := stream.NewPipelineFromSlice(quit_ch, []int{1, 2, 3, 4, 5})
```

### Transform

類型轉換階段:`I → O`。返回新嘅 `*Pipeline[O]`,所以每階段嘅元素類型都可以改變。

```go
out := p.Transform(func(x int) string { return strconv.Itoa(x) })
```

### FlatMap

一對多展開:每個元素展開成 slice `[]I`,再將結果攤平放入輸出流。

```go
out := p.FlatMap(func(x int) []int { return []int{x, x * 10} })
```

### Filter

只保留滿足條件嘅元素。

```go
evens := p.Filter(func(x int) bool { return x%2 == 0 })
```

### Tap

副作用(日誌、metrics 等)。回調以 `any` 接收元素,元素原封不動傳落去。

```go
out := p.Tap(func(v any) { fmt.Println("see:", v) })
```

### Split

按條件將流分流成兩條獨立 Pipeline:`trueBranch` 收 predicate 為 true 嘅元素,`falseBranch` 收其餘。兩條分支都可以繼續鏈接。

```go
pass, fail := p.Split(func(v any) bool { return v.(int) >= 60 })
```

### Join

將本 Pipeline 同其他同類型 Pipeline 合併成一條(fan-in)。

```go
merged := p.Join(other1, other2)
```

### FanIn

Fan-in 嘅自由函數版本:將多條 Pipeline 合併成一條。同 `Join` 唔同嘅係佢冇接收者,quit channel 要顯式傳入。

```go
merged := stream.FanIn(quit_ch, p1, p2, p3)
```

### FanOut

爭搶模式:每個元素只會畀一個下游 Pipeline(負載均衡)。適合 N 個 Worker 分擔工作量。

```go
workers := p.FanOut(3) // 3 條 Pipeline 爭搶同一條流
```

### Broadcast

將每個元素複製畀所有下游 Pipeline。n 條 Pipeline 每條都會收到所有元素。

```go
copies := p.Broadcast(3) // 每個元素都會去晒 3 條 Pipeline
```

### Parallel

並行處理:n 個 Worker 同時爭搶輸入、並發執行 `fn`,結果自動合併成一條輸出 Pipeline。

```go
out := p.Parallel(4, func(x int) string { return heavyWork(x) })
```

### Take

只取前 n 個元素,之後自動結束。

```go
first5 := p.Take(5)
```

### Skip

跳過前 n 個元素,傳遞剩餘部分。

```go
rest := p.Skip(5)
```

### Buffer

插入一個輸出 Channel 帶指定緩衝大小嘅階段,令生產者同消費者速度解耦。

```go
buffered := p.Buffer(32)
```

### Run

終端操作:返回輸出 Channel,自己手動消費(`range`)。

```go
for v := range p.Run() {
    fmt.Println(v)
}
```

### Collect

終端操作:消費晒所有輸出,返回 slice。阻塞直到 Pipeline 結束。

```go
result := p.Collect()
```

### ForEach

終端操作:逐個元素執行回調。阻塞直到 Pipeline 結束。

```go
p.ForEach(func(v int) { fmt.Println(v) })
```

## Done Channel 生命週期管理

每條 Pipeline 都持有一個 quit channel(`<-chan struct{}`,通常係 `context.Context` 嘅 `Done()` Channel)。佢喺 `NewPipeline` / `NewPipelineFromSlice` 建立一次,然後傳畀所有下游階段:

- 每個階段各自跑一個 Goroutine,對 quit channel 同輸入 Channel 做 `select`。
- **正常完成**:上游 Channel 關閉後,階段會處理完在途元素,透過 `defer close(...)` 關閉自己嘅輸出 Channel,然後退出。關閉會沿鏈傳遞,所以終端操作(`Collect`、`ForEach`、`range`)會自然返回。
- **取消**:quit channel 一關閉,每個階段即刻返回並關閉自己嘅輸出 Channel,成條鏈安全解散,唔會有 Goroutine 洄漏。一般做法係建立 Pipeline 後寫 `defer cancel()`。
- `Parallel` 會喺 n 個 Worker 全部完成後先關閉輸出(`sync.WaitGroup`),所以取消前已產生嘅結果唔會被中途截斷。

簡單規則:生產者負責 `input_ch`(寫入同關閉佢),Pipeline 負責所有內部 Channel(每個階段關閉自己嘅輸出),調用者負責 quit channel(關閉佢嚟取消)。

## 範例:鏈式 Pipeline

```go
package main

import (
    "context"
    "fmt"
    "strconv"

    "github.com/leoheung/go-patterns/parallel/stream"
)

func main() {
    // 全局 Done Channel,用嚟取消
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Pipeline: 1..5 -> 平方 -> 保留 >= 10 -> int -> string
    result := stream.NewPipelineFromSlice(ctx.Done(), []int{1, 2, 3, 4, 5}).
        Transform(func(x int) int { return x * x }).    // 1 4 9 16 25
        Filter(func(x int) bool { return x >= 10 }).    // 16 25
        Transform(func(x int) string { return strconv.Itoa(x) }).
        Collect()

    fmt.Println(result) // 輸出: [16 25]
}
```

## 範例:Parallel + Take

```go
package main

import (
    "context"
    "fmt"

    "github.com/leoheung/go-patterns/parallel/stream"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 4 個 Worker 爭搶呢條流,取前 3 個結果
    out := stream.NewPipelineFromSlice(ctx.Done(), []int{1, 2, 3, 4, 5, 6, 7, 8}).
        Parallel(4, func(x int) int { return x * 10 }).
        Take(3).
        Collect()

    fmt.Println(out) // 輸出: [10 20 30 40 50 60 70 80] 其中任意 3 個
}
```

## 範例:FanOut、處理同 FanIn

```go
package main

import (
    "fmt"

    "github.com/leoheung/go-patterns/parallel/stream"
)

func main() {
    // FanIn 收嘅係雙向 quit channel(chan struct{}),
    // 所以呢度用普通 channel,唔用 ctx.Done()
    quit := make(chan struct{})
    defer close(quit)

    src := stream.NewPipelineFromSlice(quit, []int{1, 2, 3, 4, 5, 6})

    // FanOut: 3 條 Pipeline 爭搶同一條流(每個元素只畀其中一條)
    branches := src.FanOut(3)

    // 分別處理每個分支,再合併結果(fan-in)
    outs := make([]*stream.Pipeline[int], 0, len(branches))
    for _, b := range branches {
        outs = append(outs, b.Transform(func(x int) int { return x * x }))
    }

    merged := stream.FanIn(quit, outs...)
    fmt.Println(merged.Collect()) // 全部平方 1 4 9 16 25 36(次序可能唔同)
}
```

## 特性

- **可組合**:自由鏈接階段;每經過 `Transform` / `Parallel` 階段都可以改變元素類型
- **類型安全**:全部用泛型(`Pipeline[I any]`),唔使做類型斷言
- **並發**:每個階段自己跑一個 Goroutine;`Parallel` 再加多個爭搶 Worker
- **Fan-in / Fan-out / Broadcast**:`Join`、`FanIn`、`FanOut`、`Broadcast`、`Split` 組合各種流拓撲
- **可取消**:quit(Done)channel 令成條鏈優雅解散,唔會有 Goroutine 洄漏
- **終端操作**:`Run`、`Collect`、`ForEach`,加上 `Take` / `Skip` / `Buffer` 做流量控制
