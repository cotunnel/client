package pty

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
)

type PTY struct {
	Master *os.File
	No     int
}

type WinSize struct {
	row, col, x, y uint16
}

func NewPTY(cmd *exec.Cmd) (*PTY, error) {
	return nil, errors.New("not supported for darwin")
}

func (pty *PTY) Ioctl(a2, a3 uintptr) error {
	return errors.New("not supported for darwin")
}

func (pty *PTY) SetSize(rows, cols uint16) {
	return
}
