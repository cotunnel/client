package utils

import (
	"github.com/pkg/errors"
	"os/exec"
)

func GetTerminals(terminalTagName string, idSize int) []string {
	return []string{}
}

func CmdExit(cmd *exec.Cmd) error {
	return errors.New("not supported operation system")
}

func CmdRestart() error {
	return errors.New("not supported operation system")
}
