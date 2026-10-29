# Protocol Obfuscation (argot)

`argot` (French: *a private language used by a closed group so outsiders cannot understand it*) compiles a plain protocol definition — a list of function calls with parameter schemas — into a **manifest**: a wire layout in which every self-describing element (function name, field name, field order, type tag, stable opcode) has been replaced by a value derived from a shared secret.

The result is that a client and a server which both hold the manifest can call each other's functions normally, while an observer that does **not** hold the manifest sees a stream of unattributable bytes: no magic number, no field names, no length-delimited structure it can parse, no repeatable fingerprint it can match against a known protocol.

## Division of labour: Noise vs argot

These two layers solve different problems and are meant to be used together.

| Concern | Provided by | Notes |
|---|---|---|
| Confidentiality | Noise (e.g. `Noise_NX` / `Noise_XX`) | The manifest and all call traffic travel inside the Noise transport |
| Peer authentication | Noise static keys | Only the paired peer can talk to the server at all |
| Integrity / anti-tamper | Noise AEAD | Corruption is detected before argot sees the bytes |
| **Protocol semantics removal** | **argot** | No function names, no field names, no stable structure on the wire |
| Fingerprint resistance | argot | A passive observer cannot classify the application protocol |
| Cross-language contract | argot manifest | JSON manifest + wire spec; Go is only the reference implementation |

argot is **not** an encryption scheme. If you remove Noise and rely on argot alone, your traffic is readable — it is merely unlabelled. The two are complementary: Noise protects *who* is talking and *what* they say; argot removes *what protocol they are speaking* from view.

## Architecture

```
         ┌─────────────┐    plaintext spec     ┌──────────────────────┐
         │ Protocol    │ ────────────────────> │ Cloud Argot Service  │
         │ author (you)│   (functions+schemas) │  Compile(spec, key)  │
         └─────────────┘                       └──────────┬───────────┘
                                                          │ manifest (signed)
                                                          │ distributed over Noise/HTTPS
                                    ┌─────────────────────┴─────────────────────┐
                                    ▼                                           ▼
                          ┌──────────────────┐                       ┌──────────────────┐
                          │  Private client  │                       │  Private server  │
                          │  LoadArgot(...)  │                       │  LoadArgot(...)  │
                          └────────┬─────────┘                       └─────────┬────────┘
                                   │        Noise transport + argot wire       │
                                   └─────────────────┬─────────────────────────┘
                                                     ▼
                                    observer sees: [opcode][slots][decoy][pad]
                                    — no names, no tags, no magic, no structure
```

Only the cloud service ever sees the plaintext spec **and** the obfuscation key at the same time. The client and server hold only the manifest — never the key.

## Core Concepts

| Concept | Type | Lifetime | Purpose |
|---|---|---|---|
| **spec** | `*Protocol` | authoring time | Plaintext function + parameter schema. May be public. |
| **compiler** | `*Compiler` | cloud process | Turns a spec into a manifest. Deterministic given `(key, spec)`. |
| **manifest** | `*Manifest` | long-lived artifact | The codebook. Language-neutral JSON. Secret. |
| **codec** | `*Argot` | per connection/process | `manifest` loaded into two lookup tables; does encode/decode. |

### Why `Funcs` is a map, not a slice

```go
Funcs map[string]*Func // function name → definition
```

* **Reverse lookup is the core codec operation.** Decoding maps an opcode back to a function; encoding maps a name to a layout. A map makes "name is the identity" explicit at the spec level, and maps straight onto a JSON object / Python dict / JS object in non-Go implementations.
* **Duplicate registration is impossible by construction.** With a slice, two entries named `fs.read` are both accepted and only collide at compile time; with a map, the key *is* the single source of truth.
* **No `Func.Name` field**, because the key already carries it.

Deterministic compilation is unaffected: the compiler sorts the map keys before deriving any layout, so iteration order never leaks into the output.

### Deterministic, not random

`Compile` is deterministic pseudo-random, not random:

* **Same `key` + same `spec` ⇒ byte-identical manifest, every time.** Obfuscation draws from a PRNG seeded by HMAC key derivation, over a sorted key traversal.
* Why it matters: the cloud service can be stateless and horizontally scaled (any replica can regenerate the same manifest), manifests can be cached / compared / diffed, and audit is possible.
* Rotation is *changing the language*, not re-rolling dice: change the key (or derive a per-epoch key) and recompile.

### What compilation actually obfuscates

