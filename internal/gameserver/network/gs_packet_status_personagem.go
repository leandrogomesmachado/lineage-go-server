package network

import (
	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func obterNivelClasse(classID int32) int32 {
	if classID == 0 || classID == 10 || classID == 18 || classID == 25 || classID == 31 || classID == 38 || classID == 44 || classID == 49 || classID == 53 {
		return 0
	}
	if classID >= 1 && classID <= 9 {
		return 1
	}
	if classID >= 11 && classID <= 17 {
		return 1
	}
	if classID >= 19 && classID <= 24 {
		return 1
	}
	if classID >= 26 && classID <= 30 {
		return 1
	}
	if classID >= 32 && classID <= 37 {
		return 1
	}
	if classID >= 39 && classID <= 43 {
		return 1
	}
	if classID >= 45 && classID <= 48 {
		return 1
	}
	if classID >= 50 && classID <= 52 {
		return 1
	}
	if classID >= 54 && classID <= 57 {
		return 1
	}
	return 2
}

func obterMaximoHennas(classID int32) int32 {
	nivelClasse := obterNivelClasse(classID)
	if nivelClasse < 1 {
		return 0
	}
	if nivelClasse == 1 {
		return 2
	}
	return 3
}

func obterLimitesArmazenamento() (int32, int32, int32, int32, int32, int32, int32) {
	limiteInventario := int32(100)
	limiteWarehouse := int32(120)
	limiteFreight := int32(20)
	limiteVendaPrivada := int32(8)
	limiteCompraPrivada := int32(8)
	limiteReceitaAnao := int32(50)
	limiteReceitaComum := int32(50)
	return limiteInventario, limiteWarehouse, limiteFreight, limiteVendaPrivada, limiteCompraPrivada, limiteReceitaAnao, limiteReceitaComum
}

func montarExStorageMaxCountPacket() []byte {
	escritor := novoEscritorPacket()
	limiteInventario, limiteWarehouse, limiteFreight, limiteVendaPrivada, limiteCompraPrivada, limiteReceitaAnao, limiteReceitaComum := obterLimitesArmazenamento()
	escritor.escreverC(0xfe)
	escritor.escreverH(0x2e)
	escritor.escreverD(uint32(limiteInventario))
	escritor.escreverD(uint32(limiteWarehouse))
	escritor.escreverD(uint32(limiteFreight))
	escritor.escreverD(uint32(limiteVendaPrivada))
	escritor.escreverD(uint32(limiteCompraPrivada))
	escritor.escreverD(uint32(limiteReceitaAnao))
	escritor.escreverD(uint32(limiteReceitaComum))
	pacote := escritor.bytes()
	logger.Infof("ExStorageMaxCount tamanho=%d hex=%s", len(pacote), resumirHexGameServer(pacote, 160))
	return pacote
}

func montarHennaInfoPacket(slot gsdb.CharacterSlot, hennas []gsdb.CharacterHenna) []byte {
	escritor := novoEscritorPacket()
	maximoHennas := obterMaximoHennas(slot.ClassID)
	escritor.escreverC(0xe4)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverD(uint32(maximoHennas))
	escritor.escreverD(uint32(len(hennas)))
	for _, henna := range hennas {
		escritor.escreverD(uint32(henna.SymbolID))
		escritor.escreverD(uint32(henna.SymbolID))
	}
	pacote := escritor.bytes()
	logger.Infof("HennaInfo tamanho=%d hex=%s", len(pacote), resumirHexGameServer(pacote, 160))
	return pacote
}

func montarEtcStatusUpdatePacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0xf3)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	pacote := escritor.bytes()
	logger.Infof("EtcStatusUpdate tamanho=%d hex=%s", len(pacote), resumirHexGameServer(pacote, 160))
	return pacote
}
