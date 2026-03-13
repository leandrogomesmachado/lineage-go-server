package network

import (
	"time"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
)

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

func montarVersionCheckPacket(chave []byte) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x00)
	escritor.escreverC(0x01)
	escritor.escreverB(chave[:8])
	escritor.escreverD(0x01)
	escritor.escreverD(0x01)
	return escritor.bytes()
}

func montarCharInfoPacket(player *playerAtivo) []byte {
	escritor := novoEscritorPacket()
	template, ok := obterTemplatePersonagemInicial(player.classID)
	if !ok {
		template = templatePersonagemInicial{}
	}
	radiusColisao, heightColisao := template.obterColisao(player.sexo)
	runSpd := template.runSpd
	walkSpd := template.walkSpd
	swimSpd := template.swimSpd
	escritor.escreverC(0x03)
	escreverLoc(escritor, player.x, player.y, player.z)
	escritor.escreverD(0)
	escritor.escreverD(uint32(player.objID))
	escritor.escreverS(player.nome)
	escritor.escreverD(uint32(player.race))
	escritor.escreverD(uint32(player.sexo))
	escritor.escreverD(uint32(player.classID))
	escreverZerosD(escritor, 12)
	escreverZerosH(escritor, 4)
	escritor.escreverD(0)
	escreverZerosH(escritor, 12)
	escritor.escreverD(0)
	escreverZerosH(escritor, 4)
	escritor.escreverD(uint32(player.pvpKills))
	escritor.escreverD(uint32(player.karma))
	escritor.escreverD(uint32(template.mAtkSpd))
	escritor.escreverD(uint32(template.pAtkSpd))
	escritor.escreverD(uint32(player.pvpKills))
	escritor.escreverD(uint32(player.karma))
	escritor.escreverD(uint32(runSpd))
	escritor.escreverD(uint32(walkSpd))
	escritor.escreverD(uint32(swimSpd))
	escritor.escreverD(uint32(swimSpd))
	escritor.escreverD(uint32(runSpd))
	escritor.escreverD(uint32(walkSpd))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverF(1.0)
	escritor.escreverF(1.0)
	escritor.escreverF(radiusColisao)
	escritor.escreverF(heightColisao)
	escritor.escreverD(uint32(player.hairStyle))
	escritor.escreverD(uint32(player.hairColor))
	escritor.escreverD(uint32(player.face))
	escritor.escreverS(player.titulo)
	escritor.escreverD(uint32(player.clanID))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverC(1)
	escritor.escreverC(1)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverH(0)
	escritor.escreverD(0)
	escritor.escreverC(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverD(uint32(player.classID))
	escritor.escreverD(uint32(player.cpMaximo))
	escritor.escreverD(uint32(player.cpAtual))
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escreverLoc(escritor, 0, 0, 0)
	escritor.escreverD(0)
	escritor.escreverD(uint32(player.heading))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	_ = template
	return escritor.bytes()
}

func montarAuthLoginFailPacket(motivo uint32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x14)
	escritor.escreverD(motivo)
	return escritor.bytes()
}

func montarNewCharacterSuccessPacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x17)
	escritor.escreverD(uint32(len(templatesPersonagemInicial)))
	for _, template := range templatesPersonagemInicial {
		escritor.escreverD(uint32(template.race))
		escritor.escreverD(uint32(template.classID))
		escritor.escreverD(0x46)
		escritor.escreverD(uint32(template.str))
		escritor.escreverD(0x0a)
		escritor.escreverD(0x46)
		escritor.escreverD(uint32(template.dex))
		escritor.escreverD(0x0a)
		escritor.escreverD(0x46)
		escritor.escreverD(uint32(template.con))
		escritor.escreverD(0x0a)
		escritor.escreverD(0x46)
		escritor.escreverD(uint32(template.intel))
		escritor.escreverD(0x0a)
		escritor.escreverD(0x46)
		escritor.escreverD(uint32(template.wit))
		escritor.escreverD(0x0a)
		escritor.escreverD(0x46)
		escritor.escreverD(uint32(template.men))
		escritor.escreverD(0x0a)
	}
	return escritor.bytes()
}

func montarCharCreateOkPacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x19)
	escritor.escreverD(0x01)
	return escritor.bytes()
}

