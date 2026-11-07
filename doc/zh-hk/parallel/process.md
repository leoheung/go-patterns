# 進程

`process` 套件管理作業系統子進程:拉起、監視退出、再停止——每個進程用調用者自選嘅 ID 標識。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/process"
```

## ProcessManager

`ProcessManager` 以你揀嘅 ID 追蹤一組子進程,多個 Goroutine 同時用都安全。

### 拉起進程

```go
m := process.NewProcessManager()

// dir 傳 nil = 管理者進程嘅工作目錄;env 傳 nil = 繼承我哋嘅環境變數
p, err := m.Start("recorder", `C:\apps\rec.exe`, []string{"serve"}, "", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(p.ID, p.Pid, p.StartedAt)
```

喺 Windows 上,子進程唔會自帶控制台窗口(CREATE_NO_WINDOW + 隱藏),啱背景常駐程式。

### 等待退出

```go
// 任意數量嘅 Goroutine 都可以:
<-p.Done()      // 進程退出(自己退定被 Stop)後關閉

err := p.Wait() // 阻塞到退出;返回 wait 錯誤(nil = 正常結束)
fmt.Println(p.Exited(), p.Err())
```

### 停止

```go
_ = m.Stop("recorder") // kill + 等待;冪等(未知或已退出嘅 ID = no-op)
```

### 查看狀態

```go
_ = m.Get("recorder") // *Process 或 nil
for _, p := range m.List() { ... } // 快照,按 ID 排序
fmt.Println(m.String())            // 每進程一行,最舊喺前
```

### 全部停止

```go
m.StopAll() // 並發停止所有被追蹤進程,逐一等待退出
```

## 注意

- ID 重複或仍被追蹤會令 `Start` 失敗;拉起失敗唔會留低註冊。
- 已退出嘅進程會留喺列表(狀態 `exited`),直到日後 `Start` 重用該 ID。
- `Stop` 只回報真實嘅 kill 錯誤:kill 一個已退出嘅進程係 no-op。
