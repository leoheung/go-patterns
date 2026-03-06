# WebSocket 管理器 (wrapsocket)

基於 `coder/websocket` 構建的高級 WebSocket 連接管理框架，提供連接生命週期管理、心跳檢測、分組廣播和元數據存儲等功能。

## 安裝

```go
import "github.com/leoheung/go-patterns/net/wrapsocket"
```

**依賴：**
- `github.com/coder/websocket` - WebSocket 實現

## 快速開始

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/leoheung/go-patterns/net/wrapsocket"
)

func main() {
    // 創建 Handler（使用默認選項）
    handler := wrapsocket.NewDefaultHandler(nil)

    // 設置連接回調
    handler.SetOnConnect(func(conn *wrapsocket.Conn) {
        fmt.Printf("客戶端已連接: %s\n", conn.ID)
    })

    // 設置消息回調
    handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
        fmt.Printf("收到來自 %s 的消息: %s\n", conn.ID, string(msg.Data))
    })

    // 設置斷開回調
    handler.SetOnDisconnect(func(conn *wrapsocket.Conn) {
        fmt.Printf("客戶端已斷開: %s\n", conn.ID)
    })

    http.ListenAndServe(":8080", handler)
}
```

## API 參考

### Handler

Handler 接口管理 WebSocket 升級和連接生命週期。

```go
// 使用自定義 WebSocket 接受選項創建 Handler
opts := &websocket.AcceptOptions{
    OriginPatterns: []string{"localhost", "*.example.com"},
}
handler := wrapsocket.NewDefaultHandler(opts)
```

#### 生命週期鉤子

```go
// 新客戶端連接時調用
handler.SetOnConnect(func(conn *wrapsocket.Conn) {
    // 連接已自動添加到管理器
    // 可在此發送歡迎消息或進行認證
})

// 客戶端斷開時調用
handler.SetOnDisconnect(func(conn *wrapsocket.Conn) {
    // 清理資源
})

// 收到消息時調用
handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
    // 處理收到的消息
})

// 發生錯誤時調用
handler.SetOnError(func(conn *wrapsocket.Conn, err error) {
    // 記錄或處理錯誤
})
```

#### 心跳配置

```go
config := &wrapsocket.HeartbeatConfig{
    Interval:  30 * time.Second,  // Ping 間隔
    Timeout:   10 * time.Second,  // Ping 超時
    MaxMissed: 3,                 // 最大丟失次數，超過則斷開
}
handler.SetHeartbeatConfig(config)
```

### Conn（連接）

每個 WebSocket 連接都被包裝在 `Conn` 結構中，帶有額外的元數據功能。

```go
// 連接屬性
type Conn struct {
    ID       string                 // 自動生成的 UUID
    Group    string                 // 可選的分組名稱
    metadata map[string]interface{} // 自定義元數據存儲
}
```

#### 元數據操作

```go
// 存儲自定義數據
conn.SetMetadata("user_id", "12345")
conn.SetMetadata("room", "lobby")

// 讀取數據
if val, ok := conn.GetMetadata("user_id"); ok {
    userID := val.(string)
}
```

#### 發送消息

```go
// 發送到特定連接
ctx := context.Background()
err := conn.Write(ctx, websocket.MessageText, []byte("hello"))

// 檢查連接狀態
if !conn.IsClosed() {
    // 可以安全發送
}
```

### ConnManager

管理所有活動連接，所有操作都是線程安全的。

```go
manager := handler.Manager()
```

#### 連接操作

```go
// 通過 ID 獲取連接
conn, ok := manager.Get("conn-uuid")

// 獲取所有連接
conns := manager.GetAll()

// 獲取連接數量
count := manager.Count()
```

#### 廣播

```go
// 廣播到所有連接
ctx := context.Background()
manager.Broadcast(ctx, websocket.MessageText, []byte("公告"))

// 發送到特定連接
manager.SendTo(ctx, "conn-uuid", websocket.MessageText, []byte("私信"))
```

#### 分組操作

```go
// 設置分組（在 OnConnect 回調中）
handler.SetOnConnect(func(conn *wrapsocket.Conn) {
    conn.Group = "room-1"
})

// 獲取分組內的連接
roomConns := manager.GetByGroup("room-1")

// 廣播到分組
manager.SendToGroup(ctx, "room-1", websocket.MessageText, []byte("房間消息"))
```

#### 清理

```go
// 關閉所有連接
manager.CloseAll(websocket.StatusGoingAway, "伺服器關閉")
```

### Message

從客戶端接收的消息結構。

```go
type Message struct {
    ID        string                // 連接 ID
    Type      websocket.MessageType // 消息類型（Text/Binary）
    Data      []byte                // 原始消息數據
    Timestamp time.Time             // 接收時間
}
```

## 完整範例

### 帶房間的聊天伺服器

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/coder/websocket"
    "github.com/leoheung/go-patterns/net/wrapsocket"
)

type ChatMessage struct {
    Room    string `json:"room"`
    Content string `json:"content"`
}

func main() {
    handler := wrapsocket.NewDefaultHandler(nil)
    
    // 配置心跳
    handler.SetHeartbeatConfig(&wrapsocket.HeartbeatConfig{
        Interval:  30 * time.Second,
        Timeout:   10 * time.Second,
        MaxMissed: 3,
    })
    
    manager := handler.Manager()
    
    handler.SetOnConnect(func(conn *wrapsocket.Conn) {
        fmt.Printf("用戶 %s 加入\n", conn.ID)
    })
    
    handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
        var chatMsg ChatMessage
        if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
            return
        }
        
        // 加入房間
        conn.Group = chatMsg.Room
        
        // 廣播到房間
        response, _ := json.Marshal(map[string]string{
            "from":    conn.ID,
            "content": chatMsg.Content,
        })
        manager.SendToGroup(context.Background(), chatMsg.Room, 
            websocket.MessageText, response)
    })
    
    handler.SetOnDisconnect(func(conn *wrapsocket.Conn) {
        fmt.Printf("用戶 %s 離開房間 %s\n", conn.ID, conn.Group)
    })
    
    http.ListenAndServe(":8080", handler)
}
```

### 帶認證的連接

```go
handler.SetOnConnect(func(conn *wrapsocket.Conn) {
    // 在元數據中存儲認證信息
    conn.SetMetadata("authenticated", false)
    conn.SetMetadata("auth_time", time.Now())
})

handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
    // 檢查認證狀態
    if auth, _ := conn.GetMetadata("authenticated"); !auth.(bool) {
        // 處理認證消息
        if isValidToken(string(msg.Data)) {
            conn.SetMetadata("authenticated", true)
            conn.Write(context.Background(), websocket.MessageText, 
                []byte(`{"type":"auth_success"}`))
        }
        return
    }
    
    // 處理已認證的消息
})
```

## 特性

- **線程安全**：所有連接和管理器操作都受互斥鎖保護
- **自動 UUID**：每個連接自動獲取唯一 ID
- **心跳檢測**：可配置的 Ping/Pong 機制，帶超時檢測
- **分組支持**：內建連接分組/房間功能
- **元數據**：可為連接附加自定義數據
- **生命週期鉤子**：連接、斷開、消息、錯誤回調
- **優雅關閉**：乾淨地關閉所有連接

## 注意事項

- Handler 會在連接建立時自動發送包含連接 ID 的歡迎消息
- 心跳是可選的；如未配置，則不會執行 Ping/Pong
- 所有回調都是同步執行的；長時間操作請使用 goroutine
- 管理器的 Broadcast 會自動跳過已關閉的連接
