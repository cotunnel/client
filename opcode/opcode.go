package opcode

const (
	S2CDeviceGateway        = 1000
	S2CDeviceSessionKey     = 1001
	S2CDeviceHandshake      = 1002
	S2CDeviceHandshakeOK    = 1003
	S2CDeviceRegister       = 1004
	S2CDeviceLogin          = 1005
	S2CDevicePing           = 1006
	S2CDeviceCreateTerminal = 1007
	S2CDeviceDeleteTerminal = 1008
	S2CDeviceAttachTerminal = 1009
	S2CDeviceWriteTerminal  = 1010
	S2CDeviceResizeTerminal = 1011
	S2CDeviceDetachTerminal = 1012
	S2CDeviceTunnel         = 1013

	C2SDeviceGateway        = 2000
	C2SDeviceSessionOK      = 2001
	C2SDeviceHandshake      = 2002
	C2SDeviceRegister       = 2003
	C2SDeviceLogin          = 2004
	C2SDevicePing           = 2005
	C2SDeviceCreateTerminal = 2006
	C2SDeviceDeleteTerminal = 2007
	C2SDeviceTerminals      = 2008
	C2SDeviceAttachTerminal = 2009
	C2SDeviceReadTerminal   = 2010
	C2SDeviceDetachTerminal = 2011
)
