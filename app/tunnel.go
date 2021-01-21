package app

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"
)

type Tunnel struct {
	Type             int
	ConnectionUid    string
	TunnelIp         string
	TunnelPort       int
	DeviceHost       string
	DevicePort       int
	DeviceTlsEnabled byte

	ExitDeviceReceiverLoop chan bool
	ExitTunnelReceiverLoop chan bool

	DeviceDialerTimeout  int
	TunnelDialerTimeout  int
	TunnelSessionTimeout int

	BufferSize int
}

func handleTcpTunnel(tunnel *Tunnel) {
	deviceAddr := fmt.Sprintf("%s:%d", tunnel.DeviceHost, tunnel.DevicePort)

	var err error
	var deviceConn net.Conn

	deviceConnDialer := net.Dialer{Timeout: time.Duration(tunnel.DeviceDialerTimeout) * time.Millisecond}

	if tunnel.DeviceTlsEnabled == 0 {

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

	tunnelConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", tunnel.TunnelIp, tunnel.TunnelPort), &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		deviceConn.Close()
		return
	}

	_, err = tunnelConn.Write([]byte(tunnel.ConnectionUid))
	if err != nil {
		tunnelConn.Close()
		deviceConn.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-tunnel.ExitDeviceReceiverLoop:
				return
			default:
				var buf = make([]byte, tunnel.BufferSize)
				n, err := deviceConn.Read(buf)
				if err != nil {
					deviceConn.Close()
					tunnelConn.Close()
					return
				}

				tunnelConn.Write(buf[:n])
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-tunnel.ExitTunnelReceiverLoop:
				return
			default:
				var buf = make([]byte, tunnel.BufferSize)
				n, err := tunnelConn.Read(buf)
				if err != nil {
					deviceConn.Close()
					tunnelConn.Close()
					return
				}

				deviceConn.Write(buf[:n])
			}
		}
	}()

	wg.Wait()
}

func handleUdpTunnel(tunnel *Tunnel) {

	deviceAddr := fmt.Sprintf("%s:%d", tunnel.DeviceHost, tunnel.DevicePort)

	localUdpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return
	}

	tunnelUdpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", tunnel.TunnelIp, tunnel.TunnelPort))
	if err != nil {
		return
	}

	deviceRemoteUdpAddr, err := net.ResolveUDPAddr("udp", deviceAddr)
	if err != nil {
		return
	}

	deviceConn, err := net.DialUDP("udp", localUdpAddr, deviceRemoteUdpAddr)
	if err != nil {
		return
	}

	tunnelConn, err := net.DialUDP("udp", localUdpAddr, tunnelUdpAddr)
	if err != nil {
		return
	}

	_, err = tunnelConn.Write([]byte(tunnel.ConnectionUid))
	if err != nil {
		tunnelConn.Close()
		deviceConn.Close()
		return
	}

	// start timer
	udpSessionTimer := time.NewTimer(time.Duration(tunnel.TunnelSessionTimeout) * time.Millisecond)

	go func() {
		<-udpSessionTimer.C
		deviceConn.Close()
		tunnelConn.Close()

		tunnel.ExitDeviceReceiverLoop <- true
		tunnel.ExitTunnelReceiverLoop <- true
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-tunnel.ExitDeviceReceiverLoop:
				return
			default:
				data := make([]byte, tunnel.BufferSize)
				readLen, _, err := deviceConn.ReadFromUDP(data)
				if err != nil {
					deviceConn.Close()
					tunnelConn.Close()
					return
				}

				data = data[:readLen]
				_, err = tunnelConn.Write(data)
				if err != nil {
					deviceConn.Close()
					tunnelConn.Close()
					return
				}

				udpSessionTimer.Reset(time.Duration(tunnel.TunnelSessionTimeout) * time.Millisecond)
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-tunnel.ExitTunnelReceiverLoop:
				return
			default:
				data := make([]byte, tunnel.BufferSize)
				readLen, _, err := tunnelConn.ReadFromUDP(data)
				if err != nil {
					deviceConn.Close()
					tunnelConn.Close()
					return
				}

				data = data[:readLen]
				_, err = deviceConn.Write(data)
				if err != nil {
					deviceConn.Close()
					tunnelConn.Close()
					return
				}

				udpSessionTimer.Reset(time.Duration(tunnel.TunnelSessionTimeout) * time.Millisecond)
			}
		}
	}()

	wg.Wait()
}

func (tunnel *Tunnel) Start() {

	if tunnel.Type == 1 || tunnel.Type == 2 {
		handleTcpTunnel(tunnel)
	} else if tunnel.Type == 3 {
		handleUdpTunnel(tunnel)
	} else {
		// unsupported tunnel type
	}
}
