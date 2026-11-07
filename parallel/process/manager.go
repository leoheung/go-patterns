// Package process manages OS child processes: start them, watch for
// their exit, and stop them again — each by an ID of the caller's choice.
//
// A ProcessManager is safe for use from multiple goroutines at once. It
// holds no goroutines of its own until Start gives a process one, which
// ends when that process does.
package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Process is one child process a ProcessManager started.
// Its fields are fixed when Start returns; read its state through the
// methods, which are safe from any goroutine.
type Process struct {
	// ID identifies the process to its manager, of the caller's choice.
	ID string
	// Exe is the executable path Start was given.
	Exe string
	// Args are the arguments Start was given, without the executable.
	Args []string
	// Dir is the working directory Start was given, or "" for the
	// manager process's own.
	Dir string
	// Pid is the operating system's process ID.
	Pid int
	// StartedAt is when Start returned.
	StartedAt time.Time

	proc     *os.Process
	done     chan struct{}
	exited   atomic.Bool
	exitMu   sync.Mutex
	exitEr   error
	killOnce sync.Once
	killErr  error
}

// Done reports a channel the manager closes once the process has exited,
// on its own or by Stop, which many goroutines can wait on. Wait or Err
// tell how it went.
func (p *Process) Done() <-chan struct{} { return p.done }

// Wait blocks until the process exits and returns its wait error, nil if
// it exited with success.
func (p *Process) Wait() error {
	<-p.done
	p.exitMu.Lock()
	defer p.exitMu.Unlock()
	return p.exitEr
}

// Exited reports whether the process has exited.
func (p *Process) Exited() bool { return p.exited.Load() }

// Err returns the wait error of a process that has exited, and nil of one
// still running. Call it after Done or Wait, which it does not wait for.
func (p *Process) Err() error {
	p.exitMu.Lock()
	defer p.exitMu.Unlock()
	return p.exitEr
}

// ProcessManager starts, tracks and stops a set of child processes, each
// under an ID of the caller's choice.
//
// The zero value is ready to use.
type ProcessManager struct {
	mu    sync.RWMutex
	procs map[string]*Process
}

// NewProcessManager creates an empty ProcessManager.
func NewProcessManager() *ProcessManager {
	return &ProcessManager{procs: make(map[string]*Process)}
}

// Start runs exe with args in dir and tracks it under id.
//
// dir may be "" for the manager process's working directory; env may be
// nil to inherit the manager process's environment, or a full one of
// "KEY=value" strings to give the child. A duplicate or a still-tracked
// id is an error, as is a start that fails; on Windows the child comes
// up without a console window of its own.
//
// The Process it returns is the one Get and List give back.
func (m *ProcessManager) Start(id, exe string, args []string, dir string, env []string) (*Process, error) {
	if id == "" {
		return nil, fmt.Errorf("process id is empty")
	}

	m.mu.Lock()
	if m.procs == nil {
		m.procs = make(map[string]*Process)
	}
	if old, dup := m.procs[id]; dup && !old.Exited() {
		// Only a live process blocks its id; an exited one it replaces.
		m.mu.Unlock()
		return nil, fmt.Errorf("process id already in use: %s", id)
	}
	p := &Process{
		ID:        id,
		Exe:       exe,
		Args:      args,
		Dir:       dir,
		StartedAt: time.Now(),
		done:      make(chan struct{}),
	}
	// Register before Start so a Start that fails leaves nothing behind.
	m.procs[id] = p
	m.mu.Unlock()

	cmd := exec.Command(exe, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.SysProcAttr = procAttr()
	if err := cmd.Start(); err != nil {
		m.mu.Lock()
		delete(m.procs, id)
		m.mu.Unlock()
		return nil, fmt.Errorf("start %s (%s): %w", id, exe, err)
	}
	p.proc = cmd.Process
	p.Pid = cmd.Process.Pid

	go func() {
		err := cmd.Wait()
		p.exitMu.Lock()
		p.exitEr = err
		p.exitMu.Unlock()
		p.exited.Store(true)
		close(p.done)
	}()
	return p, nil
}

// Stop kills the process under id and waits for it to exit. It is
// idempotent: an unknown id, or one whose process already exited, is a
// no-op that returns nil. It returns the kill error, if the operating
// system reports one.
func (m *ProcessManager) Stop(id string) error {
	m.mu.RLock()
	p := m.procs[id]
	m.mu.RUnlock()
	if p == nil || p.Exited() {
		return nil
	}
	return p.stop()
}

// StopAll stops every tracked process concurrently and waits until each
// has exited. Use it on the way out of a program that starts children.
func (m *ProcessManager) StopAll() {
	var wg sync.WaitGroup
	for _, p := range m.List() {
		if p.Exited() {
			continue
		}
		wg.Add(1)
		go func(p *Process) {
			defer wg.Done()
			_ = p.stop()
		}(p)
	}
	wg.Wait()
}

// Get returns the process under id, or nil of none.
func (m *ProcessManager) Get(id string) *Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.procs[id]
}

// List returns a snapshot of the tracked processes, by ID. Processes that
// exited stay listed until a later Start reuses their ID.
func (m *ProcessManager) List() []*Process {
	m.mu.RLock()
	list := make([]*Process, 0, len(m.procs))
	for _, p := range m.procs {
		list = append(list, p)
	}
	m.mu.RUnlock()
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}

// String lays out the tracked processes, one per line, oldest first.
func (m *ProcessManager) String() string {
	list := m.List()
	sort.SliceStable(list, func(i, j int) bool { return list[i].StartedAt.Before(list[j].StartedAt) })
	out := "ProcessManager{"
	for i, p := range list {
		if i > 0 {
			out += " | "
		}
		state := "running"
		if p.Exited() {
			state = "exited"
		}
		out += fmt.Sprintf("%s pid=%d %s", p.ID, p.Pid, state)
	}
	return out + "}"
}

// stop kills the process at most once and waits for its exit.
func (p *Process) stop() error {
	p.killOnce.Do(func() {
		if !p.exited.Load() {
			if err := p.proc.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				p.killErr = err
			}
		}
	})
	<-p.done
	return p.killErr
}
