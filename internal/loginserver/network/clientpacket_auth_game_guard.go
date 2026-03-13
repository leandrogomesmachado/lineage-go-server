package network

type authGameGuardPacket struct {
	sessionID uint32
	data1     uint32
	data2     uint32
	data3     uint32
	data4     uint32
}

func lerAuthGameGuardPacket(dados []byte) *authGameGuardPacket {
	leitor := novoLeitorPacketCliente(dados)
	return &authGameGuardPacket{
		sessionID: leitor.lerD(),
		data1:     leitor.lerD(),
		data2:     leitor.lerD(),
		data3:     leitor.lerD(),
		data4:     leitor.lerD(),
	}
}
