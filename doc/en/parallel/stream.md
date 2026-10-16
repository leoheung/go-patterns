# Stream

Composable, type-safe streaming pipelines built on Go channels. Chain stages like `Transform`, `Filter` and `Parallel` together, then consume the result with `Collect` or `ForEach`. Every pipeline carries a global quit (Done) channel so the whole chain can be shut down gracefully.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/stream"
```

## API Reference

### NewPipeline

Create a pipeline from an existing input channel. The quit channel is the global termination signal (typically `ctx.Done()`), and is propagated to every downstream stage automatically.

```go
// quit_ch: global quit / Done channel
// input_ch: input channel (the caller writes to and closes it)
p := stream.NewPipeline(quit_ch, input_ch)
```

### NewPipelineFromSlice

Create a pipeline from a slice. The items are pushed into an internal buffered channel and the channel is closed automatically.

```go
p := stream.NewPipelineFromSlice(quit_ch, []int{1, 2, 3, 4, 5})
```

### Transform

Type conversion stage: `I → O`. Returns a new `*Pipeline[O]` so the element type can change at every stage.

```go
out := p.Transform(func(x int) string { return strconv.Itoa(x) })
```

### FlatMap

One-to-many expansion: each element is expanded to a slice `[]I` and the results are flattened into the output stream.

```go
out := p.FlatMap(func(x int) []int { return []int{x, x * 10} })
```

### Filter

Keep only the elements that satisfy the predicate.

```go
evens := p.Filter(func(x int) bool { return x%2 == 0 })
```

### Tap

Side effect (logging, metrics, ...). The callback receives the element as `any` and the element is passed through unchanged.

```go
out := p.Tap(func(v any) { fmt.Println("see:", v) })
```

### Split

Split the stream into two independent pipelines by a predicate: `trueBranch` receives the elements where the predicate is true, `falseBranch` the rest. Each branch can be chained further.

```go
pass, fail := p.Split(func(v any) bool { return v.(int) >= 60 })
```

### Join

Merge the pipeline with other pipelines of the same element type into one (fan-in).

```go
merged := p.Join(other1, other2)
```

### FanIn

Free function version of fan-in: merge several pipelines into one. Unlike `Join`, there is no receiver; the quit channel is passed explicitly.

```go
merged := stream.FanIn(quit_ch, p1, p2, p3)
```

### FanOut

Competing-consumer mode: each element is delivered to exactly one downstream pipeline (load balancing). Good for parallel work shared among N workers.

```go
workers := p.FanOut(3) // 3 pipelines competing for the same stream
```

### Broadcast

Copy every element to all downstream pipelines. Each of the n pipelines receives every element.

```go
copies := p.Broadcast(3) // every element goes to all 3 pipelines
```

### Parallel

Parallel processing: n workers compete on the input, run `fn` concurrently, and the results are merged automatically into one output pipeline.

```go
out := p.Parallel(4, func(x int) string { return heavyWork(x) })
```

### Take

Take only the first n elements, then finish automatically.

```go
first5 := p.Take(5)
```

### Skip

Skip the first n elements and pass through the rest.

```go
rest := p.Skip(5)
```

### Buffer

Insert a stage whose output channel is buffered with the given size, decoupling producer and consumer speeds.

```go
buffered := p.Buffer(32)
```

### Run

Terminal operation: return the output channel for manual consumption (`range`).

```go
for v := range p.Run() {
    fmt.Println(v)
}
```

### Collect

Terminal operation: consume everything and return a slice. Blocks until the pipeline finishes.

```go
result := p.Collect()
```

### ForEach

Terminal operation: consume each element with a callback. Blocks until the pipeline finishes.

```go
p.ForEach(func(v int) { fmt.Println(v) })
```

## Done Channel Lifecycle

Every pipeline holds a quit channel (`<-chan struct{}`, typically a `context.Context`'s `Done()` channel). It is created once in `NewPipeline` / `NewPipelineFromSlice` and propagated to all downstream stages:

- Each stage runs in its own goroutine and does `select` on both the quit channel and its input channel.
- **Normal completion**: when the upstream channel is closed, the stage drains what is left in flight, closes its own output channel via `defer close(...)`, and exits. Closing propagates down the chain, so terminal operations (`Collect`, `ForEach`, `range`) return naturally.
- **Cancellation**: when the quit channel is closed, every stage returns immediately and closes its output channel, so the whole chain unwinds without leaks. Callers usually write `defer cancel()` after building the pipeline.
- `Parallel` closes its output after all n workers finish (`sync.WaitGroup`), so results produced before cancellation are never truncated mid-write.

Rule of thumb: the producer owns `input_ch` (writes to it and closes it), the pipeline owns all internal channels (each stage closes its own output), and the caller owns the quit channel (closes it to cancel).

## Example: Chained Pipeline

```go
package main

import (
    "context"
    "fmt"
    "strconv"

    "github.com/leoheung/go-patterns/parallel/stream"
)

func main() {
    // Global Done channel for cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Pipeline: 1..5 -> square -> keep >= 10 -> int -> string
    result := stream.NewPipelineFromSlice(ctx.Done(), []int{1, 2, 3, 4, 5}).
        Transform(func(x int) int { return x * x }).    // 1 4 9 16 25
        Filter(func(x int) bool { return x >= 10 }).    // 16 25
        Transform(func(x int) string { return strconv.Itoa(x) }).
        Collect()

    fmt.Println(result) // Output: [16 25]
}
```

## Example: Parallel + Take

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

    // 4 workers compete on the stream, take the first 3 results
    out := stream.NewPipelineFromSlice(ctx.Done(), []int{1, 2, 3, 4, 5, 6, 7, 8}).
        Parallel(4, func(x int) int { return x * 10 }).
        Take(3).
        Collect()

    fmt.Println(out) // Output: any 3 of [10 20 30 40 50 60 70 80]
}
```

## Example: FanOut, Processing and FanIn

```go
package main

import (
    "fmt"

    "github.com/leoheung/go-patterns/parallel/stream"
)

func main() {
    // FanIn takes a bidirectional quit channel (chan struct{}),
    // so use a plain channel here instead of ctx.Done()
    quit := make(chan struct{})
    defer close(quit)

    src := stream.NewPipelineFromSlice(quit, []int{1, 2, 3, 4, 5, 6})

    // FanOut: 3 pipelines compete for the same stream (each element goes to one of them)
    branches := src.FanOut(3)

    // Process each branch, then merge the results back (fan-in)
    outs := make([]*stream.Pipeline[int], 0, len(branches))
    for _, b := range branches {
        outs = append(outs, b.Transform(func(x int) int { return x * x }))
    }

    merged := stream.FanIn(quit, outs...)
    fmt.Println(merged.Collect()) // all squares 1 4 9 16 25 36 (order may vary)
}
```

## Features

- **Composable**: chain stages freely; the element type can change at each `Transform` / `Parallel` stage
- **Type-safe**: everything is generics (`Pipeline[I any]`), no type assertions needed
- **Concurrent**: each stage runs in its own goroutine; `Parallel` adds competing workers
- **Fan-in / Fan-out / Broadcast**: `Join`, `FanIn`, `FanOut`, `Broadcast`, `Split` for stream topologies
- **Cancellable**: the quit (Done) channel unwinds the whole chain gracefully without goroutine leaks
- **Terminal ops**: `Run`, `Collect`, `ForEach`, plus `Take` / `Skip` / `Buffer` for flow control
