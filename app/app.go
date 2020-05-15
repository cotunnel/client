package app

import (
	"client/cog"
	"client/opcode"
	"client/terminal"
	"client/utils"
	"fmt"
	"github.com/cotunnel/packet"
	"github.com/inconshreveable/go-update"
	"github.com/monnand/dhkx"
	"github.com/orcaman/concurrent-map"
	"github.com/shirou/gopsutil/host"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
	"unicode/utf8"
)

const Version = "1.0.0"
const TerminalTagName = "cotunnel"
const TerminalUIdSize = 4

type App struct {
	Options *Options

	Conn net.Conn

	DHGroup       *dhkx.DHGroup
	DHPrivateKey  *dhkx.DHKey
	EncryptionKey []byte

	IsLoggedIn  bool
	IsEncrypted bool

	Terminals cmap.ConcurrentMap

	WriteMutex sync.Mutex

	SafeReconnect bool
}

type Options struct {
	Key   string `hcl:"key"`
	Exit  bool   `hcl:"key"`
	Token string `hcl:"token"`
	Path  string `hcl:"path"`
}

func New(options *Options) (*App, error) {
	return &App{
		Options: options,

		DHGroup:       nil,
		DHPrivateKey:  nil,
		EncryptionKey: make([]byte, 0),

		Terminals: cmap.New(),

		SafeReconnect: false,
	}, nil
}

func (app *App) Run() error {

	go func() {

		for {
			conn, err := net.Dial("tcp", "server.cotunnel.com:17339")
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			if !app.SafeReconnect {
				cog.Print(cog.INFO, "Connected to the server.")
			}

			app.Conn = conn
			app.SafeReconnect = false

			var cache = make([]byte, 0)

			for {
				buffer := make([]byte, 8192)
				n, err := conn.Read(buffer)
				if err != nil {
					break
				}

				if n == 0 {
					break
				}

				data := buffer[0:n]

				cache = append(cache, data...)
				if len(cache) < 4 {
					continue
				}

				for {
					packetSize := (int16)((int16(cache[0]&0xFF) << 8) | int16(cache[1]&0xFF))
					if int(packetSize+4) > len(cache) {
						break
					}

					p := packet.Packet{}

					if app.IsEncrypted {
						decryptedPacketBytes, err := utils.DecryptAES(app.EncryptionKey, cache[4:packetSize+4])
						if err != nil {
							// clear cache
							cache = make([]byte, 0)
							break
						}

						p.Fill(decryptedPacketBytes)
						cog.PrintPacket(cog.DEBUG, 0, p.OpCode, decryptedPacketBytes)
					} else {
						p.Fill(cache[0 : packetSize+4])
						cog.PrintPacket(cog.DEBUG, 0, p.OpCode, cache[0:packetSize+4])
					}

					switch p.OpCode {
					case opcode.S2CDeviceGateway:
						app.S2CDeviceGatewayHandler(p)
					case opcode.S2CDeviceSessionKey:
						app.S2CDeviceSessionKeyHandler(p)
					case opcode.S2CDeviceHandshake:
						app.S2CDeviceHandshakeHandler(p)
					case opcode.S2CDeviceHandshakeOK:
						app.S2CDeviceHandshakeOKHandler()
					case opcode.S2CDeviceRegister:
						app.S2CDeviceRegisterHandler(p)
					case opcode.S2CDeviceLogin:
						app.S2CDeviceLoginHandler(p)
					case opcode.S2CDevicePing:
						app.S2CDevicePingHandler()
					case opcode.S2CDeviceCreateTerminal:
						app.S2CDeviceCreateTerminalHandler(p)
					case opcode.S2CDeviceDeleteTerminal:
						app.S2CDeviceDeleteTerminalHandler(p)
					case opcode.S2CDeviceAttachTerminal:
						app.S2CDeviceAttachTerminalHandler(p)
					case opcode.S2CDeviceWriteTerminal:
						app.S2CDeviceWriteTerminalHandler(p)
					case opcode.S2CDeviceResizeTerminal:
						app.S2CDeviceResizeTerminalHandler(p)
					case opcode.S2CDeviceDetachTerminal:
						app.S2CDeviceDetachTerminalHandler(p)
					case opcode.S2CDeviceTunnel:
						app.S2CDeviceTunnelHandler(p)
					default:
					}

					cache = cache[packetSize+4:]
					if len(cache) == 0 {
						break
					}
				}
			}

			if !app.SafeReconnect {
				cog.Print(cog.INFO, "The server was disconnected. The client will try to reconnect after 10 seconds.")
				time.Sleep(10 * time.Second)
			}

			// reset variables
			app.IsEncrypted = false
			app.DHGroup = nil
			app.DHPrivateKey = nil
			app.EncryptionKey = make([]byte, 0)
		}
	}()

	return nil
}

