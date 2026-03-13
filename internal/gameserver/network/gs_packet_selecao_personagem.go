package network

import (
	"time"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
)

func montarCharSelectedPacket(sessionID uint32, slot gsdb.CharacterSlot) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x15)
	escritor.escreverS(slot.CharName)
	escritor.escreverD(uint32(slot.ObjID))
	escritor.escreverS(slot.Title)
	escritor.escreverD(sessionID)
	escritor.escreverD(uint32(slot.ClanID))
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.Sex))
	escritor.escreverD(uint32(slot.Race))
	escritor.escreverD(uint32(slot.ClassID))
	escritor.escreverD(1)
	escritor.escreverD(uint32(slot.X))
	escritor.escreverD(uint32(slot.Y))
	escritor.escreverD(uint32(slot.Z))
	escritor.escreverF(float64(slot.CurHp))
	escritor.escreverF(float64(slot.CurMp))
	escritor.escreverD(uint32(slot.Sp))
	escritor.escreverQ(uint64(slot.Exp))
	escritor.escreverD(uint32(slot.Level))
	escritor.escreverD(uint32(slot.Karma))
	escritor.escreverD(uint32(slot.PkKills))
	template, ok := obterTemplatePersonagemInicial(slot.ClassID)
	if !ok {
		template = templatePersonagemInicial{}
	}
	escritor.escreverD(uint32(template.intel))
	escritor.escreverD(uint32(template.str))
	escritor.escreverD(uint32(template.con))
	escritor.escreverD(uint32(template.men))
	escritor.escreverD(uint32(template.dex))
	escritor.escreverD(uint32(template.wit))
	for i := 0; i < 30; i++ {
		escritor.escreverD(0)
	}
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.ClassID))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	return escritor.bytes()
}

func montarCharSelectInfoPacket(loginName string, sessionID uint32, slots []gsdb.CharacterSlot) []byte {
	escritor := novoEscritorPacket()
	agora := time.Now().UnixMilli()
	escritor.escreverC(0x13)
	escritor.escreverD(uint32(len(slots)))
	for _, slot := range slots {
		escritor.escreverS(slot.CharName)
		escritor.escreverD(uint32(slot.ObjID))
		escritor.escreverS(loginName)
		escritor.escreverD(sessionID)
		escritor.escreverD(uint32(slot.ClanID))
		escritor.escreverD(0)
		escritor.escreverD(uint32(slot.Sex))
		escritor.escreverD(uint32(slot.Race))
		escritor.escreverD(uint32(slot.BaseClass))
		escritor.escreverD(1)
		escritor.escreverD(uint32(slot.X))
		escritor.escreverD(uint32(slot.Y))
		escritor.escreverD(uint32(slot.Z))
		escritor.escreverF(float64(slot.CurHp))
		escritor.escreverF(float64(slot.CurMp))
		escritor.escreverD(uint32(slot.Sp))
		escritor.escreverQ(uint64(slot.Exp))
		escritor.escreverD(uint32(slot.Level))
		escritor.escreverD(uint32(slot.Karma))
		escritor.escreverD(uint32(slot.PkKills))
		escritor.escreverD(uint32(slot.PvpKills))
		for i := 0; i < 7; i++ {
			escritor.escreverD(0)
		}
		for i := 0; i < 17; i++ {
			escritor.escreverD(0)
		}
		for i := 0; i < 17; i++ {
			escritor.escreverD(0)
		}
		escritor.escreverD(uint32(slot.HairStyle))
		escritor.escreverD(uint32(slot.HairColor))
		escritor.escreverD(uint32(slot.Face))
		escritor.escreverF(float64(slot.MaxHp))
		escritor.escreverF(float64(slot.MaxMp))
		tempoDelete := uint32(0)
		if slot.AccessLevel < 0 {
			tempoDelete = ^uint32(0)
		}
		if slot.AccessLevel >= 0 && slot.DeleteTime > agora {
			tempoDelete = uint32((slot.DeleteTime - agora) / 1000)
		}
		escritor.escreverD(tempoDelete)
		escritor.escreverD(uint32(slot.ClassID))
		escritor.escreverD(1)
		escritor.escreverC(0)
		escritor.escreverD(0)
	}
	return escritor.bytes()
}
