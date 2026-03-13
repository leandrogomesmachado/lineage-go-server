package network

type requestServerListPacket struct {
	loginOkID1 uint32
	loginOkID2 uint32
}

func lerRequestServerListPacket(dados []byte) *requestServerListPacket {
	leitor := novoLeitorPacketCliente(dados)
	return &requestServerListPacket{
		loginOkID1: leitor.lerD(),
		loginOkID2: leitor.lerD(),
	}
}