func (app *App) S2CDeviceTunnelHandler(p packet.Packet) {

	tunnelType, err := p.ReadByte()
	if err != nil {
		return
	}

	connectionUid, err := p.ReadString()
	if err != nil {
		return
	}

	tunnelIp, err := p.ReadString()
	if err != nil {
		return
	}

	tunnelPort, err := p.ReadInteger()
	if err != nil {
		return
	}

	devicePort, err := p.ReadInteger()
	if err != nil {
		return
	}

	deviceTlsEnabled, err := p.ReadByte()
	if err != nil {
		return
	}

	tunnel := Tunnel{
		Type:             int(tunnelType),
		ConnectionUid:    connectionUid,
		TunnelIp:         tunnelIp,
		TunnelPort:       tunnelPort,
		DevicePort:       devicePort,
		DeviceTlsEnabled: deviceTlsEnabled,
	}

	go tunnel.Start()
}

func (app *App) S2CDeviceResizeTerminalHandler(p packet.Packet) {

	terminalUId, err := p.ReadString()
	if err != nil {
		return
	}

	thisTerminal := app.GetTerminal(terminalUId)
	if thisTerminal == nil {
		return
	}

	listenerId, err := p.ReadString()
	if err != nil {
		return
	}

	listener, found := thisTerminal.Listeners.Get(listenerId)
	if !found {
		return
	}

	rows, err := p.ReadInteger()
	if err != nil {
		return
	}

	cols, err := p.ReadInteger()
	if err != nil {
		return
	}

	listener.(*terminal.Listener).Pty.SetSize(uint16(rows), uint16(cols))
}

func (app *App) S2CDeviceWriteTerminalHandler(p packet.Packet) {

	terminalUId, err := p.ReadString()
	if err != nil {
		return
	}

	thisTerminal := app.GetTerminal(terminalUId)
	if thisTerminal == nil {
		return
	}

	listenerId, err := p.ReadString()
	if err != nil {
		return
	}

	listener, found := thisTerminal.Listeners.Get(listenerId)
	if !found {
		return
	}

	buff, err := p.ReadBytes()
	if err != nil {
		return
	}

	listener.(*terminal.Listener).Write(buff)
}

func (app *App) S2CDeviceAttachTerminalHandler(p packet.Packet) {

	terminalUId, err := p.ReadString()
	if err != nil {
		return
	}

	thisTerminal := app.GetTerminal(terminalUId)
	if thisTerminal == nil {
		return
	}

	listenerId, err := p.ReadString()
	if err != nil {
		return
	}

	_, found := thisTerminal.Listeners.Get(listenerId)
	if found {
		return
	}

	listener, err := thisTerminal.Attach(listenerId)
	if err != nil {
		return
	}

	thisTerminal.Listeners.Set(listenerId, listener)

	r := packet.Packet{}
	r.New(opcode.C2SDeviceAttachTerminal)
	r.WriteString(thisTerminal.Id)
	r.WriteString(listener.Id)
	app.Write(r)

	go func() {
		buff := make([]byte, 4096-utf8.UTFMax, 4096)
		for {
			n, err := listener.Pty.Master.Read(buff)
			if err != nil {
				break
			}

			for n < cap(buff)-1 {
				r, _ := utf8.DecodeLastRune(buff[:n])
				if r != utf8.RuneError {
					break
				}
				listener.Pty.Master.Read(buff[n : n+1])
				n++
			}

			r := packet.Packet{}
			r.New(opcode.C2SDeviceReadTerminal)
			r.WriteString(thisTerminal.Id)
			r.WriteString(listener.Id)
			r.WriteBytes(terminal.FilterInvalidUTF8(buff[:n]))
			app.Write(r)
		}
	}()
}