| Element | Plaintext | After compilation |
|---|---|---|
| Function identity | `"fs.read"` | `uint16` opcode from a keyed shuffle of the function set |
| Parameter order | declaration order | shuffled per function by keyed PRNG |
| Field identity | `"path"`, `"offset"` | no field name on the wire at all — position only |
| Type indication | type tag | absent; decode is driven by the manifest |
| Framing | message boundaries | padded to keyed bucket sizes |
| Decoys | none | keyed number of fake slots interleaved |

The opcode is produced by a **keyed shuffle** of the function set, not by `hash(name)`. A hash is trivially reversed by dictionary enumeration of plausible function names; a shuffle reveals neither the name space nor the registration order.

## Installation

```go
import "github.com/leoheung/go-patterns/net/argot"
```

**Dependencies:** none beyond the Go standard library (`crypto/hmac`, `crypto/sha256`, `encoding/json`, `math/rand/v2`).

## Quick Start (Cloud First)

The cloud side is always the first step: obfuscate the protocol, then distribute the manifest.

```go
package main

import (
    "crypto/ed25519"
    "fmt"
    "log"

    "github.com/leoheung/go-patterns/net/argot"
)

func main() {
    // 1. The protocol author's plaintext spec.
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

    // 2. Obfuscate. The key lives in the cloud only (KMS / HSM in production).
    key := loadObfuscationKey()             // []byte, never leaves the cloud
    manifest, err := argot.NewCompiler(key).Compile(proto)
    if err != nil {
        log.Fatal(err)
    }

    // 3. Serialize and sign.
    payload, err := argot.MarshalManifest(manifest)
    if err != nil {
        log.Fatal(err)
    }
    sig := ed25519.Sign(signingKey, payload)

    fmt.Printf("manifest fingerprint = %s\n", manifest.Fingerprint())
    serveSignedManifest(payload, sig) // GET /argot/v1/latest
}
```

## Complete Example: Cloud-First Obfuscation Flow

A single end-to-end walkthrough: define the protocol, obfuscate it in the cloud, distribute the manifest, then run one call over Noise.

### Step 0 — Define the plaintext protocol

This is the only place where names and schemas exist in readable form.

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

### Step 1 — Cloud: obfuscate the protocol

The cloud service is the only place where the spec and the key meet. Its entire job is `Compile` plus signing.

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
    key       []byte            // from KMS; never logged, never returned
    signer    ed25519.PrivateKey
    mu        sync.RWMutex
    manifest  []byte            // cached serialized manifest
    signature []byte
}

func NewArgotService(key []byte, signer ed25519.PrivateKey) *ArgotService {
    return &ArgotService{key: key, signer: signer}
}

// Publish compiles the spec and caches the signed manifest.
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

