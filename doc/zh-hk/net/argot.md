# 協議混淆 (argot)

`argot`（法語：*封閉群體內部使用、讓外人聽不懂的暗語*）把一份明文協議定義 —— 一串帶參數 schema 的函數呼叫 —— 編譯成一份 **手冊（manifest）**：線上格式中，所有自我描述的要素（函數名、欄位名、欄位順序、型別標記、穩定 opcode）都被替換成由共享密鑰推導出的值。

結果是：同時持有手冊的 client 與 server 能照常互相呼叫函數，而 **未** 持有手冊的觀察者只看到一串無法歸因的位元組 —— 沒有 magic number、沒有欄位名、沒有可解析的長度結構、沒有可與已知協議比對的重複指紋。

## 分工：Noise 與 argot

兩者解決不同問題，設計上搭配使用。

| 關注點 | 提供者 | 說明 |
|---|---|---|
| 機密性 | Noise（如 `Noise_NX` / `Noise_XX`） | 手冊與所有呼叫流量都走 Noise 通道 |
| 對端認證 | Noise 靜態密鑰 | 只有配對的對端能與 server 通話 |
| 完整性 / 防篡改 | Noise AEAD | 損壞在 argot 見到資料前就被偵測 |
| **移除協議語意** | **argot** | 線上無函數名、無欄位名、無穩定結構 |
| 抗指紋 | argot | 被動觀察者無法分類應用層協議 |
| 跨語言契約 | argot 手冊 | JSON 手冊 + 線上格式規範；Go 僅為參考實作 |

argot **不是** 加密方案。若移除 Noise 只靠 argot，流量仍可被讀取 —— 它只是沒有標籤。兩者互補：Noise 保護 *誰在說話、說了什麼*；argot 從視野中移除 *他們在說哪種協議*。

## 架構

```
         ┌─────────────┐    明文協議定義      ┌──────────────────────┐
         │  協議作者    │ ───────────────────> │ 雲端 Argot Service    │
         │  （你）      │ （函數 + schema）     │  Compile(spec, key)  │
         └─────────────┘                       └──────────┬───────────┘
                                                          │ 手冊（已簽名）
                                                          │ 經 Noise/HTTPS 分發
                                    ┌─────────────────────┴─────────────────────┐
                                    ▼                                           ▼
                          ┌──────────────────┐                       ┌──────────────────┐
                          │ 私密 client      │                       │ 私密 server      │
                          │ LoadArgot(...)   │                       │ LoadArgot(...)   │
                          └────────┬─────────┘                       └─────────┬────────┘
                                   │      Noise 傳輸 + argot 線上格式           │
                                   └─────────────────┬─────────────────────────┘
                                                     ▼
                                    觀察者看到：[opcode][slots][誘餌][填充]
                                    —— 無名稱、無標記、無 magic、無結構
```

只有雲端服務會同時見到明文協議定義 **與** 混淆密鑰。client 與 server 只持有手冊，永不接觸密鑰。

## 核心概念

| 概念 | 型別 | 生命週期 | 用途 |
|---|---|---|---|
| **協議定義（spec）** | `*Protocol` | 撰寫期 | 明文函數 + 參數 schema，本身可公開 |
| **編譯器（compiler）** | `*Compiler` | 雲端行程 | 把 spec 編譯成手冊；給定 `(key, spec)` 即確定性 |
| **手冊（manifest）** | `*Manifest` | 長期產物 | 解碼字典；語言無關 JSON；機密 |
| **編解碼器（codec）** | `*Argot` | 每連線 / 行程 | 手冊載入成兩張查找表，負責編解碼 |

### 為何 `Funcs` 是 map 而非 slice

```go
Funcs map[string]*Func // 函數名 → 定義
```

* **反查是 codec 的核心操作。** 解碼要把 opcode 映射回函數；編碼要把名稱映射到佈局。map 讓「名稱即身份」在 spec 層就顯式化，且直接對應非 Go 實作中的 JSON object / Python dict / JS object。
* **重複註冊在結構上不可能發生。** 用 slice 時兩個同名 `fs.read` 都會被接受，直到編譯期才衝突；用 map 時 key *就是* 唯一真源。
* **不再需要 `Func.Name` 欄位**，因為 key 已承載它。

