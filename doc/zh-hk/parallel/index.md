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

### [Pools](./pool.md)
Worker Pool、支援 Context 嘅異步任務池同通用對象池。

特性：
- WorkerPool - 基於信號量嘅固定大小工作池
- AsyncPoolV2 - 支援 Context 嘅異步池,內置有界任務隊列同統計
- ObjPool - 通用對象池,避免頻繁分配同 GC
- [AsyncPool](./asyncpool.md) - 帶緩衝區支持的異步任務池

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

### [Token](./token.md)
基於令牌的同步原語。

特性：
- BoolToken - 線程安全的布爾值
- StaticTokens - 固定數量的令牌用於限流

### [Stream](./stream.md)
可組合嘅流式數據處理 Pipeline。

特性：
- 可鏈式階段(Transform、Filter、FlatMap、Parallel、Buffer 等)
- Fan-in / Fan-out / Broadcast
- 終端操作(Collect、ForEach、Take、Skip)
- 基於 Done channel 嘅生命週期管理

### [Communication](./communication.md)
消息傳遞同時間抽象嘅接口定義。

特性：
- Message / MessageHeader 結構
- MQ 接口(Publisher、Subscriber、Broadcaster、HttpClient)
- Clock 接口