func (app *App) S2CDeviceDetachTerminalHandler(p packet.Packet) {

	terminalUId, err := p.ReadString()
	if err != nil {
		return
	}

	thisTerminal := app.GetTerminal(terminalUId)
	if thisTerminal == nil {
		return
	}

	listenerId, err := p.ReadString()
	if err != nil {
		return
	}

	listener, found := thisTerminal.Listeners.Get(listenerId)
	if !found {
		return
	}

	thisTerminal.Listeners.Remove(listenerId)
	listener.(*terminal.Listener).Close()

	r := packet.Packet{}
	r.New(opcode.C2SDeviceDetachTerminal)
	r.WriteString(thisTerminal.Id)
	r.WriteString(listener.(*terminal.Listener).Id)
	app.Write(r)
}

func (app *App) S2CDeviceDeleteTerminalHandler(p packet.Packet) {

	terminalUId, err := p.ReadString()
	if err != nil {
		return
	}

	thisTerminal := app.GetTerminal(terminalUId)
	if thisTerminal == nil {
		return
	}

	app.DeleteTerminal(thisTerminal.Id)

	r := packet.Packet{}
	r.New(opcode.C2SDeviceDeleteTerminal)
	r.WriteString(thisTerminal.Id)
	app.Write(r)
}

func (app *App) S2CDeviceCreateTerminalHandler(p packet.Packet) {

	createdBy, err := p.ReadString()
	if err != nil {
		return
	}

	terminalId := ""
	// Create terminal id (check already exists?)
	for {
		terminalId = utils.RandomNumbers(TerminalUIdSize)
		if app.GetTerminal(terminalId) == nil {
			break
		}
	}

	newTerminal, cmd, err := terminal.Create(TerminalTagName, terminalId)
	if err != nil {
		return
	}

	go func() {
		time.Sleep(100 * time.Millisecond)

		utils.CmdExit(cmd)

		p := packet.Packet{}
		p.New(opcode.C2SDeviceCreateTerminal)
		p.WriteString(newTerminal.Id)
		p.WriteString(createdBy)
		app.Write(p)
	}()

	app.Terminals.Set(newTerminal.Id, newTerminal)
}

func (app *App) S2CDevicePingHandler() {

	// Get memory usage
	_, _, memoryUsedPercent := utils.GetMemoryUsage()

	// Get disk usage
	diskStatus := utils.GetDiskUsage()

	// Get cpu usage
	cpuUsage := utils.GetCpuUsage()

	// Response
	p := packet.Packet{}
	p.New(opcode.C2SDevicePing)
	p.WriteInteger(diskStatus.Used)
	p.WriteInteger(int(memoryUsedPercent * 100))
	p.WriteInteger(int(cpuUsage))
	app.Write(p)
}

func (app *App) S2CDeviceLoginHandler(p packet.Packet) {

	loginStatus, err := p.ReadByte()
	if err != nil {
		return
	}

	if loginStatus > 0 {
		if loginStatus == 4 {

			versionNumber, err := p.ReadString()
			if err != nil {
				cog.Print(cog.ERROR, "Login failed. Try again later.")
				app.Exit()
				return
			}

			app.Update(versionNumber)

		} else {
			cog.Print(cog.ERROR, "Login failed. Try again later.")
			app.Exit()
			return
		}
	}

	terminalUIds := utils.GetTerminals(TerminalTagName, TerminalUIdSize)

	r := packet.Packet{}
	r.New(opcode.C2SDeviceTerminals)
	r.WriteInteger(len(terminalUIds))

	for _, terminalUId := range terminalUIds {
		app.Terminals.Set(terminalUId, &terminal.Terminal{TagName: TerminalTagName, Id: terminalUId, Listeners: cmap.New()})
		r.WriteString(terminalUId)
	}

	// Send active terminals to server
	app.Write(r)
}

