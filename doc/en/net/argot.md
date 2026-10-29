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

That is the whole contract. Everything else is application code. The Cross-Language Examples section below shows this flow fully implemented in TypeScript and Java.

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

## Cross-Language Examples

The Go cloud service compiled the three-function spec above into a manifest with fingerprint `7f6be209…` and signed it for distribution. Below, a **TypeScript client** and a **Java server** both consume that same manifest — neither needs the Go toolchain, the obfuscation key, or the compiler. The manifest plus the wire specification above is the entire contract.

Both listings were cross-verified against the Go reference implementation: the frame encoded by the TypeScript client is byte-identical to the Go encoder's output for the same call, and the result frame encoded by the Java server decodes correctly under both Go and TypeScript.

### TypeScript client (Node.js)

Node 18+, zero dependencies. Run directly with `node --experimental-strip-types`.

```ts
// argot.ts — client-side reference codec
import { randomBytes } from "node:crypto";

interface Slot { ref: string; type: string; encoding: string }
interface FuncLayout { opcode: number; params: Slot[]; results: Slot[]; decoys: Slot[] }
interface Manifest {
  funcs: Record<string, FuncLayout>;
  pad_policy: { decoy_min: number; decoy_max: number; bucket: number };
}
type Value = number | bigint | string | boolean | Uint8Array | null | object;

// ---- byte-level primitives ----

class Reader {
  private dv: DataView;
  private off = 0;
  private buf: Uint8Array;
  constructor(buf: Uint8Array) {
    this.buf = buf;
    this.dv = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
  }
  u8(): number { return this.dv.getUint8(this.off++); }
  u16(): number { const v = this.dv.getUint16(this.off); this.off += 2; return v; }
  u32(): number { const v = this.dv.getUint32(this.off); this.off += 4; return v; }
  bytes(n: number): Uint8Array { const v = this.buf.subarray(this.off, this.off + n); this.off += n; return v; }
  uvarint(): bigint { // unsigned LEB128
    let result = 0n, shift = 0n;
    for (;;) {
      const b = this.u8();
      result |= BigInt(b & 0x7f) << shift;
      if ((b & 0x80) === 0) return result;
      shift += 7n;
    }
  }
}

class Writer {
  private parts: number[] = [];
  u16(v: number): this { this.parts.push((v >> 8) & 0xff, v & 0xff); return this; }
  uvarint(v: bigint): this {
    let n = v;
    for (;;) {
      const b = Number(n & 0x7fn);
      n >>= 7n;
      this.parts.push(n === 0n ? b : b | 0x80);
      if (n === 0n) return this;
    }
  }
  fixed64(bits: bigint): this { // raw 64-bit pattern, big-endian
    for (let i = 7n; i >= 0n; i--) this.parts.push(Number((bits >> (i * 8n)) & 0xffn));
    return this;
  }
  payload(p: Uint8Array, len32: boolean): this {
    if (len32) {
      this.parts.push((p.length >>> 24) & 0xff, (p.length >>> 16) & 0xff,
                      (p.length >>> 8) & 0xff, p.length & 0xff);
    } else {
      if (p.length > 0xffff) throw new Error("argot: payload too long for len16");
      this.parts.push((p.length >> 8) & 0xff, p.length & 0xff);
    }
    for (const b of p) this.parts.push(b);
    return this;
  }
  pad(bucket: number): this {
    if (bucket > 1 && this.parts.length % bucket !== 0) {
      this.parts.push(...new Array(bucket - (this.parts.length % bucket)).fill(0));
    }
    return this;
  }
  out(): Uint8Array { return Uint8Array.from(this.parts); }
}

const encoder = new TextEncoder();
const decoder = new TextDecoder();
const zigzag = (n: bigint): bigint => (n << 1n) ^ (n >> 63n);

function toBigint(v: Value): bigint {
  if (typeof v === "bigint") return v;
  if (typeof v === "number") return BigInt(Math.trunc(v));
  throw new Error(`argot: cannot use ${typeof v} as int`);
}

function writeSlot(w: Writer, s: Slot, v: Value): void {
  switch (s.encoding) {
    case "varint":        w.uvarint(toBigint(v)); break;
    case "varint-zigzag": w.uvarint(zigzag(toBigint(v))); break;
    case "fixed64": {
      const f64 = new DataView(new ArrayBuffer(8));
      if (s.type === "int")       f64.setBigInt64(0, toBigint(v));
      else if (s.type === "uint") f64.setBigUint64(0, toBigint(v));
      else                        f64.setFloat64(0, Number(v));
      w.fixed64(f64.getBigUint64(0));
      break;
    }
    case "fixed8": w.u8(v === true ? 1 : 0); break;
    case "len16-prefixed":
    case "len32-prefixed": {
      const p = s.type === "json"   ? encoder.encode(JSON.stringify(v))
             : s.type === "bytes" ? (v as Uint8Array)
             :                       encoder.encode(String(v));
      w.payload(p, s.encoding === "len32-prefixed");
      break;
    }
    default: throw new Error(`argot: unknown encoding ${s.encoding}`);
  }
}

function readSlot(r: Reader, s: Slot): Value {
  switch (s.encoding) {
    case "varint":        return r.uvarint();
    case "varint-zigzag": { const u = r.uvarint(); return (u >> 1n) ^ -(u & 1n); }
    case "fixed64": {
      const raw = r.bytes(8);
      const f64 = new DataView(raw.buffer, raw.byteOffset, 8);
      if (s.type === "int")  return f64.getBigInt64(0);
      if (s.type === "uint") return f64.getBigUint64(0);
      return f64.getFloat64(0);
    }
    case "fixed8": return r.u8() === 1;
    case "len16-prefixed":
    case "len32-prefixed": {
      const n = s.encoding === "len16-prefixed" ? r.u16() : r.u32();
      const p = r.bytes(n);
      if (s.type === "json")   return JSON.parse(decoder.decode(p));
      if (s.type === "bytes")  return p;
      return decoder.decode(p);
    }
    default: throw new Error(`argot: unknown encoding ${s.encoding}`);
  }
}

function randomDecoy(type: string): Value {
  switch (type) {
    case "bytes":  return randomBytes(4 + Math.floor(Math.random() * 17));
    case "string": return randomBytes(8).toString("hex");
    case "int":    return BigInt(Math.floor(Math.random() * 200001) - 100000);
    case "uint":   return BigInt(Math.floor(Math.random() * 100000));
    case "float":  return Math.random() * 100;
    case "bool":   return Math.random() < 0.5;
    default:       return { v: Math.floor(Math.random() * 1000) }; // json
  }
}

// ---- the codec itself: two maps over the manifest ----

export class Argot {
  private byName: Map<string, FuncLayout>;
  private byOp = new Map<number, { name: string; layout: FuncLayout }>();
  private man: Manifest;

  constructor(man: Manifest) {
    this.man = man;
    this.byName = new Map(Object.entries(man.funcs));
    for (const [name, layout] of Object.entries(man.funcs)) {
      this.byOp.set(layout.opcode, { name, layout });
    }
  }

  encodeCall(name: string, args: Record<string, Value>): Uint8Array {
    const layout = this.byName.get(name);
    if (!layout) throw new Error(`argot: unknown function ${name}`);
    for (const s of layout.params) if (!(s.ref in args)) throw new Error(`argot: missing arg ${s.ref}`);
    for (const k of Object.keys(args)) {
      if (!layout.params.some(s => s.ref === k)) throw new Error(`argot: unexpected arg ${k}`);
    }

    const w = new Writer();
    w.u16(layout.opcode);
    for (const s of layout.params) writeSlot(w, s, args[s.ref]);
    for (const s of layout.decoys) writeSlot(w, s, randomDecoy(s.type)); // peer discards these
    w.pad(this.man.pad_policy.bucket);
    return w.out();
  }

  decodeResult(data: Uint8Array, name: string): { err: string; values: Record<string, Value> } {
    const layout = this.byName.get(name);
    if (!layout) throw new Error(`argot: unknown function ${name}`);
    const r = new Reader(data);
    if (r.u16() !== layout.opcode) throw new Error("argot: result opcode mismatch");
    const status = r.u8();
    if (status === 1) { // application error: len16-prefixed text
      return { err: decoder.decode(r.bytes(r.u16())), values: {} };
    }
    if (status !== 0) throw new Error(`argot: invalid status 0x${status.toString(16)}`);
    const values: Record<string, Value> = {};
    for (const s of layout.results) values[s.ref] = readSlot(r, s);
    return { err: "", values };
  }
}
```

