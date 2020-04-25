package app

import (
	"fmt"
	"net"
	"time"
	"sync"
)

type Tunnel struct {
	Type          int
	ConnectionUid string
	TunnelIp      string
	TunnelPort    int
	DevicePort    int
}

func (t *Tunnel) Start() {

	deviceConnDialer := net.Dialer{Timeout: 1 * time.Second}
	deviceConn, err := deviceConnDialer.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", t.DevicePort))
	if err != nil {
		return
	}

	tunnelConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", t.TunnelIp, t.TunnelPort))
	if err != nil {
		deviceConn.Close()
		return
	}

	_, err = tunnelConn.Write([]byte(t.ConnectionUid))
	if err != nil {
		tunnelConn.Close()
		deviceConn.Close()
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			var buf = make([]byte, 2048)
			n, err := deviceConn.Read(buf)
			if err != nil {
				deviceConn.Close()
				tunnelConn.Close()
				break
			}

			tunnelConn.Write(buf[:n])
		}
	}()

	go func() {
		defer wg.Done()

		for {
			var buf = make([]byte, 2048)
			n, err := tunnelConn.Read(buf)
			if err != nil {
				deviceConn.Close()
				tunnelConn.Close()
				break
			}

			deviceConn.Write(buf[:n])
		}
	}()

	wg.Wait()
}
