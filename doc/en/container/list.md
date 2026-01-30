# List

A generic dynamic array implementation supporting Python list and JavaScript Array operations.

## Installation

```go
import "github.com/leoheung/go-patterns/container/list"
```

## Basic Operations

### Create a List

```go
// Create a new empty list
l := list.New[int]()

// Create a list from a slice
l := list.From([]int{1, 2, 3})

// Get length and capacity
length := l.Len()
capacity := l.Cap()

// Convert to slice
slice := l.ToSlice()

// Clone the list
clone := l.Clone()
```

## Element Access

```go
// Get element by index (supports negative indices)
elem := l.Get(0)      // First element
elem := l.Get(-1)     // Last element

// Set element by index
l.Set(0, 10)

// Safe element access
if elem, ok := l.At(0); ok {
    // Element exists
}

// Return a new List with element modified at index
l2 := l.With(0, 100)
```

## Adding and Removing Elements

```go
// Append elements to the end
l.Append(4, 5)
l.Push(6) // Alias for Append

// Extend with a slice
l.Extend([]int{7, 8})

// Add elements to the beginning
l.Unshift(0, -1)

// Insert at index (supports negative index)
l.Insert(1, 99)

// Remove and return the first element
if elem, ok := l.Shift(); ok {
    // Handle element
}

// Remove and return the last element
if elem, ok := l.Pop(); ok {
    // Handle element
}

// Remove the first occurrence of a value
l.RemoveFirst(5, func(a, b int) bool { return a == b })

// Remove element at index
if elem, ok := l.RemoveAt(2); ok {
    // Handle element
}

// Clear the list
l.Clear()
```

## Advanced Mutations (JS-like)

```go
// Splice: Remove/Replace elements
removed := l.Splice(start, deleteCount, items...)

// CopyWithin: Copy a range of elements within the list
l.CopyWithin(target, start, end)

// Fill: Fill a range with a value
l.Fill(value, start, end)
```

## Search and Query

```go
// Check if list contains an element
contains := l.Includes(5, func(a, b int) bool { return a == b })

// Find index of element
index := l.IndexOf(5, func(a, b int) bool { return a == b })
lastIndex := l.LastIndexOf(5, func(a, b int) bool { return a == b })

// Count occurrences
count := l.Count(5, func(a, b int) bool { return a == b })

// Find elements
if elem, ok := l.Find(func(v, i int) bool { return v > 10 }); ok { /* ... */ }
index := l.FindIndex(func(v, i int) bool { return v > 10 })

// Find from last
if elem, ok := l.FindLast(func(v, i int) bool { return v > 10 }); ok { /* ... */ }
lastIdx := l.FindLastIndex(func(v, i int) bool { return v > 10 })
```

## Transformation and Iteration

```go
// Read-only iteration
l.ForEach(func(v int, i int) { fmt.Println(v) })

// Concurrent iteration
l.ForEachAsync(ctx, maxGoroutines, func(v int, i int) { /* ... */ })

// Map elements to new list (Generic package function)
newList := list.Map(l, func(v int, i int) string { return fmt.Sprintf("%d", v) })

// Map to any (Method)
anyList := l.Map(func(v int, i int) any { return v * 2 })

// Concurrent Map
res, err := list.MapAsync(ctx, l, 4, func(v int, i int) int { return v * v })

// Filter elements
filtered := l.Filter(func(v int, i int) bool { return v > 5 })

// Reduce elements (Generic package function)
result := list.Reduce(l, 0, func(acc int, v int, i int) int { return acc + v })

// Every / Some
allMatch := l.Every(func(v int, i int) bool { return v > 0 })
anyMatch := l.Some(func(v int, i int) bool { return v > 100 })
```

## Sorting and Reversing

```go
// Sort in place
l.Sort(func(a, b int) bool { return a < b })

// Get sorted copy
lSorted := l.ToSorted(func(a, b int) bool { return a < b })

// Reverse in place
l.Reverse()

// Get reversed copy
lReversed := l.ToReversed()

// Join elements to string
str := l.Join(", ", func(v int) string { return fmt.Sprint(v) })
```

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/container/list"
)

func main() {
    // Create and populate list
    l := list.From([]int{3, 1, 4, 1, 5, 9, 2, 6})

    // Filter even numbers
    evens := l.Filter(func(v, i int) bool { return v%2 == 0 })
    fmt.Println("Even numbers:", evens.ToSlice())

    // Map to squares
    squares := list.Map(evens, func(v, i int) int { return v * v })
    fmt.Println("Squares:", squares.ToSlice())

    // Sort
    squares.Sort(func(a, b int) bool { return a < b })
    fmt.Println("Sorted:", squares.ToSlice())

    // Reduce to sum
    sum := list.Reduce(squares, 0, func(acc, v, i int) int { return acc + v })
    fmt.Println("Sum:", sum)
    
    // Join
    fmt.Println("Joined:", squares.Join(" | ", func(v int) string { return fmt.Sprint(v) }))
}
```

## Features

- **Generics**: Type-safe operations for any data type.
- **Python/JS Semantics**: Familiar methods like `Append`, `Pop`, `Splice`, `Map`, `Filter`.
- **Negative Indexing**: Supports `l.Get(-1)` for last element.
- **Concurrency Support**: `ForEachAsync` and `MapAsync` for parallel processing.
