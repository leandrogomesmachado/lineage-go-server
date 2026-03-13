package network

const (
	failReasonNoText                = 0
	failReasonSystemErrorLoginLater = 1
	motivoCriacaoFalhou             = 0
	motivoMuitosPersonagens         = 1
	motivoNomeJaExiste              = 2
	motivoNomeIncorreto             = 4
	motivoExclusaoFalhou            = 1
	motivoMembroDeClanNaoPode       = 2
	motivoLiderClanNaoPode          = 3
	versaoProtocoloInterlude1       = 737
	versaoProtocoloInterlude2       = 740
	versaoProtocoloInterlude3       = 744
	versaoProtocoloInterlude4       = 746
)

func escreverLoc(escritor *escritorPacket, x int32, y int32, z int32) {
	escritor.escreverD(uint32(x))
	escritor.escreverD(uint32(y))
	escritor.escreverD(uint32(z))
}

func escreverZerosD(escritor *escritorPacket, quantidade int) {
	for i := 0; i < quantidade; i++ {
		escritor.escreverD(0)
	}
}

func escreverZerosH(escritor *escritorPacket, quantidade int) {
	for i := 0; i < quantidade; i++ {
		escritor.escreverH(0)
	}
}

func escreverZerosC(escritor *escritorPacket, quantidade int) {
	for i := 0; i < quantidade; i++ {
		escritor.escreverC(0)
	}
}

func montarVersionCheckPacket(chave []byte) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x00)
	escritor.escreverC(0x01)
	escritor.escreverB(chave[:8])
	escritor.escreverD(0x01)
	escritor.escreverD(0x01)
	return escritor.bytes()
}
