# 並發

`parallel` 套件為 Go 提供並發模式及原語，幫助開發者編寫高效的並行程式。

## 模組

### [Barrier](./barrier.md)
允許多個 Goroutine 互相等待的同步原語。

特性：
- 循環屏障實現
- 支援條件變數
- 線程安全操作

### [Limiter](./limiter.md)
用於控制操作速率的速率限制器。

特性：
- 靜態速率限制
- 令牌桶演算法
- 可配置的時間間隔

### [Mutex](./mutex.md)
簡單的互斥鎖實現。

特性：
- 基本的 Lock/Unlock 操作
- 基於 Channel 實現
- 簡單高效

### [Pipeline](./pipeline.md)
用於數據處理的 Pipeline 模式。

特性：
- Fan-in 及 Fan-out 模式
- 廣播模式
- Take 操作
- 可組合的 Pipeline 階段

### [Worker Pool](./pool.md)
用於管理並發任務的 Worker Pool 模式。

特性：
- 固定大小的 Worker Pool
- 任務隊列
- 優雅關閉

### [PubSub](./pubsub.md)
發布-訂閱模式實現。

特性：
- 多個訂閱者
- 基於主題的消息傳遞
- 異步消息傳遞

### [Read-Write Lock](./rwlock.md)
支援多個讀取者或單個寫入者的讀寫鎖。

特性：
- 多個並發讀取者
- 獨佔寫入者存取
- 無飢餓實現

### [Semaphore](./semaphore.md)
用於限制資源並發存取的信號量。

特性：
- 可配置的許可數
- Acquire/Release 操作
- Channel 及條件變數實現