確定性編譯不受影響：編譯器在推導任何佈局前先排序 map 的 key，迭代順序永不會滲入產物。

### 是確定性偽隨機，不是隨機

`Compile` 是確定性偽隨機：

* **同 `key` + 同 `spec` ⇒ 每次產出位元組完全相同的手冊。** 混淆值來自以 HMAC 密鑰派生播種的 PRNG，並在排序後的 key 上定序走訪。
* 意義：雲端服務可無狀態、水平擴充（任何副本都能重新產生同一份手冊），手冊可快取 / 比對 / diff，且可稽核。
* 輪換是 *換語言*，不是重擲骰子：換密鑰（或派生 per-epoch 密鑰）後重新編譯。

### 編譯實際混淆了什麼

| 要素 | 明文 | 編譯後 |
|---|---|---|
| 函數身份 | `"fs.read"` | 由 keyed shuffle 產生的 `uint16` opcode |
| 參數順序 | 宣告順序 | 每函數由 keyed PRNG 打亂 |
| 欄位身份 | `"path"`、`"offset"` | 線上完全沒有欄位名 —— 只有位置 |
| 型別標示 | 型別標記 | 不存在；解碼由手冊驅動 |
| 框架 | 訊息邊界 | 填充至 keyed 分桶大小 |
| 誘餌 | 無 | keyed 數量的假槽位穿插其中 |

opcode 由 **keyed shuffle** 產生，而非 `hash(name)`。雜湊可被「枚舉可能的函數名」字典反推；shuffle 既不揭露名稱空間，也不洩露註冊順序。

## 安裝

```go
import "github.com/leoheung/go-patterns/net/argot"
```

**依賴：** 除 Go 標準庫（`crypto/hmac`、`crypto/sha256`、`encoding/json`、`math/rand/v2`）外無其他依賴。

## 快速開始（雲端優先）

雲端永遠是第一步：先混淆協議，再分發手冊。

```go
package main

import (
    "crypto/ed25519"
    "fmt"
    "log"

    "github.com/leoheung/go-patterns/net/argot"
)

func main() {
    // 1. 協議作者的明文定義。
    proto := &argot.Protocol{
        Name: "mytools", Version: 1,
        Funcs: map[string]*argot.Func{
            "fs.read": {
                Params: []argot.Param{
                    {Name: "path", Type: argot.ValString},
                    {Name: "offset", Type: argot.ValInt},
                },
                Results: []argot.Param{{Name: "data", Type: argot.ValBytes}},
            },
        },
    }

    // 2. 混淆。密鑰只存在於雲端（生產環境為 KMS / HSM）。
    key := loadObfuscationKey()             // []byte，永不離開雲端
    manifest, err := argot.NewCompiler(key).Compile(proto)
    if err != nil {
        log.Fatal(err)
    }

    // 3. 序列化並簽名。
    payload, err := argot.MarshalManifest(manifest)
    if err != nil {
        log.Fatal(err)
    }
    sig := ed25519.Sign(signingKey, payload)

    fmt.Printf("手冊指紋 = %s\n", manifest.Fingerprint())
    serveSignedManifest(payload, sig) // GET /argot/v1/latest
}
```

## 完整範例：雲端優先的混淆流程

一次端到端走查：定義協議 → 在雲端混淆 → 分發手冊 → 在 Noise 之上跑一次呼叫。

### 步驟 0 —— 定義明文協議

這是名稱與 schema 唯一以可讀形式存在的地方。

```go
proto := &argot.Protocol{
    Name: "mytools", Version: 1,
    Funcs: map[string]*argot.Func{
        "fs.read": {
            Params: []argot.Param{
                {Name: "path", Type: argot.ValString},
                {Name: "offset", Type: argot.ValInt},
            },
            Results: []argot.Param{{Name: "data", Type: argot.ValBytes}},
        },
        "fs.write": {
            Params: []argot.Param{
                {Name: "path", Type: argot.ValString},
                {Name: "data", Type: argot.ValBytes},
            },
            Results: []argot.Param{{Name: "written", Type: argot.ValInt}},
        },
        "exec.run": {
            Params: []argot.Param{
                {Name: "cmd", Type: argot.ValString},
                {Name: "args", Type: argot.ValJSON},
            },
            Results: []argot.Param{
                {Name: "exitCode", Type: argot.ValInt},
                {Name: "output", Type: argot.ValString},
            },
        },
    },
}
```

