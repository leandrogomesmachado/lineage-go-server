package network

func montarMoveToLocationPacket(player *playerAtivo, destinoX int32, destinoY int32, destinoZ int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x01)
	escritor.escreverD(uint32(player.objID))
	escreverLoc(escritor, destinoX, destinoY, destinoZ)
	escreverLoc(escritor, player.x, player.y, player.z)
	return escritor.bytes()
}

func montarMoveToLocationPacketComOrigem(player *playerAtivo, destinoX int32, destinoY int32, destinoZ int32, origemX int32, origemY int32, origemZ int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x01)
	escritor.escreverD(uint32(player.objID))
	escreverLoc(escritor, destinoX, destinoY, destinoZ)
	escreverLoc(escritor, origemX, origemY, origemZ)
	return escritor.bytes()
}

func montarValidateLocationPacket(player *playerAtivo) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x61)
	escritor.escreverD(uint32(player.objID))
	escreverLoc(escritor, player.x, player.y, player.z)
	escritor.escreverD(uint32(player.heading))
	return escritor.bytes()
}

func montarDeleteObjectPacket(objID int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x12)
	escritor.escreverD(uint32(objID))
	return escritor.bytes()
}

func montarRestartResponsePacket(sucesso bool) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x5f)
	if sucesso {
		escritor.escreverD(1)
		return escritor.bytes()
	}
	escritor.escreverD(0)
	return escritor.bytes()
}
