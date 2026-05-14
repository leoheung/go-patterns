# Color Print

## Overview

The `utils` package provides color printing functionality for terminal output using ANSI escape codes.

## Color Types

```go
const (
    Red        // Red
    Green      // Green
    BrightBlue // Bright Blue
    Magenta    // Purple/Magenta
    Cyan       // Cyan
)
```

## API Reference

### `PrintlnColor(c color, text string)`

Prints text in the specified color and adds a newline. Falls back to plain text if the color is not found.

```go
import "github.com/leoheung/go-patterns/utils"

utils.PrintlnColor(utils.Red, "This is red text")
utils.PrintlnColor(utils.Green, "This is green text")
utils.PrintlnColor(utils.BrightBlue, "This is blue text")
utils.PrintlnColor(utils.Magenta, "This is purple text")
utils.PrintlnColor(utils.Cyan, "This is cyan text")
```

### `GetNextColor() color`

Returns the next color in rotation (Red → Green → BrightBlue → Magenta → Cyan → Red...).

```go
c1 := utils.GetNextColor() // Red
c2 := utils.GetNextColor() // Green
c3 := utils.GetNextColor() // BrightBlue
```

## Example

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/utils"
)

func main() {
	colors := []utils.color{
		utils.Red,
		utils.Green,
		utils.BrightBlue,
		utils.Magenta,
		utils.Cyan,
	}

	messages := []string{
		"Error message",
		"Success message",
		"Info message",
		"Warning message",
		"Debug message",
	}

	for i, msg := range messages {
		utils.PrintlnColor(colors[i], msg)
	}

	fmt.Println("\nUsing GetNextColor():")
	for i := 0; i < 3; i++ {
		c := utils.GetNextColor()
		utils.PrintlnColor(c, fmt.Sprintf("Message %d", i+1))
	}
}
```

## Notes

- Colors are rendered using ANSI escape codes
- Not all terminals support all colors
- Colors are automatically reset after each line
- `GetNextColor()` cycles through colors for easy alternating patterns