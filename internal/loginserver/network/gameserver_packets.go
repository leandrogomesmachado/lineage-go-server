package network

type gameServerAuthPacket struct {
	desiredID           byte
	aceitaIDAlternativo bool
	hostReservado       bool
	hostName            string
	porta               uint16
	maxPlayers          uint32
	hexID               []byte
}

func lerGameServerAuthPacket(dados []byte) *gameServerAuthPacket {
	leitor := novoLeitorPacketCliente(dados)
	desiredID := leitor.lerC()
	aceitaIDAlternativo := leitor.lerC() != 0
	hostReservado := leitor.lerC() != 0
	hostName := leitor.lerS()
	porta := leitor.lerH()
	maxPlayers := leitor.lerD()
	tamanhoHex := int(leitor.lerD())
	return &gameServerAuthPacket{
		desiredID:           desiredID,
		aceitaIDAlternativo: aceitaIDAlternativo,
		hostReservado:       hostReservado,
		hostName:            hostName,
		porta:               porta,
		maxPlayers:          maxPlayers,
		hexID:               leitor.lerB(tamanhoHex),
	}
}

type playerAuthRequestPacket struct {
	conta      string
	playOkID1  uint32
	playOkID2  uint32
	loginOkID1 uint32
	loginOkID2 uint32
}

func lerPlayerAuthRequestPacket(dados []byte) *playerAuthRequestPacket {
	leitor := novoLeitorPacketCliente(dados)
	return &playerAuthRequestPacket{
		conta:      leitor.lerS(),
		playOkID1:  leitor.lerD(),
		playOkID2:  leitor.lerD(),
		loginOkID1: leitor.lerD(),
		loginOkID2: leitor.lerD(),
	}
}

type playerInGamePacket struct {
	contas []string
}

func lerPlayerInGamePacket(dados []byte) *playerInGamePacket {
	leitor := novoLeitorPacketCliente(dados)
	quantidade := int(leitor.lerH())
	contas := make([]string, 0, quantidade)
	for i := 0; i < quantidade; i++ {
		contas = append(contas, leitor.lerS())
	}
	return &playerInGamePacket{contas: contas}
}

type playerLogoutPacket struct {
	conta string
}

func lerPlayerLogoutPacket(dados []byte) *playerLogoutPacket {
	leitor := novoLeitorPacketCliente(dados)
	return &playerLogoutPacket{conta: leitor.lerS()}
}
