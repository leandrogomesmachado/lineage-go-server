package network

type requestServerLoginPacket struct {
	loginOkID1 uint32
	loginOkID2 uint32
	serverID   byte
}

func lerRequestServerLoginPacket(dados []byte) *requestServerLoginPacket {
	leitor := novoLeitorPacketCliente(dados)
	return &requestServerLoginPacket{
		loginOkID1: leitor.lerD(),
		loginOkID2: leitor.lerD(),
		serverID:   leitor.lerC(),
	}
}
