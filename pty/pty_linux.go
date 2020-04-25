package pty

import (
	"github.com/creack/pty"
	syscall "golang.org/x/sys/unix"
	"os"
	"os/exec"
	"unsafe"
)

type PTY struct {
	Master *os.File
	No     int
}

type WinSize struct {
	row, col, x, y uint16
}

func NewPTY(cmd *exec.Cmd) (*PTY, error) {
	pty, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &PTY{
		Master: pty,
		No:     0,
	}, nil
}

func (pty *PTY) Ioctl(a2, a3 uintptr) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, pty.Master.Fd(), a2, a3); errno != 0 {
		return errno
	}
	return nil
}

func (pty *PTY) SetSize(rows, cols uint16) {

	window := WinSize{row: rows, col: cols, x: 0, y: 0}
	pty.Ioctl(syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&window)))
}
