# 漂亮打印

## 概述

`utils` 包提供了用於調試和日誌記錄的漂亮打印工具，使用 `spew` 庫進行全面的對象表示。

## API 參考

### `PPprettyPrint(v any)`

以漂亮的格式打印對象（類似 Python 的 `pprint`）。

```go
import "github.com/leoheung/go-patterns/utils"

type Person struct {
    Name string
    Age  int
}

utils.PrettyPrint(Person{Name: "Alice", Age: 30})
```

### `PrettyObjStr(v any) string`

返回對象的漂亮字符串表示而不打印（適用於日誌記錄）。

```go
str := utils.PrettyObjStr(data)
fmt.Println(str)
```

### `JSONalizeStr(v any) string`

將對象編碼為漂亮的 JSON 字符串。如果 JSON 編碼失敗，則回退到 `PrettyObjStr`。

```go
jsonStr := utils.JSONalizeStr(data)
fmt.Println(jsonStr)
```

### `DeJSONalizeStr(s string, v any) error`

將 JSON 字符串解碼到提供的對象中。目標必須是非 nil 指針。

```go
jsonStr := `{"Name": "Bob", "Age": 25}`
var person Person
err := utils.DeJSONalizeStr(jsonStr, &person)
if err != nil {
    log.Fatal(err)
}
```

## 示例

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
	// 示例 1: 漂亮打印
	utils.PrettyPrint(map[string]int{"a": 1, "b": 2, "c": 3})

	// 示例 2: 獲取漂亮字符串
	data := struct {
		Title string
		Items []string
	}{
		Title: "購物清單",
		Items: []string{"牛奶", "雞蛋", "麵包"},
	}
	str := utils.PrettyObjStr(data)
	fmt.Println("漂亮字符串:", str)

	// 示例 3: JSON 字符串
	config := Config{
		Server:   struct{ Host string; Port int }{Host: "localhost", Port: 8080},
		Database: struct{ Host string; Port int; Username string; Password string }{Host: "localhost", Port: 5432, Username: "user", Password: "pass"},
	}
	jsonStr := utils.JSONalizeStr(config)
	fmt.Println("JSON:", jsonStr)

	// 示例 4: 解碼 JSON
	jsonData := `{"Name": "Charlie", "Age": 35}`
	var person struct {
		Name string
		Age  int
	}
	err := utils.DeJSONalizeStr(jsonData, &person)
	if err != nil {
		fmt.Printf("錯誤: %v\n", err)
	} else {
		fmt.Printf("解碼結果: %+v\n", person)
	}
}
```

## 配置

漂亮打印函數使用以下設置：
- **縮進**: 2 個空格
- **禁用指針地址**: 啟用（更清晰的輸出）
- **禁用容量**: 啟用（更清晰的輸出）
- **排序鍵**: 啟用（確定的輸出）

## 注意事項

- `PPprettyPrint` 打印到標準輸出
- `PrettyObjStr` 和 `JSONalizeStr` 返回字符串以供日誌記錄/存儲
- `DeJSONalizeStr` 需要非 nil 指針作為目標
- 如果編碼失敗，JSON 會回退到漂亮打印