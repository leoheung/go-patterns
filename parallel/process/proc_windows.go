//go:build windows

package process

import "syscall"

const createNoWindow = 0x08000000

// procAttr keeps a child off the screen: no console window of its own
// (CREATE_NO_WINDOW), and no taskbar flash for a hidden one.
func procAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: createNoWindow, HideWindow: true}
}