### 步驟 1 —— 雲端：混淆協議

雲端服務是 spec 與密鑰唯一相遇之處。它的全部職責就是 `Compile` 加簽名。

```go
package main

import (
    "crypto/ed25519"
    "crypto/rand"
    "log"
    "net/http"
    "sync"

    "github.com/leoheung/go-patterns/net/argot"
)

type ArgotService struct {
    key       []byte            // 來自 KMS；永不記錄、永不回傳
    signer    ed25519.PrivateKey
    mu        sync.RWMutex
    manifest  []byte            // 快取的序列化手冊
    signature []byte
}

func NewArgotService(key []byte, signer ed25519.PrivateKey) *ArgotService {
    return &ArgotService{key: key, signer: signer}
}

// Publish 編譯 spec 並快取已簽名的手冊。
func (s *ArgotService) Publish(proto *argot.Protocol) error {
    man, err := argot.NewCompiler(s.key).Compile(proto)
    if err != nil {
        return err
    }
    payload, err := argot.MarshalManifest(man)
    if err != nil {
        return err
    }
    sig := ed25519.Sign(s.signer, payload)

    s.mu.Lock()
    s.manifest, s.signature = payload, sig
    s.mu.Unlock()
    return nil
}

// Handler 在加密的分發通道上提供已簽名的手冊。
func (s *ArgotService) Handler() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /argot/v1/latest", func(w http.ResponseWriter, r *http.Request) {
        s.mu.RLock()
        payload, sig := s.manifest, s.signature
        s.mu.RUnlock()
        if payload == nil {
            http.Error(w, "no manifest published", http.StatusServiceUnavailable)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("X-Argot-Signature", sigBase64(sig))
        w.Write(payload)
    })
    return mux
}

func main() {
    _, signer, _ := ed25519.GenerateKey(rand.Reader)

    svc := NewArgotService(loadObfuscationKey(), signer)
    if err := svc.Publish(mytoolsProtocol()); err != nil {
        log.Fatal(err)
    }
    log.Fatal(http.ListenAndServe(":8443", svc.Handler())) // 置於 Noise/HTTPS + mTLS 之後
}
```

**此步驟的要點**

* `Compile` 是確定性的，所以服務可以丟掉快取、按需重編譯並得到相同產物 —— 無需備份手冊。
* spec 永不抵達 client；密鑰永不離開雲端。client 只收到派生出的手冊。
* 分發前必須簽名。被篡改的手冊不是「混淆稍微壞掉」—— 而是攻擊者選定的線上語意。

### 步驟 2 —— 雙端：取得、驗證、載入

同一段程式碼同時跑在私密 client 與私密 server 上。

```go
func loadManifest(httpc *http.Client, verifyKey ed25519.PublicKey) (*argot.Argot, error) {
    // 1. 在加密通道內取得（Noise over TCP，或 HTTPS + 憑證固定）。
    resp, err := httpc.Get("https://argot.example.com/argot/v1/latest")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    payload, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // 2. 信任任何一個位元組之前先驗簽。
    sig, err := base64.StdEncoding.DecodeString(resp.Header.Get("X-Argot-Signature"))
    if err != nil {
        return nil, err
    }
    if !ed25519.Verify(verifyKey, payload, sig) {
        return nil, errors.New("argot: 手冊簽名不符")
    }

    // 3. 載入為 codec：兩張查找表，全程不涉及密鑰。
    a, err := argot.LoadArgot(payload)
    if err != nil {
        return nil, err
    }

    // 4. Noise 握手後，確認雙方說的是同一套黑話。
    if a.Fingerprint() != expectedFingerprint {
        return nil, errors.New("argot: 對端說的是另一套黑話")
    }
    return a, nil
}
```

