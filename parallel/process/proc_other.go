//go:build !windows

package process

import "syscall"

// procAttr is nothing on the other systems: children follow the parent's
// process group and terminal, as callers there expect.
func procAttr() *syscall.SysProcAttr { return nil }