Ints decode as `BigInt` — narrow with `Number()` when you know the value fits in ±2^53.

Usage:

```ts
import { readFileSync } from "node:fs";

const manifest: Manifest = JSON.parse(readFileSync("manifest.json", "utf8")); // fetched + signature-verified
const a = new Argot(manifest);

// Application code stays name-based and order-free.
const frame = a.encodeCall("fs.read", { path: "/etc/hosts", offset: 0 });
noiseConn.write(frame); // → delivered to the Java server below

const reply = /* frame hex decoded by the Java server below:
                1e85000016636f6e74656e7473206f66202f6574632f686f7374730000… */;
const res = a.decodeResult(reply, "fs.read");
if (res.err === "") {
  const data = Buffer.from(res.values.data as Uint8Array).toString("utf8");
}
```

### Java server (JDK 17+)

Pure JDK, no third-party libraries. Records + `ByteBuffer` + `ByteArrayOutputStream`.

```java
// ArgotServer.java — server-side reference codec
import java.io.ByteArrayOutputStream;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.*;

public class ArgotServer {

    // ---- manifest types (deserialize the signed JSON with Jackson/Gson in production;
    //      never hand-copy layouts — read them from the manifest the cloud service signed) ----
    record Slot(String ref, String type, String encoding) {}
    record FuncLayout(int opcode, List<Slot> params, List<Slot> results, List<Slot> decoys) {}
    record Manifest(Map<String, FuncLayout> funcs, int bucket) {}
    record Call(String name, Map<String, Object> args) {}
    record Result(String err, Map<String, Object> values) {}

    static final class ArgotCodec {
        private final Map<String, FuncLayout> byName = new HashMap<>();
        private final Map<Integer, String> byOp = new HashMap<>();
        private final int bucket;

        ArgotCodec(Manifest man) {
            this.bucket = man.bucket();
            man.funcs().forEach((name, layout) -> {
                byName.put(name, layout);
                byOp.put(layout.opcode(), name);
            });
        }

        /** decodeCall: [opcode u16][param slots][decoy slots][padding] */
        Call decodeCall(byte[] frame) {
            ByteBuffer r = ByteBuffer.wrap(frame);
            int op = r.getShort() & 0xffff;
            String name = byOp.get(op);
            if (name == null) throw new IllegalArgumentException("unknown opcode 0x" + Integer.toHexString(op));
            FuncLayout layout = byName.get(name);

            Map<String, Object> args = new HashMap<>();
            for (Slot s : layout.params()) args.put(s.ref(), readSlot(r, s));
            for (Slot s : layout.decoys()) readSlot(r, s); // must be consumed to advance the offset
            // trailing bytes are padding — ignored
            return new Call(name, args);
        }

        /** encodeResult: [opcode u16][status u8][result slots | len16 err][padding] */
        byte[] encodeResult(String name, Result res) {
            FuncLayout layout = byName.get(name);
            if (layout == null) throw new IllegalArgumentException("unknown function " + name);

            ByteArrayOutputStream w = new ByteArrayOutputStream();
            w.write((layout.opcode() >> 8) & 0xff); w.write(layout.opcode() & 0xff);
            if (res.err() != null && !res.err().isEmpty()) {
                w.write(0x01);
                byte[] p = res.err().getBytes(StandardCharsets.UTF_8);
                w.write((p.length >> 8) & 0xff); w.write(p.length & 0xff);
                w.writeBytes(p);
            } else {
                w.write(0x00);
                for (Slot s : layout.results()) writeSlot(w, s, res.values().get(s.ref()));
            }
            while (bucket > 1 && w.size() % bucket != 0) w.write(0); // padding
            return w.toByteArray();
        }

        private static Object readSlot(ByteBuffer r, Slot s) {
            switch (s.encoding()) {
                case "varint":        return readUvarint(r);
                case "varint-zigzag": { long u = readUvarint(r); return (u >>> 1) ^ -(u & 1); }
                case "fixed64": {     // ByteBuffer is big-endian by default
                    long bits = r.getLong();
                    return s.type().equals("float") ? Double.longBitsToDouble(bits) : bits;
                }
                case "fixed8": {
                    int b = r.get() & 0xff;
                    if (b > 1) throw new IllegalArgumentException("invalid bool byte " + b);
                    return b == 1;
                }
                case "len16-prefixed": return readPayload(r, s, r.getShort() & 0xffff);
                case "len32-prefixed": return readPayload(r, s, r.getInt() & 0x7fffffff);
                default: throw new IllegalArgumentException("unknown encoding " + s.encoding());
            }
        }

        private static Object readPayload(ByteBuffer r, Slot s, int n) {
            byte[] p = new byte[n];
            r.get(p);
            return s.type().equals("bytes") ? p : new String(p, StandardCharsets.UTF_8);
            // json slots are length-prefixed JSON *text* — parse with your JSON library
        }

        private static long readUvarint(ByteBuffer r) { // unsigned LEB128
            long result = 0; int shift = 0;
            for (;;) {
                byte b = r.get();
                result |= (long) (b & 0x7f) << shift;
                if ((b & 0x80) == 0) return result;
                shift += 7;
            }
        }

        private static void writeSlot(ByteArrayOutputStream w, Slot s, Object v) {
            switch (s.encoding()) {
                case "varint" -> writeUvarint(w, ((Number) v).longValue());
                case "varint-zigzag" -> {
                    long n = ((Number) v).longValue();
                    writeUvarint(w, (n << 1) ^ (n >> 63));
                }
                case "fixed64" -> {
                    long bits = s.type().equals("float")
                            ? Double.doubleToRawLongBits(((Number) v).doubleValue())
                            : ((Number) v).longValue();
                    for (int i = 56; i >= 0; i -= 8) w.write((int) (bits >>> i) & 0xff);
                }
                case "fixed8" -> w.write(((Boolean) v) ? 1 : 0);
                case "len16-prefixed", "len32-prefixed" -> {
                    byte[] p = s.type().equals("bytes") ? (byte[]) v
                            : String.valueOf(v).getBytes(StandardCharsets.UTF_8);
                    if (s.encoding().equals("len32-prefixed")) {
                        w.write((p.length >>> 24) & 0xff); w.write((p.length >>> 16) & 0xff);
                        w.write((p.length >>> 8) & 0xff);  w.write(p.length & 0xff);
                    } else {
                        w.write((p.length >> 8) & 0xff);   w.write(p.length & 0xff);
                    }
                    w.writeBytes(p);
                }
                default -> throw new IllegalArgumentException("unknown encoding " + s.encoding());
            }
        }

        private static void writeUvarint(ByteArrayOutputStream w, long v) {
            while ((v & ~0x7fL) != 0) { w.write((int) (v & 0x7f) | 0x80); v >>>= 7; }
            w.write((int) v);
        }
    }

    public static void main(String[] args) throws Exception {
        // In production: Manifest man = jsonParser.parse(signatureVerifiedManifestBytes);
        Manifest man = loadManifest();
        ArgotCodec codec = new ArgotCodec(man);

        // A call frame produced by the TypeScript client (or the Go reference codec):
        byte[] frame = hexToBytes("1e85000a2f6574632f686f737473" + "00".repeat(49));
        Call call = codec.decodeCall(frame);       // → {fs.read, {path="/etc/hosts", offset=0}}

        // Execute and reply:
        Map<String, Object> values = new HashMap<>();
        values.put("data", "contents of /etc/hosts".getBytes(StandardCharsets.UTF_8));
        byte[] reply = codec.encodeResult(call.name(), new Result(null, values));
        noiseConn.write(reply); // → decoded by the TypeScript client's decodeResult
    }
}
```

### Interop checklist

When porting the codec to any language, these are the details that make or break interoperation:

- The `opcode` is an unsigned 16-bit **big-endian** integer; every length prefix and `fixed64` is big-endian too, while LEB128 varints are least-significant-byte first.
- Slot order is **exactly the `params` / `results` array order in the manifest** — never the declaration order of the original spec.
- Decoy count and layout come from the manifest, not the wire: encode random bytes, decode and discard.
- The result frame reuses the request's opcode and carries exactly one status byte (`0x00` results / `0x01` len16-prefixed error text).
- `json` slots are length-prefixed compact JSON *text*; the type information lives only in the manifest.
- Padding is trailing zeros up to the next `bucket` multiple; decoders simply ignore the remainder.

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
