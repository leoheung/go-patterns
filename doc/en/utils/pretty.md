# Pretty Print

## Overview

The `utils` package provides pretty printing utilities for debugging and logging, using the `spew` library for comprehensive object representation.

## API Reference

### `PPrettyPrint(v any)`

Prints an object in a pretty format (similar to Python's `pprint`).

```go
import "github.com/leoheung/go-patterns/utils"

type Person struct {
    Name string
    Age  int
}

utils.PrettyPrint(Person{Name: "Alice", Age: 30})
```

### `PrettyObjStr(v any) string`

Returns a pretty string representation of an object without printing (useful for logging).

```go
str := utils.PrettyObjStr(data)
fmt.Println(str)
```

### `JSONalizeStr(v any) string`

Encodes an object to a pretty JSON string. Falls back to `PrettyObjStr` if JSON encoding fails.

```go
jsonStr := utils.JSONalizeStr(data)
fmt.Println(jsonStr)
```

### `DeJSONalizeStr(s string, v any) error`

Decodes a JSON string into the provided object. The target must be a non-nil pointer.

```go
jsonStr := `{"Name": "Bob", "Age": 25}`
var person Person
err := utils.DeJSONalizeStr(jsonStr, &person)
if err != nil {
    log.Fatal(err)
}
```

## Example

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/utils"
)

type Config struct {
	Server struct {
		Host string
		Port int
	}
	Database struct {
		Host     string
		Port     int
		Username string
		Password string
	}
}

func main() {
	// Example 1: Pretty print
	utils.PPprettyPrint(map[string]int{"a": 1, "b": 2, "c": 3})

	// Example 2: Get pretty string
	data := struct {
		Title string
		Items []string
	}{
		Title: "Shopping List",
		Items: []string{"Milk", "Eggs", "Bread"},
	}
	str := utils.PrettyObjStr(data)
	fmt.Println("Pretty string:", str)

	// Example 3: JSON string
	config := Config{
		Server:   struct{ Host string; Port int }{Host: "localhost", Port: 8080},
		Database: struct{ Host string; Port int; Username string; Password string }{Host: "localhost", Port: 5432, Username: "user", Password: "pass"},
	}
	jsonStr := utils.JSONalizeStr(config)
	fmt.Println("JSON:", jsonStr)

	// Example 4: Decode JSON
	jsonData := `{"Name": "Charlie", "Age": 35}`
	var person struct {
		Name string
		Age  int
	}
	err := utils.DeJSONalizeStr(jsonData, &person)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Decoded: %+v\n", person)
	}
}
```

## Configuration

The pretty print functions use these settings:
- **Indent**: 2 spaces
- **Disable Pointer Addresses**: Enabled (cleaner output)
- **Disable Capacities**: Enabled (cleaner output)
- **Sort Keys**: Enabled (deterministic output)

## Notes

- `PPprettyPrint` prints to stdout
- `PrettyObjStr` and `JSONalizeStr` return strings for logging/storage
- `DeJSONalizeStr` requires a non-nil pointer as the target
- JSON fallback to pretty print if encoding fails