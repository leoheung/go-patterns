# 容器

`container` 套件為 Go 提供通用數據結構，設計為高效且易於使用。

## 模組

### [List](./list.md)
通用動態陣列實現，支援 Python list 及 JavaScript Array 操作。

特性：
- 支援泛型
- 動態調整大小
- 支援負數索引
- 豐富的 API（Append、Push、Pop、Shift、Unshift 等）
- 函數式操作（Map、Filter、Reduce）

### [Message Queue](./msgqueue.md)
輕量級基於 Topic 嘅消息隊列(`MQLite`),透過 HTTP 回調將消息推送畀訂閱者。實現位於 `parallel/communication/mq`。

特性：
- 基於 Topic 嘅 HTTP 回調式發布/訂閱
- 非阻塞發布,支援自動重試
- Root Context 生命週期管理

### [Priority Queue](./pq.md)
通用優先隊列實現，支援自定義優先級比較。

特性：
- 支援泛型
- 自定義比較函數
- 二元堆積實現
- 高效的入隊/出隊操作

### [Cache](./cache.md)
用於儲存及檢索數據的快取實現。

特性：
- 鍵值儲存
- 支援 TTL
- 線程安全操作

### [Renewable Context](./context.md)
可復活嘅 `context.Context` 實現,內置泛型數據載體。

特性：
- 取消同重新激活(復活)Context
- 復活事件訂閱
- 泛型嘅數據讀寫(Get/Set Data)

### [Circular Stack](./circular.md)
固定容量環形棧,支援訂閱通知。

特性：
- 固定容量,覆蓋最舊條目
- Push 時嘅訂閱通知
- 支援泛型

### [Subscribe](./subscribe.md)
可觀察值提供者,支援訂閱者通知。

特性：
- 可觀察值 + 變更通知
- 訂閱/取消訂閱觀察者
- 支援泛型

### [Safe Map](./safemap.md)
分片並發 Map。

特性：
- 分片鎖,高並發
- Get/Set/Compute 操作
- 支援泛型鍵值

### [Safe Slice](./safeslice.md)
並發安全嘅 Slice。

特性：
- 線程安全嘅 append/read/remove
- 條件移除(RemoveIf)
- 支援泛型

### [Skip List](./skiplist.md)
機率式跳表,用於有序鍵值儲存。

特性：
- 期望 O(log n) 插入/查找/刪除
- 有序遍歷
- 並發友好設計

### [Tree](./tree.md)
樹狀結構:treap(bst)、dbtree 同二元堆積(heap)。

特性：
- bst — Treap(隨機化二元搜尋樹)
- dbtree — 距離/方位樹
- heap — 泛型二元堆積
