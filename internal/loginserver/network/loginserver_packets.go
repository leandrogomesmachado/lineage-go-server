package network

func montarInitLSPacket(chavePublica []byte) []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(0x00)
	escritor.escreverD(0x0101)
	escritor.escreverD(uint32(len(chavePublica)))
	escritor.escreverB(chavePublica)
	return escritor.bytes()
}

func montarAuthResponsePacket(serverID byte, nome string) []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(0x02)
	escritor.escreverC(serverID)
	escritor.escreverS(nome)
	return escritor.bytes()
}

func montarPlayerAuthResponsePacket(conta string, sucesso bool) []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(0x03)
	escritor.escreverS(conta)
	if sucesso {
		escritor.escreverC(1)
		return escritor.bytes()
	}
	escritor.escreverC(0)
	return escritor.bytes()
}

func montarLoginServerFailPacket(motivo byte) []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(0x01)
	escritor.escreverC(motivo)
	return escritor.bytes()
}