`LoadArgot` 一趟建好兩個索引：`byName`（函數名 → `FuncLayout`，編碼用）與 `byOp`（opcode → 函數名，解碼用）。這兩張 map *就是* 整個 codec；非 Go 實作只要能從手冊建出同樣的兩個 dict 就已完備。

### 步驟 3 —— 在 Noise 之上的一次呼叫

Client 端：

```go
// 應用程式碼完全不變：按名稱呼叫、按名稱傳參。
payload, err := a.EncodeCall("fs.read", map[string]any{
    "path":   "/etc/hosts",
    "offset": 0,
})
if err != nil {
    return err
}
noiseConn.Write(payload) // 這一行之上不讀取名稱/順序/型別
```

Server 端：

```go
func serve(conn *noise.Conn, a *argot.Argot, reg map[string]Handler) error {
    buf := make([]byte, 64*1024)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            return err
        }

        // 解碼：opcode → 名稱，槽位 → 具名參數。
        name, args, err := a.DecodeCall(buf[:n])
        if err != nil {
            return err // 協議層錯誤：永不上線
        }

        h, ok := reg[name]
        if !ok {
            return fmt.Errorf("argot: 函數 %q 無處理器", name)
        }

        values, appErr := h(args)
        res := argot.Result{Values: values}
        if appErr != nil {
            res.Err = appErr.Error() // 應用層錯誤：會上線傳輸
        }

        reply, err := a.EncodeResult(name, res)
        if err != nil {
            return err
        }
        conn.Write(reply)
    }
}
```

Client 端解讀回覆：

```go
res, err := a.DecodeResult(reply, "fs.read") // 名稱用來選擇結果佈局
if err != nil {
    return err
}
if res.Err != "" {
    return errors.New(res.Err)
}
fmt.Printf("讀取 %d 位元組\n", len(res.Values["data"].([]byte)))
```

### 觀察者看到什麼

上例的線上位元組（數值僅為示意 —— 實際位元組取決於密鑰）：

```
bc 55  00  0a 2f 65 74 63 2f 68 6f 73 74 73  d6 9e 3c  ...
│      │   │  └────────── 誘餌槽位（解碼時丟棄）──────────┘
│      │   └──── offset=0，varint-zigzag → 0x00
│      └──────── （槽位 1：path）len16-prefixed → 0x000a 後接 10 個 UTF-8 位元組
└─────────────── opcode 0x55bc = 21948（keyed shuffle 的產物）
```

| 視角 | 可還原的資訊 |
|---|---|
| 觀察者，無手冊 | 什麼都沒有。無 magic、無欄位名、無型別標記。欄位邊界無法推導；填充讓長度失去意義。 |
| 觀察者，手冊洩露 | 全部 —— 手冊 *就是* 解碼字典。 |
| client / server，已載入手冊 | 函數身份、欄位身份、順序、型別，以及誘餌（丟棄）。 |

由此可見：**手冊是最核心的資產**。它的機密性就是整個混淆保證。這也是為何它要走 Noise/HTTPS 分發、要簽名，且永不作為 fallback 內嵌在二進位中。

## API 參考

### Spec 型別

```go
type ValueType int

const (
    ValString ValueType = iota // UTF-8 字串
    ValBytes                   // 原始 []byte
    ValInt                     // int64
    ValUint                    // uint64
    ValFloat                   // float64
    ValBool                    // bool
    ValJSON                    // 任意 JSON 值（以 any 透傳）
)

type Protocol struct {
    Name    string           // 協議名（僅元資訊，永不上線）
    Version uint32           // 雙端必須一致
    Funcs   map[string]*Func // 函數名 → 定義；key 即身份
}

type Func struct {
    Params  []Param // 宣告順序僅為可讀性；線上順序會被打亂
    Results []Param
}

type Param struct {
    Name string    // 僅編譯期使用；永不上線
    Type ValueType
}
```

### Compiler

```go
type Compiler struct{ /* key，永不外露 */ }

func NewCompiler(key []byte) *Compiler
func (c *Compiler) Compile(p *Protocol) (*Manifest, error)
```

