package network

import "github.com/leandrogomesmachado/l2raptors-go/pkg/protocol"

func (lc *LoginClient) montarInitPacket(moduloRSA []byte, chaveBlowfish []byte) []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(protocol.Init)
	escritor.escreverD(lc.sessionID)
	escritor.escreverD(0x0000c621)
	escritor.escreverB(moduloRSA)
	escritor.escreverB(make([]byte, 16))
	escritor.escreverB(chaveBlowfish)
	escritor.escreverC(0x00)
	return escritor.bytes()
}

func (lc *LoginClient) montarGGAuthPacket() []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(0x0B)
	escritor.escreverD(lc.sessionID)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	return escritor.bytes()
}

func (lc *LoginClient) montarLoginOkPacket() []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(protocol.LoginOk)
	escritor.escreverD(lc.sessionKey.LoginOkID1)
	escritor.escreverD(lc.sessionKey.LoginOkID2)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0x03ea)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverB(make([]byte, 16))
	return escritor.bytes()
}

func (lc *LoginClient) montarPlayOkPacket() []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(protocol.PlayOk)
	escritor.escreverD(lc.sessionKey.PlayOkID1)
	escritor.escreverD(lc.sessionKey.PlayOkID2)
	return escritor.bytes()
}

func montarLoginFailPacket(motivo byte) []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(protocol.LoginFail)
	escritor.escreverC(motivo)
	return escritor.bytes()
}

func montarPlayFailPacket(motivo byte) []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(protocol.PlayFail)
	escritor.escreverC(motivo)
	return escritor.bytes()
}

func montarServerListPacket() []byte {
	escritor := novoEscritorPacketServidor()
	escritor.escreverC(protocol.ServerList)
	escritor.escreverC(1)
	escritor.escreverC(0)
	escritor.escreverC(1)
	escritor.escreverB([]byte{127, 0, 0, 1})
	escritor.escreverD(7777)
	escritor.escreverC(0)
	escritor.escreverC(1)
	escritor.escreverH(0)
	escritor.escreverH(1000)
	escritor.escreverC(1)
	escritor.escreverD(0)
	escritor.escreverC(1)
	return escritor.bytes()
}
