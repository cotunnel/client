package app

import (
	"fmt"
	"net"
	"sync"
	"crypto/tls"
	"time"
)

type Tunnel struct {
	Type          int
	ConnectionUid string
	TunnelIp      string
	TunnelPort    int
	DevicePort    int
	DeviceTlsEnabled    byte
}

func (t *Tunnel) Start() {

	var err error
	var deviceConn net.Conn

	deviceAddr := fmt.Sprintf("127.0.0.1:%d", t.DevicePort)

	deviceConnDialer := net.Dialer{Timeout: 1 * time.Second}

	if t.DeviceTlsEnabled == 0 {

		deviceConn, err = deviceConnDialer.Dial("tcp", deviceAddr)
		if err != nil {
			return
		}
	} else {

		deviceConn, err = tls.DialWithDialer(&deviceConnDialer, "tcp", deviceAddr, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return
		}
	}

	tunnelConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", t.TunnelIp, t.TunnelPort), &tls.Config{
		InsecureSkipVerify: true,
	})

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