func (app *App) S2CDeviceRegisterHandler(p packet.Packet) {

	status, err := p.ReadByte()
	if err != nil {
		return
	}

	if status == 0 {
		// registration success
		tokenString, err := p.ReadString()
		if err != nil {
			return
		}

		// remove if exists
		os.Remove(app.Options.Path + "/cotunnel.key")

		f, err := os.Create(app.Options.Path + "/cotunnel.key")
		if err != nil {
			cog.Print(cog.ERROR, err.Error())
			cog.Print(cog.ERROR, "Registration failed. Try again later.")
			app.Exit()
			return
		}

		defer f.Close()

		_, err = f.Write([]byte(tokenString))
		if err != nil {
			cog.Print(cog.ERROR, err.Error())
			cog.Print(cog.ERROR, "Registration failed. Try again later.")
			app.Exit()
			return
		}

		if app.Options.Exit {
			cog.Print(cog.INFO, "Successfully registered.")
			cog.Print(cog.INFO, "You must start Cotunnel client again without --key command.")
			app.Exit()
			return
		} else {
			cog.Print(cog.INFO, "Successfully registered and working now.")
			app.Options.Token = tokenString
			app.Options.Key = ""
			app.SafeReconnect = true
			app.Conn.Close()
		}
	} else if status == 1 {
		cog.Print(cog.ERROR, "Registration failed. Try again later.")
		app.Exit()
	} else if status == 2 {
		cog.Print(cog.ERROR, "Invalid key. Try again later.")
		app.Exit()
	} else if status == 3 {
		cog.Print(cog.ERROR, "You can't add a free device to your account.")
		app.Exit()
	} else if status == 4 {
		cog.Print(cog.ERROR, "Insufficient funds. Please visit billing page.")
		app.Exit()
	}
}

func (app *App) S2CDeviceHandshakeOKHandler() {

	app.IsEncrypted = true

	if len(app.Options.Key) > 0 {
		// Register
		p := packet.Packet{}
		p.New(opcode.C2SDeviceRegister)
		p.WriteString(app.Options.Key)

		hostInfoStat, _ := host.Info()

		p.WriteString(hostInfoStat.Hostname)
		p.WriteString(hostInfoStat.OS)
		p.WriteString(runtime.GOARCH)
		p.WriteString(hostInfoStat.Platform)
		p.WriteString(hostInfoStat.PlatformVersion)

		_, totalMemoryTotal, _ := utils.GetMemoryUsage()
		p.WriteInteger(totalMemoryTotal)

		diskUsage := utils.GetDiskUsage()
		p.WriteInteger(diskUsage.All)

		app.Write(p)
	} else if len(app.Options.Token) > 0 {

		// Login
		p := packet.Packet{}
		p.New(opcode.C2SDeviceLogin)
		p.WriteString(Version)
		p.WriteString(app.Options.Token)
		app.Write(p)
	} else {
		os.Exit(1)
	}
}

func (app *App) S2CDeviceHandshakeHandler(p packet.Packet) {

	serverHandshakeBytes, err := p.ReadBytes()
	if err != nil {
		return
	}

	handshakeBytes, err := utils.DecryptAES(app.EncryptionKey, serverHandshakeBytes)
	if err != nil {
		return
	}

	clientHandshakeBytes, err := utils.EncryptAES(app.EncryptionKey, handshakeBytes)
	if err != nil {
		return
	}

	s := packet.Packet{}
	s.New(opcode.C2SDeviceHandshake)
	s.WriteBytes(clientHandshakeBytes)
	app.Write(s)
}