`Compile` 只在雲端執行。它在推導佈局前先排序 spec 的 map key，然後輸出不含密鑰的手冊。

### Manifest

```go
type Manifest struct {
    ArgotVersion uint32                 // 格式版本；跨語言相容性的關卡
    ProtoName    string
    ProtoVersion uint32
    Funcs        map[string]*FuncLayout // 明文名 → 混淆佈局
    PadPolicy    PadPolicy
}

type FuncLayout struct {
    Opcode  uint16       // keyed shuffle 的產物；與名稱、順序無關
    Params  []SlotLayout // 已按線上順序排列
    Results []SlotLayout
    Decoys  []SlotLayout // 假槽位，解碼時丟棄
}

type SlotLayout struct {
    Ref      string  // 此槽位承載的 spec 參數（誘餌為生成名）
    Type     ValueType
    Encoding WireEnc // 編譯期選定
}

func MarshalManifest(m *Manifest) ([]byte, error)
func LoadManifest(data []byte) (*Manifest, error)
```

### Codec（Go 參考實作）

```go
type Argot struct{ /* manifest + byName + byOp */ }

func LoadArgot(manifestJSON []byte) (*Argot, error)
func (a *Argot) Manifest() *Manifest
func (a *Argot) Fingerprint() string // 手冊摘要，Noise 握手後交換

func (a *Argot) EncodeCall(name string, args map[string]any) ([]byte, error)
func (a *Argot) DecodeCall(data []byte) (name string, args map[string]any, err error)

type Result struct {
    Err    string         // 應用錯誤文字；空字串代表成功
    Values map[string]any // 依 Results schema 命名
}

func (a *Argot) EncodeResult(name string, r Result) ([]byte, error)
func (a *Argot) DecodeResult(data []byte, name string) (Result, error)
```

刻意分離的兩類錯誤：

* **協議層錯誤**（未知 opcode、位元組不足/多餘、編碼錯誤）以 Go `error` 在本地回傳，**永不上線** —— 上線會洩露是哪個 opcode 出錯。
* **應用層錯誤**（`Result.Err`）上線傳輸，因為對端需要它。

### Dispatcher（可選）

```go
type Handler func(args map[string]any) (map[string]any, error)

type Dispatcher struct{ /* argot + handlers */ }

func NewDispatcher(a *Argot) *Dispatcher
func (d *Dispatcher) Register(name string, h Handler)
func (d *Dispatcher) Handle(payload []byte) (reply []byte, err error)
```

`Handle` 是 server 端的膠水：解碼 → 分派 → 編碼。它本身不含任何混淆邏輯，可以省略；上方範例已用 inline 方式完成同樣工作。

## 手冊（交換格式）

手冊是唯一跨越語言邊界的產物，採 JSON 以便任何執行環境消費。

```json
{
  "argot_version": 1,
  "proto_name": "mytools",
  "proto_version": 1,
  "pad_policy": { "decoy_min": 0, "decoy_max": 2, "bucket": 64 },
  "funcs": {
    "fs.read": {
      "opcode": 48213,
      "params": [
        { "ref": "offset", "type": "int",    "encoding": "varint-zigzag" },
        { "ref": "path",   "type": "string", "encoding": "len16-prefixed" }
      ],
      "results": [
        { "ref": "data",   "type": "bytes",  "encoding": "len32-prefixed" }
      ],
      "decoys": [
        { "ref": "d0",     "type": "bytes",  "encoding": "len16-prefixed" }
      ]
    },
    "fs.write": {
      "opcode": 12077,
      "params": [
        { "ref": "data",   "type": "bytes",  "encoding": "len32-prefixed" },
        { "ref": "path",   "type": "string", "encoding": "len16-prefixed" }
      ],
      "results": [
        { "ref": "written","type": "int",    "encoding": "varint" }
      ],
      "decoys": []
    },
    "exec.run": {
      "opcode": 30861,
      "params": [
        { "ref": "args",   "type": "json",   "encoding": "len32-prefixed" },
        { "ref": "cmd",    "type": "string", "encoding": "len16-prefixed" }
      ],
      "results": [
        { "ref": "output", "type": "string", "encoding": "len16-prefixed" },
        { "ref": "exitCode","type": "int",   "encoding": "varint-zigzag" }
      ],
      "decoys": [
        { "ref": "d0",     "type": "bool",   "encoding": "fixed8" }
      ]
    }
  }
}
```

