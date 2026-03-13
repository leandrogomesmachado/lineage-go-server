package network

type requestAuthLoginPacket struct {
	dadosCriptografados []byte
}

func lerRequestAuthLoginPacket(dados []byte) *requestAuthLoginPacket {
	leitor := novoLeitorPacketCliente(dados)
	return &requestAuthLoginPacket{
		dadosCriptografados: leitor.lerB(128),
	}
}
