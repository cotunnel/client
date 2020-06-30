package utils

import (
	"github.com/willdonnelly/passwd"
	syscall "golang.org/x/sys/unix"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func IsRootUser() bool {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	// output has trailing \n
	// need to remove the \n
	// otherwise it will cause error for strconv.Atoi
	// log.Println(output[:len(output)-1])

	// 0 = root, 501 = non-root user
	i, err := strconv.Atoi(string(output[:len(output)-1]))
	if err != nil {
		return false
	}

	if i == 0 {
		return true
	} else {
		return false
	}
}

func GetTerminals(terminalTagName string, idSize int) []string {

	var terminalIds = make([]string, 0)

	out, _ := exec.Command("screen", "-list").Output()
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		split := strings.Split(line, ".")
		if len(split) >= 2 && strings.HasPrefix(split[1], terminalTagName) {
			terminalIds = append(terminalIds, split[1][len(terminalTagName):len(terminalTagName)+idSize])
		}
	}

	return terminalIds
}

func GetShell() string {
	if os.Getenv("SHELL") != "" {
		return os.Getenv("SHELL")
	}
	return "/bin/bash"
}

func GetDefaultShell(username string) string {
	if runtime.GOOS == "darwin" {
		return "/bin/bash"
	}

	entry, err := GetUserEntry(username)
	if err != nil {
		log.Println("term: couldn't get default shell ", err)
		return "/bin/bash"
	}

	return entry.Shell
}

func GetUserEntry(username string) (*passwd.Entry, error) {
	entries, err := passwd.Parse()
	if err != nil {
		return nil, err
	}

	user, ok := entries[username]
	if !ok {
		return nil, err
	}

	if user.Shell == "" {
		return nil, err
	}

	return &user, nil
}

func CmdExit(cmd *exec.Cmd) {
	cmd.Process.Signal(syscall.Signal(1))
	cmd.Wait()
}

func CmdRestart() error {
	binary, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	// trim --key args. we don't want to register after restart
	var manipulatedArgs = make([]string, 0)
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--key" || os.Args[i] == "-key" {
			i++
			continue
		} else {
			manipulatedArgs = append(manipulatedArgs, os.Args[i])
		}
	}

	execErr := syscall.Exec(binary, manipulatedArgs, os.Environ())
	if execErr != nil {
		return err
	}

	return nil
}