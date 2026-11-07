# Process

The `process` package manages OS child processes: start them, watch for their exit, and stop them again — each by an ID of the caller's choice.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/process"
```

## ProcessManager

A `ProcessManager` tracks a set of child processes under IDs of your choice. It is safe for use from multiple goroutines at once.

### Start a Process

```go
m := process.NewProcessManager()

// nil dir = the manager's working directory; nil env = inherit ours
p, err := m.Start("recorder", `C:\apps\rec.exe`, []string{"serve"}, "", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(p.ID, p.Pid, p.StartedAt)
```

On Windows the child comes up without a console window of its own (CREATE_NO_WINDOW + hidden), which suits background helpers.

### Watch for Exit

```go
// From any number of goroutines:
<-p.Done()      // closes once the process exits, on its own or by Stop

err := p.Wait() // blocks until exit; returns the wait error (nil = success)
fmt.Println(p.Exited(), p.Err())
```

### Stop

```go
_ = m.Stop("recorder") // kill + wait; idempotent (unknown or exited id = no-op)
```

### Inspect

```go
_ = m.Get("recorder") // *Process or nil
for _, p := range m.List() { ... } // snapshot, sorted by ID
fmt.Println(m.String())            // one line per process, oldest first
```

### Shut Everything Down

```go
m.StopAll() // stops every tracked process concurrently and waits for each
```

## Notes

- A duplicate or still-tracked ID makes `Start` fail; a start that fails leaves nothing registered.
- Exited processes stay listed (as `exited`) until a later `Start` reuses their ID.
- `Stop` reports only real kill errors: killing a process that already exited is a no-op.