注意：`fs.read` 的參數是 `offset, path` —— 與 spec 宣告順序相反。`fs.write` 的結果用 `varint`，而 `fs.read` 的參數用 `varint-zigzag`，即使邏輯型別相同。沒有密鑰時，這些都無法從明文 spec 推導。

| 欄位 | 含義 |
|---|---|
| `argot_version` | 線上/手冊格式版本；不同值的對端無法互通。 |
| `proto_name`、`proto_version` | 回溯至撰寫期 spec 的可追蹤性；永不上線。 |
| `pad_policy.decoy_min/max` | 每訊息注入假槽位的數量範圍。 |
| `pad_policy.bucket` | 訊息長度向上取整至該值的倍數。 |
| `funcs.<name>.opcode` | 線上身份；於整份手冊內唯一。 |
| `params` / `results` | **陣列索引 = 線上位置。** 這就是全部的「順序」機制。 |
| `decoys` | 附加槽位；解碼後丟棄。 |

### 用其他語言實作 codec

不需要 Go 工具鏈、不需要密鑰、不需要編譯器。拿到手冊即可：

1. 建 `byName: dict[name → FuncLayout]` 與 `byOp: dict[opcode → name]`。
2. 編碼時查 `byName[name]`，以大端 u16 寫出 `opcode`，然後 **依陣列順序** 走 `params`，逐個按 `encoding` 與 `type` 編碼參數。
3. 寫出 `decoys` 槽位，內容任意（對端會丟棄；用隨機內容更好）。
4. 把訊息向上取整至 `pad_policy.bucket` 的倍數。
5. 解碼時讀 u16 opcode，經 `byOp` 解析，再按同一順序走同一批槽位，逐個綁定到 `ref`。

這就是全部契約。其餘都是應用程式碼。

## 線上格式規範

本節為規範性內容。遵循它的兩個實作可互通。

### 呼叫幀

```
+-----------+----------------+----------------+--------------+--------+
| opcode u16| slot[0]        | slot[1]        | 誘餌槽位      | 填充    |
| （大端）   |                |                | （0..n）      |        |
+-----------+----------------+----------------+--------------+--------+
```

* `opcode`：2 位元組大端，取值來自手冊的 `funcs.<name>.opcode`。
* 槽位：先依手冊陣列順序放 `len(params)` 個真實槽位，再放 `len(decoys)` 個誘餌槽位，最後填充。
* 填充：任意位元組，使總長達到 `pad_policy.bucket` 的下一個倍數。填充長度隱含（傳輸框架已知總長）。
* 無 magic number、無版本位元組、無欄位數量、無終止符。

### 結果幀

```
+-----------+---------+---------------------+--------------+--------+
| opcode u16| status  | （結果槽位    |      | 誘餌槽位      | 填充    |
|           | u8      |   或錯誤字串）       |              |        |
+-----------+---------+---------------------+--------------+--------+
```

* `opcode`：與請求相同的值，讓解碼端知道該套用哪個 `results` 佈局 —— 這也是 `DecodeResult` 需顯式傳入函數名的原因。
* `status`：`0x00` 成功、`0x01` 應用錯誤。其他值即協議錯誤。
* 成功時：依手冊順序寫 `len(results)` 個槽位。
* 錯誤時：單一 `len16-prefixed` UTF-8 字串（應用錯誤文字），然後填充；不寫結果佈局。
* status 位元組是整個協議中唯一的自我描述要素；它承載兩個位元的資訊，不承載任何協議身份。

### 編碼變體

`WireEnc` 名稱是手冊中使用的穩定識別符。