// Handler serves the signed manifest over the encrypted distribution channel.
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
    key, err := rand.Read(make([]byte, 32)) // replaced by KMS in production
    if err != nil {
        log.Fatal(err)
    }
    _, signer, _ := ed25519.GenerateKey(rand.Reader)

    svc := NewArgotService(loadObfuscationKey(), signer)
    if err := svc.Publish(mytoolsProtocol()); err != nil {
        log.Fatal(err)
    }
    log.Fatal(http.ListenAndServe(":8443", svc.Handler())) // behind Noise/HTTPS + mTLS
}
```

**Key points about this step**

* `Compile` is deterministic, so the service can drop the cache and recompile on demand with identical output — no manifest backup required.
* The spec never reaches a client. The key never leaves the cloud. Clients receive only the derived manifest.
* Sign the manifest before distribution. A tampered manifest is not "slightly broken obfuscation" — it is attacker-chosen wire semantics.

### Step 2 — Both ends: fetch, verify, load

The same code runs on the private client and on the private server.

```go
func loadManifest(httpc *http.Client, verifyKey ed25519.PublicKey) (*argot.Argot, error) {
    // 1. Fetch inside the encrypted channel (Noise over TCP, or HTTPS + pinning).
    resp, err := httpc.Get("https://argot.example.com/argot/v1/latest")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    payload, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // 2. Verify the signature before trusting a single byte of it.
    sig, err := base64.StdEncoding.DecodeString(resp.Header.Get("X-Argot-Signature"))
    if err != nil {
        return nil, err
    }
    if !ed25519.Verify(verifyKey, payload, sig) {
        return nil, errors.New("argot: manifest signature mismatch")
    }

    // 3. Load into a codec: two lookup tables, no key involved.
    a, err := argot.LoadArgot(payload)
    if err != nil {
        return nil, err
    }

    // 4. After the Noise handshake, confirm both peers speak the same language.
    if a.Fingerprint() != expectedFingerprint {
        return nil, errors.New("argot: peer speaks a different argot")
    }
    return a, nil
}
```

`LoadArgot` builds both indexes in one pass: `byName` (function name → `FuncLayout`, used when encoding) and `byOp` (opcode → function name, used when decoding). This pair of maps *is* the whole codec; a non-Go implementation is complete once it can build the same two dicts from the manifest.

### Step 3 — One call over Noise

Client side:

```go
// Application code is unchanged: it calls by name, with named arguments.
payload, err := a.EncodeCall("fs.read", map[string]any{
    "path":   "/etc/hosts",
    "offset": 0,
})
if err != nil {
    return err
}
noiseConn.Write(payload) // No read of names/order/types happens above this line
```

Server side:

```go
func serve(conn *noise.Conn, a *argot.Argot, reg map[string]Handler) error {
    buf := make([]byte, 64*1024)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            return err
        }

        // Decode: opcode → name, slots → named args.
        name, args, err := a.DecodeCall(buf[:n])
        if err != nil {
            return err // protocol-level error: never sent on the wire
        }

        h, ok := reg[name]
        if !ok {
            return fmt.Errorf("argot: no handler for %q", name)
        }

        values, appErr := h(args)
        res := argot.Result{Values: values}
        if appErr != nil {
            res.Err = appErr.Error() // application-level error: travels on the wire
        }

        reply, err := a.EncodeResult(name, res)
        if err != nil {
            return err
        }
        conn.Write(reply)
    }
}
```

Client side, decoding the reply:

```go
res, err := a.DecodeResult(reply, "fs.read") // name selects the result layout
if err != nil {
    return err
}
if res.Err != "" {
    return errors.New(res.Err)
}
fmt.Printf("read %d bytes\n", len(res.Values["data"].([]byte)))
```

### What the observer sees

The wire bytes for the call above (illustrative values — the real bytes depend on the key):

```
bc 55  00  0a 2f 65 74 63 2f 68 6f 73 74 73  d6 9e 3c  ...
│      │   │  └────────── decoy slot (discarded) ──────────┘
│      │   └──── offset=0, varint-zigzag → 0x00
│      └──────── (slot 1: path) len16-prefixed → 0x000a then 10 UTF-8 bytes
└─────────────── opcode 0x55bc = 21948 (keyed shuffle result)
```

| Viewpoint | What is recoverable |
|---|---|
| Observer, no manifest | Nothing. No magic, no field names, no type tags. Field boundaries are not derivable; padding makes lengths meaningless. |
| Observer, manifest leaked | Everything — the manifest *is* the codebook. |
| Client / server, manifest loaded | Function identity, field identity, order, types, and decoys (discarded). |

Note the consequence: **the manifest is the crown jewel**. Its confidentiality is the entire obfuscation guarantee. That is why it is distributed over Noise/HTTPS, signed, and never embedded in the binary as a fallback.

## API Reference

### Spec types

```go
type ValueType int

const (
    ValString ValueType = iota // UTF-8 string
    ValBytes                   // raw []byte
    ValInt                     // int64
    ValUint                    // uint64
    ValFloat                   // float64
    ValBool                    // bool
    ValJSON                    // arbitrary JSON value (passed through as any)
)

type Protocol struct {
    Name    string           // protocol name (metadata only, never on the wire)
    Version uint32           // must match on both ends
    Funcs   map[string]*Func // function name → definition; key is the identity
}

type Func struct {
    Params  []Param // declaration order is for readability only; the wire order is shuffled
    Results []Param
}

type Param struct {
    Name string    // compile-time only; never on the wire
    Type ValueType
}
```

### Compiler

```go
type Compiler struct{ /* key, never exposed */ }

func NewCompiler(key []byte) *Compiler
func (c *Compiler) Compile(p *Protocol) (*Manifest, error)
```

`Compile` runs in the cloud only. It sorts the spec's map keys before deriving layouts, then emits the manifest without retaining the key.

### Manifest

```go
type Manifest struct {
    ArgotVersion uint32                 // format version; the cross-language compatibility gate
    ProtoName    string
    ProtoVersion uint32
    Funcs        map[string]*FuncLayout // plaintext name → obfuscated layout
    PadPolicy    PadPolicy
}

type FuncLayout struct {
    Opcode  uint16       // keyed-shuffle output; no relation to name or order
    Params  []SlotLayout // already in wire order
    Results []SlotLayout
    Decoys  []SlotLayout // fake slots, discarded on decode
}

type SlotLayout struct {
    Ref      string  // which spec param this slot carries (generated name for decoys)
    Type     ValueType
    Encoding WireEnc // chosen at compile time
}