func montarCharCreateFailPacket(motivo uint32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x1a)
	escritor.escreverD(motivo)
	return escritor.bytes()
}

func montarCharDeleteOkPacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x23)
	return escritor.bytes()
}

func montarCharDeleteFailPacket(motivo uint32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x24)
	escritor.escreverD(motivo)
	return escritor.bytes()
}

func montarSSQInfoPacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0xf8)
	escritor.escreverH(256)
	return escritor.bytes()
}

func montarActionFailedPacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x25)
	return escritor.bytes()
}

func montarSkillListPacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x58)
	escritor.escreverD(0)
	return escritor.bytes()
}

func montarItemListPacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x1b)
	escritor.escreverH(0)
	escritor.escreverH(0)
	return escritor.bytes()
}

func montarShortCutInitPacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x45)
	escritor.escreverD(0)
	return escritor.bytes()
}

func montarSkillCoolTimePacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0xc1)
	escritor.escreverD(0)
	return escritor.bytes()
}

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

func montarUserInfoPacket(slot gsdb.CharacterSlot) []byte {
	escritor := novoEscritorPacket()
	template, ok := obterTemplatePersonagemInicial(slot.ClassID)
	if !ok {
		template = templatePersonagemInicial{}
	}
	radiusColisao, heightColisao := template.obterColisao(slot.Sex)
	runSpd := template.runSpd
	walkSpd := template.walkSpd
	swimSpd := template.swimSpd
	escritor.escreverC(0x04)
	escritor.escreverD(uint32(slot.X))
	escritor.escreverD(uint32(slot.Y))
	escritor.escreverD(uint32(slot.Z))
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.ObjID))
	escritor.escreverS(slot.CharName)
	escritor.escreverD(uint32(slot.Race))
	escritor.escreverD(uint32(slot.Sex))
	escritor.escreverD(uint32(slot.ClassID))
	escritor.escreverD(uint32(slot.Level))
	escritor.escreverQ(uint64(slot.Exp))
	escritor.escreverD(uint32(template.str))
	escritor.escreverD(uint32(template.dex))
	escritor.escreverD(uint32(template.con))
	escritor.escreverD(uint32(template.intel))
	escritor.escreverD(uint32(template.wit))
	escritor.escreverD(uint32(template.men))
	escritor.escreverD(uint32(slot.MaxHp))
	escritor.escreverD(uint32(slot.CurHp))
	escritor.escreverD(uint32(slot.MaxMp))
	escritor.escreverD(uint32(slot.CurMp))
	escritor.escreverD(uint32(slot.Sp))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(20)
	escreverZerosD(escritor, 17)
	escreverZerosD(escritor, 17)
	escreverZerosH(escritor, 14)
	escritor.escreverD(0)
	escreverZerosH(escritor, 12)
	escritor.escreverD(0)
	escreverZerosH(escritor, 4)
	escritor.escreverD(uint32(template.pAtk))
	escritor.escreverD(uint32(template.pAtkSpd))
	escritor.escreverD(uint32(template.pDef))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(uint32(template.mAtk))
	escritor.escreverD(uint32(template.mAtkSpd))
	escritor.escreverD(uint32(template.pAtkSpd))
	escritor.escreverD(uint32(template.mDef))
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.Karma))
	escritor.escreverD(uint32(runSpd))
	escritor.escreverD(uint32(walkSpd))
	escritor.escreverD(uint32(swimSpd))
	escritor.escreverD(uint32(swimSpd))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverF(1.0)
	escritor.escreverF(1.0)
	escritor.escreverF(radiusColisao)
	escritor.escreverF(heightColisao)
	escritor.escreverD(uint32(slot.HairStyle))
	escritor.escreverD(uint32(slot.HairColor))
	escritor.escreverD(uint32(slot.Face))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.PkKills))
	escritor.escreverD(uint32(slot.PvpKills))
	escritor.escreverH(0)
	escritor.escreverC(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverC(0)
	escritor.escreverD(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.ClassID))
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.MaxCp))
	escritor.escreverD(uint32(slot.CurCp))
	escritor.escreverC(0)
	escritor.escreverD(uint32(slot.MaxCp))
	escritor.escreverD(uint32(slot.CurCp))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escreverLoc(escritor, 0, 0, 0)
	escritor.escreverD(0)
	escritor.escreverC(1)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	return escritor.bytes()
}

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