| 編碼 | 線上形式 | 適用 |
|---|---|---|
| `varint` | 無號 LEB128 | `ValUint`、非負 `ValInt` |
| `varint-zigzag` | zigzag 後無號 LEB128 | `ValInt`（緊湊處理負數） |
| `fixed64` | 8 位元組，IEEE-754 大端 | `ValFloat`；編譯器選定固定變體時亦可用於 `ValInt`/`ValUint` |
| `fixed8` | 1 位元組，`0x00` / `0x01` | `ValBool` |
| `len16-prefixed` | 2 位元組大端長度，後接負載 | `ValString`（UTF-8）、≤ 65535 的 `ValBytes` |
| `len32-prefixed` | 4 位元組大端長度，後接負載 | > 65535 的 `ValBytes`、`ValJSON` |

`ValJSON` 以緊湊 JSON 序列化後加長度前綴；它是「負載」型別而非自我描述型別 —— 型別資訊只存在於手冊中。

編譯器按 keyed PRNG 為每個槽位在適用變體中選擇，因此同一邏輯型別可能在不同函數中呈現不同形式，密鑰輪換後又會再變。

## 安全模型

**argot 做到的**

* 從線上移除函數名、欄位名、欄位順序、型別標記、magic number、穩定框架與穩定長度。
* 讓被動觀察者無法分類應用層協議、無法為其建立簽名。
* 讓兩套以不同密鑰編譯的協議流量彼此不可關聯。

**argot 做不到的**

* 它不加密。沒有 Noise 時位元組可讀。
* 它不認證對端。那是 Noise 的職責。
* 它不抵抗能迫使對端展露行為的主動探測者。
* 它不防端點淪陷：能讀 client 記憶體的人就能讀到手冊，混淆歸零。
* 流量分析只是被削弱而非被擊敗：填充緩解長度指紋，但不消除時序訊號。

**威脅與對策**

| 威脅 | 對策 |
|---|---|
| 被動 DPI 分類協議 | 無手冊的線上流量沒有可識別結構 |
| 藉穩定 opcode 做重放偵測 / 協議指紋 | opcode 隨密鑰而異；輪換即全部改變 |
| 分發路徑淪陷導致手冊被替換 | ed25519 簽名，載入前驗證 |
| 手冊洩露 | 混淆失效；以新密鑰重新 `Compile` 恢復。Noise 仍保護內容。 |
| 雲端服務淪陷 | spec 與密鑰暴露；重新輪換密鑰並重新發布。把 spec 撰寫側與簽名密鑰分離。 |

**函數名問題。** `funcs` 保留明文名作為 map key，便於閱讀與除錯。由於手冊本就只經認證加密通道分發、本身就是機密產物，這不增加暴露面：手冊無論如何都必須受保護。若你的威脅模型要求手冊在部分洩露下仍可承受，改用 `HMAC(key, name)` 作為 key —— 代價是失去人類可讀性。

**版本語意。** `argot_version` 管格式相容性；`proto_version` 管應用相容性。函數集或 schema 變更時遞增 `proto_version`；Noise 握手後的指紋比對會讓不匹配立即顯式失敗，而不是演變成連串解碼錯誤。

## 特性

* **結構上即語言無關** —— 手冊是 JSON，線上格式以位元組規範；Go 僅為參考實作
* **雲端優先** —— 只有服務同時持有 spec 與密鑰
* **確定性編譯** —— 無狀態服務、可重現的手冊、可比對的產物
* **線上無自我描述要素** —— 無名稱、無標記、無 magic、無穩定指紋
* **簽名分發** —— 載入前驗證 ed25519 簽名
* **每連線指紋校驗** —— 對端確認說的是同一套黑話
* **零第三方依賴** —— 僅標準庫
* **傳輸無關** —— 設計上置於 Noise 之下，但可跑在任何位元流傳輸之上

## 注意事項

* 編譯器只在雲端執行。client 與 server 永不會見到混淆密鑰，只見到派生出的手冊。
* 編譯是確定性的：同一 `(key, spec)` 產生同一手冊，所以服務無需持久化手冊。
* `Compile` 的輸出本身不是安全邊界；分發通道與簽名才是。
* 換密鑰會改變每個 opcode、每個參數順序與每個編碼選擇 —— 這就是輪換機制。
* 可選的 `Dispatcher` 不含任何混淆邏輯，僅為便利膠水。
* `argot` 與加密無關；請與 Noise 搭配使用，而非取代它。