func (app *App) S2CDeviceSessionKeyHandler(p packet.Packet) {

	serverPubKeyBytes, err := p.ReadBytes()
	if err != nil {
		return
	}

	app.DHGroup, err = dhkx.GetGroup(0)
	if err != nil {
		return
	}

	app.DHPrivateKey, err = app.DHGroup.GeneratePrivateKey(nil)
	if err != nil {
		return
	}

	pub := app.DHPrivateKey.Bytes()

	serverPubKey := dhkx.NewPublicKey(serverPubKeyBytes)

	computeKey, err := app.DHGroup.ComputeKey(serverPubKey, app.DHPrivateKey)
	if err != nil {
		return
	}

	app.EncryptionKey = computeKey.Bytes()

	if len(app.EncryptionKey) > 32 {
		app.EncryptionKey = app.EncryptionKey[:32]
	}

	s := packet.Packet{}
	s.New(opcode.C2SDeviceSessionOK)
	s.WriteBytes(pub)
	app.Write(s)
}

func (app *App) S2CDeviceGatewayHandler(p packet.Packet) {

	status, err := p.ReadByte()
	if err != nil {
		return
	}

	if status == 1 {

	} else if status == 2 {
		s := packet.Packet{}
		s.New(opcode.C2SDeviceGateway)
		app.Write(s)
	}

}

func (app *App) GetTerminal(terminalUId string) *terminal.Terminal {
	thisTerminal, found := app.Terminals.Get(terminalUId)
	if !found {
		return nil
	}

	return thisTerminal.(*terminal.Terminal)
}

func (app *App) DeleteTerminal(terminalUId string) {

	thisTerminal, found := app.Terminals.Get(terminalUId)
	if !found {
		return
	}

	thisTerminal.(*terminal.Terminal).Close()
	app.Terminals.Remove(terminalUId)
}

func (app *App) Update(versionNumber string) {

	cog.Print(cog.INFO, "Updating the client...")

	binaryName := "cotunnel"

	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	updateUrl := fmt.Sprintf("https://s3.amazonaws.com/cotunnel/client/build/%s/%s/%s/%s", versionNumber, runtime.GOOS, runtime.GOARCH, binaryName)

	resp, err := http.Get(updateUrl)
	if err != nil {
		cog.Print(cog.ERROR, err.Error())
		cog.Print(cog.ERROR, "Update failed.")
		os.Exit(0)
		return
	}

	if resp.StatusCode != 200 {
		cog.Print(cog.ERROR, "Update failed.")
		os.Exit(0)
		return
	}

	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		cog.Print(cog.ERROR, "Update failed.")
		os.Exit(0)
		return
	}

	if err := utils.CmdRestart(); err != nil {
		cog.Print(cog.ERROR, "Restart failed. You should restart the app.")
		os.Exit(0)
		return
	}

	os.Exit(0)
}

func (app *App) Write(p packet.Packet) {

	app.WriteMutex.Lock()
	defer app.WriteMutex.Unlock()

	packetBytes := p.GetBytes()

	cog.PrintPacket(cog.DEBUG, 1, p.OpCode, packetBytes)

	if app.IsEncrypted {

		encryptedPacketBytes, err := utils.EncryptAES(app.EncryptionKey, packetBytes)
		if err != nil {
			return
		}

		var size = len(encryptedPacketBytes)

		var ret []byte
		ret = append(ret, (byte)((size>>8)&0xFF))
		ret = append(ret, (byte)(size&0xFF))
		ret = append(ret, (byte)((0>>8)&0xFF))
		ret = append(ret, (byte)(0&0xFF))

		for i := 0; i < len(encryptedPacketBytes); i++ {
			ret = append(ret, encryptedPacketBytes[i])
		}

		app.Conn.Write(ret)
	} else {
		app.Conn.Write(packetBytes)
	}
}

func (app *App) Exit() {
	os.Exit(0)
}
