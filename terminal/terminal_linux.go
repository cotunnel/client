package terminal

import (
	"client/pty"
	cmap "github.com/orcaman/concurrent-map"
	"client/utils"
	"os/exec"
	"os/user"
	"syscall"
	"unicode/utf8"
)

type Terminal struct {
	TagName	  string
	Id        string
	Listeners cmap.ConcurrentMap
}

type Listener struct {
	Id  string
	Cmd *exec.Cmd
	Pty *pty.PTY
}

func Create(terminalTagName string, terminalId string) (*Terminal, *exec.Cmd, error) {

	newTerminal := Terminal{
		TagName:   terminalTagName,
		Id:        terminalId,
		Listeners: cmap.New(),
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, nil, err
	}

	defaultShell := utils.GetDefaultShell(currentUser.Username)
	args := []string{"-e^Bb", "-s", defaultShell, "-d", "-m", "-S", newTerminal.TagName + newTerminal.Id}

	cmd := exec.Command("screen", args...)
	cmd.Dir = currentUser.HomeDir

	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	return &newTerminal, cmd, nil
}

func (t *Terminal)Attach(listenerId string) (*Listener, error) {
	cmd := exec.Command("screen", []string{"-rx", t.TagName + t.Id}...)
	ptyIo, err := pty.NewPTY(cmd)
	if err != nil {
		return nil, err
	}

	listener := Listener{
		Id:  listenerId,
		Cmd: cmd,
		Pty: ptyIo,
	}

	return &listener, nil
}

func (t *Terminal) Close() {

	t.Listeners.IterCb(func(key string, value interface{}) {
		value.(*Listener).Close()
		value.(*Listener).Cmd.Wait()
		value.(*Listener).Pty.Master.Close()
	})

	exec.Command("screen", []string{"-X", "-S", t.TagName + t.Id, "kill"}...).Run()
}

func (l *Listener) Write(buff []byte) {
	l.Pty.Master.Write(buff)
}

func (l *Listener) Close() {
	l.Cmd.Process.Signal(syscall.Signal(1))
	l.Cmd.Wait()
	l.Pty.Master.Close()
}

func FilterInvalidUTF8(buf []byte) []byte {
	i := 0
	j := 0
	for {
		r, l := utf8.DecodeRune(buf[i:])
		if l == 0 {
			break
		}
		if r < 0xD800 {
			if i != j {
				copy(buf[j:], buf[i:i+l])
			}
			j += l
		}
		i += l
	}
	return buf[:j]
}
