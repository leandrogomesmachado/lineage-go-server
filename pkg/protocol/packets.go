package protocol

// Client Packets (recebidos do cliente)
const (
	RequestAuthLogin   = 0x00
	RequestServerLogin = 0x02
	RequestServerList  = 0x05
	AuthGameGuard      = 0x07
	RequestGGAuth      = 0x1B // GameGuard authentication
)

// Server Packets (enviados para o cliente)
const (
	Init          = 0x00
	LoginFail     = 0x01
	AccountKicked = 0x02
	LoginOk       = 0x03
	ServerList    = 0x04
	PlayFail      = 0x06
	PlayOk        = 0x07
)

// Login Fail Reasons
const (
	ReasonSystemError       = 0x01
	ReasonUserOrPassWrong   = 0x02
	ReasonUserOrPassWrong2  = 0x03
	ReasonAccessFailed      = 0x04
	ReasonAccountInUse      = 0x07
	ReasonServerOverloaded  = 0x0f
	ReasonServerMaintenance = 0x10
	ReasonTempPassExpired   = 0x11
	ReasonDualBox           = 0x23
)