func MarshalManifest(m *Manifest) ([]byte, error)
func LoadManifest(data []byte) (*Manifest, error)
```

### Codec (Go reference implementation)

```go
type Argot struct{ /* manifest + byName + byOp */ }

func LoadArgot(manifestJSON []byte) (*Argot, error)
func (a *Argot) Manifest() *Manifest
func (a *Argot) Fingerprint() string // digest of the manifest, exchanged after the Noise handshake

func (a *Argot) EncodeCall(name string, args map[string]any) ([]byte, error)
func (a *Argot) DecodeCall(data []byte) (name string, args map[string]any, err error)

type Result struct {
    Err    string         // application error text; empty means success
    Values map[string]any // named per the Results schema
}

func (a *Argot) EncodeResult(name string, r Result) ([]byte, error)
func (a *Argot) DecodeResult(data []byte, name string) (Result, error)
```

Two error classes, deliberately kept apart:

* **Protocol errors** (unknown opcode, missing/extra bytes, bad encoding) are Go `error`s returned locally. They are never transmitted — sending them would leak which opcode failed.
* **Application errors** (`Result.Err`) travel on the wire, because the peer needs them.

### Dispatcher (optional)

```go
type Handler func(args map[string]any) (map[string]any, error)

type Dispatcher struct{ /* argot + handlers */ }

