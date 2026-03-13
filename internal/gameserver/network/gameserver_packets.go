package network

import "github.com/leandrogomesmachado/l2raptors-go/pkg/protocol"

func boolParaByte(valor bool) byte {
	if valor {
		return 0x01
	}
	return 0x00
}

func montarBlowFishKeyPacket(chaveCriptografada []byte) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x00)
	escritor.escreverD(uint32(len(chaveCriptografada)))
	escritor.escreverB(chaveCriptografada)
	return escritor.bytes()
}

func montarAuthRequestPacket(serverID byte, aceitarIDAlternativo bool, hexID []byte, host string, porta uint16, reservarHost bool, maxPlayers uint32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x01)
	escritor.escreverC(serverID)
	escritor.escreverC(boolParaByte(aceitarIDAlternativo))
	escritor.escreverC(boolParaByte(reservarHost))
	escritor.escreverS(host)
	escritor.escreverH(porta)
	escritor.escreverD(maxPlayers)
	escritor.escreverD(uint32(len(hexID)))
	escritor.escreverB(hexID)
	return escritor.bytes()
}

func montarPlayerAuthRequestPacket(conta string, sessionKey *protocol.SessionKey) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x05)
	escritor.escreverS(conta)
	escritor.escreverD(sessionKey.PlayOkID1)
	escritor.escreverD(sessionKey.PlayOkID2)
	escritor.escreverD(sessionKey.LoginOkID1)
	escritor.escreverD(sessionKey.LoginOkID2)
	return escritor.bytes()
}

func montarPlayerInGamePacket(contas []string) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x02)
	escritor.escreverH(uint16(len(contas)))
	for _, conta := range contas {
		escritor.escreverS(conta)
	}
	return escritor.bytes()
}
