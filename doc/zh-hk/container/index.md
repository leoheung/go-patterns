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
基於 Channel 的消息隊列實現，具備基本隊列操作。

特性：
- 基於 Channel 實現
- 支援 Context 取消
- 隊列生命週期管理
- 線程安全操作

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
