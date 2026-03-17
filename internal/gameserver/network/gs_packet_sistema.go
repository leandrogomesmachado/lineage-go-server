package network

const (
	msgIDYouDidS1Dano            int32 = 35
	msgIDAvoidedS1Attack         int32 = 42
	msgIDMissedTarget            int32 = 43
	msgIDCriticalHit             int32 = 44
	msgIDEarnedS1Experience      int32 = 45
	msgIDYouEarnedS1ExpS2Sp      int32 = 95
	msgIDYouIncreasedYourLevel   int32 = 96
	msgIDShieldDefenceSuccessful int32 = 111
)

const (
	statusAttrNivel = int32(1)
	statusAttrExp   = int32(2)
	statusAttrCurHp = int32(9)
	statusAttrMaxHp = int32(10)
	statusAttrCurMp = int32(11)
	statusAttrMaxMp = int32(12)
	statusAttrSp    = int32(13)
	statusAttrCurCp = int32(33)
	statusAttrMaxCp = int32(34)
)

func montarSystemMessageSimples(msgID int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x64)
	escritor.escreverD(uint32(msgID))
	escritor.escreverD(0)
	return escritor.bytes()
}

func montarSystemMessageNumero(msgID int32, valor int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x64)
	escritor.escreverD(uint32(msgID))
	escritor.escreverD(1)
	escritor.escreverD(1)
	escritor.escreverD(uint32(valor))
	return escritor.bytes()
}

func montarSystemMessageDoisNumeros(msgID int32, valor1 int32, valor2 int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x64)
	escritor.escreverD(uint32(msgID))
	escritor.escreverD(2)
	escritor.escreverD(1)
	escritor.escreverD(uint32(valor1))
	escritor.escreverD(1)
	escritor.escreverD(uint32(valor2))
	return escritor.bytes()
}

func montarSystemMessageNome(msgID int32, nome string) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x64)
	escritor.escreverD(uint32(msgID))
	escritor.escreverD(1)
	escritor.escreverD(0)
	escritor.escreverS(nome)
	return escritor.bytes()
}

func montarStatusUpdatePacket(objID int32, atributos [][2]int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x0e)
	escritor.escreverD(uint32(objID))
	escritor.escreverD(uint32(len(atributos)))
	for _, attr := range atributos {
		escritor.escreverD(uint32(attr[0]))
		escritor.escreverD(uint32(attr[1]))
	}
	return escritor.bytes()
}