func NewDispatcher(a *Argot) *Dispatcher
func (d *Dispatcher) Register(name string, h Handler)
func (d *Dispatcher) Handle(payload []byte) (reply []byte, err error)
```

`Handle` is the server-side glue: decode → dispatch → encode. It carries no obfuscation logic of its own and can be omitted; the example above does the same work inline.

## The Manifest (Interchange Format)

The manifest is the only artifact that crosses the language boundary. It is JSON so that any runtime can consume it.

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

Observe: `fs.read`'s params appear as `offset, path` — the reverse of the spec's declaration order. `fs.write`'s results use plain `varint` while `fs.read`'s params use `varint-zigzag`, for the same logical type. Nothing here is derivable from the plaintext spec without the key.

| Field | Meaning |
|---|---|
| `argot_version` | Wire/manifest format version. A peer with a different value cannot interoperate. |
| `proto_name`, `proto_version` | Traceability back to the authoring spec. Never transmitted. |
| `pad_policy.decoy_min/max` | Range of fake slots injected per message. |
| `pad_policy.bucket` | Message length is rounded up to a multiple of this. |
| `funcs.<name>.opcode` | Wire identity. Unique across the manifest. |
| `params` / `results` | **Array index = wire position.** This is the entire "order" mechanism. |
| `decoys` | Appended slots; decoded and discarded. |

### Implementing a codec in another language

No Go toolchain, no key, and no compiler are needed. Given the manifest:

1. Build `byName: dict[name → FuncLayout]` and `byOp: dict[opcode → name]`.
2. For an encode, look up `byName[name]`, write `opcode` big-endian u16, then walk `params` **in array order**, encoding each argument per its `encoding` and `type`.
3. Write the `decoys` slots with bytes drawn from any source (they are discarded by the peer; random content is fine and preferable).
4. Round the message up to the next `pad_policy.bucket` multiple.
5. For a decode, read the u16 opcode, resolve via `byOp`, then walk the same slots in the same order, binding each to `ref`.

That is the whole contract. Everything else is application code.

## Wire Format Specification

This section is normative. Two implementations that follow it interoperate.

### Call frame

```
+-----------+----------------+----------------+--------------+--------+
| opcode u16| slot[0]        | slot[1]        | decoy slots  | pad    |
| (big end.)|                |                | (0..n)       |        |
+-----------+----------------+----------------+--------------+--------+
```

* `opcode`: 2 bytes, big-endian, value taken from the manifest's `funcs.<name>.opcode`.
* Slots: exactly `len(params)` real slots in manifest array order, then `len(decoys)` decoy slots, then padding.
* Padding: arbitrary bytes to reach the next multiple of `pad_policy.bucket`. Padding length is implicit (total length is already known from the transport framing).
* No magic number, no version byte, no field count, no terminator.

### Result frame

```
+-----------+---------+---------------------+--------------+--------+
| opcode u16| status  | (result slots  |    | decoy slots  | pad    |
|           | u8      |  or error string)   |              |        |
+-----------+---------+---------------------+--------------+--------+
```

* `opcode`: same value as the request's opcode, so a decoder knows which `results` layout applies — which is why `DecodeResult` takes the function name explicitly.
* `status`: `0x00` success, `0x01` application error. Any other value is a protocol error.
* On success: `len(results)` slots in manifest order.
* On error: a single `len16-prefixed` UTF-8 string (the application error text), then padding. The results layout is not written.
* The status byte is the only self-describing element in the entire protocol; it carries two bits of information and no protocol identity.

### Encoding variants

`WireEnc` names are stable identifiers used in the manifest.

| Encoding | Wire form | Applies to |
|---|---|---|
| `varint` | unsigned LEB128 | `ValUint`, non-negative `ValInt` |
| `varint-zigzag` | zigzag then unsigned LEB128 | `ValInt` (handles negatives compactly) |
| `fixed64` | 8 bytes, IEEE-754 big-endian | `ValFloat`; also usable for `ValInt`/`ValUint` when the compiler picks a fixed variant |
| `fixed8` | 1 byte, `0x00` / `0x01` | `ValBool` |
| `len16-prefixed` | 2 bytes big-endian length, then payload | `ValString` (UTF-8), `ValBytes` ≤ 65535 |
| `len32-prefixed` | 4 bytes big-endian length, then payload | `ValBytes` > 65535, `ValJSON` |

`ValJSON` is serialized with the usual compact JSON encoding and then length-prefixed; it is a "payload" type, not a self-describing one — the type information lives only in the manifest.

The compiler chooses among applicable variants per slot using the keyed PRNG, so the same logical type may appear differently in different functions, and differently again after a key rotation.

## Security Model

**What argot does**

* Removes function names, field names, field order, type tags, magic numbers, stable framing, and stable lengths from the wire.
* Denies a passive observer any way to classify the application protocol or build a signature for it.
* Makes traffic from two different protocols compiled with different keys non-correlatable.

**What argot does not do**

* It does not encrypt. Bytes are readable without Noise.
* It does not authenticate peers. Noise does that.
* It does not resist an active prober who can force the peer to reveal behaviour.
* It does not protect against endpoint compromise: whoever reads the client's memory reads the manifest, and the obfuscation collapses to zero.
* Traffic analysis is only weakened, not defeated: padding mitigates length fingerprinting but does not eliminate timing signals.

**Threats and their answers**

| Threat | Answer |
|---|---|
| Passive DPI classifying the protocol | Manifest-absent wire has no identifying structure |
| Replay detection / protocol fingerprinting via stable opcodes | Opcodes are per-key; rotation changes them all |
| Tampered manifest injected by a compromised distribution path | ed25519 signature, verified before load |
| Manifest leaks | Obfuscation is lost; re-key (`Compile` with a new key) to restore it. Noise still protects content. |
| Cloud service is compromised | Spec and key are exposed; re-key and re-publish. Keep the spec authoring side separate from the signing key. |

**The function-name question.** `funcs` keeps plaintext names as map keys for readability and debugging. Because the manifest is distributed only over authenticated, encrypted channels and is a secret artifact regardless, this adds no exposure: the manifest must be protected either way. If your threat model requires the manifest to withstand partial disclosure, derive keys as `HMAC(key, name)` instead — at the cost of losing human readability.

**Version semantics.** `argot_version` governs format compatibility; `proto_version` governs application compatibility. Bump `proto_version` whenever the function set or schemas change; the fingerprint check after the Noise handshake turns a mismatch into an immediate, explicit failure rather than a stream of decode errors.

## Features

* **Language-neutral by construction** — the manifest is JSON, the wire format is specified in bytes; Go is only the reference implementation
* **Cloud-first** — only the service ever holds both the spec and the key
* **Deterministic compilation** — stateless service, reproducible manifests, diffable artifacts
* **No self-describing wire elements** — no names, no tags, no magic, no stable fingerprints
* **Signed distribution** — ed25519 signature verified before the manifest is loaded
* **Per-connection fingerprint check** — peers confirm they speak the same argot
* **Zero third-party dependencies** — standard library only
* **Transport-agnostic** — designed to sit beneath Noise, but works over any byte-stream transport

## Notes

* The compiler runs in the cloud only. Clients and servers never see the obfuscation key, only the derived manifest.
* Compilation is deterministic: the same `(key, spec)` produces the same manifest, so the service needs no manifest persistence.
* `Compile` output is not a security boundary by itself; the distribution channel and signature are.
* Changing the key changes every opcode, every parameter order, and every encoding choice — this is the rotation mechanism.
* The optional `Dispatcher` carries no obfuscation logic; it is convenience glue only.
* `argot` is unrelated to encryption; use it together with Noise, not instead of it.
