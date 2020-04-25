package terminal

import (
	"client/pty"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"os/exec"
	"unicode/utf8"
)

type Terminal struct {
	TagName   string
	Id        string
	Listeners cmap.ConcurrentMap
}

type Listener struct {
	Id  string
	Cmd *exec.Cmd
	Pty *pty.PTY
}

func Create(terminalTagName string, terminalId string) (*Terminal, *exec.Cmd, error) {
	return nil, nil, errors.New("not supported for windows")
}

func (t *Terminal) Attach(listenerId string) (*Listener, error) {
	return nil, errors.New("not supported for windows")
}

func (t *Terminal) Close() {

}

func (l *Listener) Write(buff []byte) {

}

func (l *Listener) Close() {

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
