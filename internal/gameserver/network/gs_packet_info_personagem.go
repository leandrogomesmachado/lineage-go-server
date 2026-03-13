package network

import (
	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func obterItemEquipadoPorSlot(itens []gsdb.CharacterItem, slotPaperdoll int32) *gsdb.CharacterItem {
	for i := range itens {
		item := &itens[i]
		if item.Loc != "PAPERDOLL" {
			continue
		}
		if item.LocData != slotPaperdoll {
			continue
		}
		return item
	}
	return nil
}

func escreverPaperdollObjectIDs(escritor *escritorPacket, itens []gsdb.CharacterItem) {
	slots := []int32{16, 2, 1, 3, 5, 4, 6, 7, 8, 9, 10, 11, 12, 13, 7, 15, 14}
	for _, slot := range slots {
		item := obterItemEquipadoPorSlot(itens, slot)
		if item == nil {
			escritor.escreverD(0)
			continue
		}
		escritor.escreverD(uint32(item.ObjectID))
	}
}

func escreverPaperdollItemIDsCharInfo(escritor *escritorPacket, itens []gsdb.CharacterItem) {
	slots := []int32{16, 6, 7, 8, 9, 10, 11, 12, 13, 7, 15, 14}
	for _, slot := range slots {
		item := obterItemEquipadoPorSlot(itens, slot)
		if item == nil {
			escritor.escreverD(0)
			continue
		}
		escritor.escreverD(uint32(item.ItemID))
	}
}

func escreverPaperdollItemIDs(escritor *escritorPacket, itens []gsdb.CharacterItem) {
	slots := []int32{16, 2, 1, 3, 5, 4, 6, 7, 8, 9, 10, 11, 12, 13, 7, 15, 14}
	for _, slot := range slots {
		item := obterItemEquipadoPorSlot(itens, slot)
		if item == nil {
			escritor.escreverD(0)
			continue
		}
		escritor.escreverD(uint32(item.ItemID))
	}
}

func montarCharInfoPacket(player *playerAtivo, itens []gsdb.CharacterItem) []byte {
	escritor := novoEscritorPacket()
	template, ok := obterTemplatePersonagemInicial(player.classID)
	if !ok {
		template = templatePersonagemInicial{}
	}
	radiusColisao, heightColisao := template.obterColisao(player.sexo)
	runSpd := template.runSpd
	walkSpd := template.walkSpd
	swimSpd := template.swimSpd
	logger.Infof("CharInfo nome=%s classID=%d nivel=%d str=%d dex=%d con=%d int=%d wit=%d men=%d run=%d walk=%d swim=%d hpAtual=%d hpMax=%d mpAtual=%d mpMax=%d cpAtual=%d cpMax=%d", player.nome, player.classID, player.nivel, template.str, template.dex, template.con, template.intel, template.wit, template.men, runSpd, walkSpd, swimSpd, player.hpAtual, player.hpMaximo, player.mpAtual, player.mpMaximo, player.cpAtual, player.cpMaximo)
	escritor.escreverC(0x03)
	escreverLoc(escritor, player.x, player.y, player.z)
	escritor.escreverD(0)
	escritor.escreverD(uint32(player.objID))
	escritor.escreverS(player.nome)
	escritor.escreverD(uint32(player.race))
	escritor.escreverD(uint32(player.sexo))
	escritor.escreverD(uint32(player.classID))
	escreverPaperdollItemIDsCharInfo(escritor, itens)
	escritor.escreverD(uint32(player.face))
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverD(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverD(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
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
	escritor.escreverD(uint32(player.classID))
	escritor.escreverD(uint32(player.cpMaximo))
	escritor.escreverD(uint32(player.cpAtual))
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverD(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escreverLoc(escritor, 0, 0, 0)
	escritor.escreverD(0)
	escritor.escreverD(uint32(player.heading))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	pacote := escritor.bytes()
	logger.Infof("CharInfo tamanho=%d hex=%s", len(pacote), resumirHexGameServer(pacote, 256))
	return pacote
}

func montarUserInfoPacket(slot gsdb.CharacterSlot, itens []gsdb.CharacterItem) []byte {
	escritor := novoEscritorPacket()
	template, ok := obterTemplatePersonagemInicial(slot.ClassID)
	if !ok {
		template = templatePersonagemInicial{}
	}
	limiteInventario, _, _, _, _, _, _ := obterLimitesArmazenamento()
	radiusColisao, heightColisao := template.obterColisao(slot.Sex)
	runSpd := template.runSpd
	walkSpd := template.walkSpd
	swimSpd := template.swimSpd
	classIDExibicao := slot.ClassID
	if slot.BaseClass > 0 && slot.BaseClass != slot.ClassID {
		classIDExibicao = slot.BaseClass
	}
	hpMaximo := slot.MaxHp
	if hpMaximo <= 0 {
		hpMaximo = template.obterHpMaximoPorNivel(slot.Level)
	}
	mpMaximo := slot.MaxMp
	if mpMaximo <= 0 {
		mpMaximo = template.obterMpMaximoPorNivel(slot.Level)
	}
	cpMaximo := slot.MaxCp
	if cpMaximo <= 0 {
		cpMaximo = template.obterCpMaximoPorNivel(slot.Level)
	}
	hpAtual := slot.CurHp
	if hpAtual <= 0 || hpAtual > hpMaximo {
		hpAtual = hpMaximo
	}
	mpAtual := slot.CurMp
	if mpAtual <= 0 || mpAtual > mpMaximo {
		mpAtual = mpMaximo
	}
	cpAtual := slot.CurCp
	if cpAtual < 0 || cpAtual > cpMaximo {
		cpAtual = cpMaximo
	}
	pesoAtual := int32(1000)
	limitePeso := int32(10000)
	evasao := int32(33)
	precisao := int32(33)
	critico := int32(4)
	if template.dex > 0 {
		evasao += template.dex / 10
		precisao += template.dex / 10
		critico += template.dex / 10
	}
	logger.Infof("UserInfo nome=%s classID=%d nivel=%d str=%d dex=%d con=%d int=%d wit=%d men=%d run=%d walk=%d swim=%d curHp=%d maxHp=%d curMp=%d maxMp=%d curCp=%d maxCp=%d", slot.CharName, classIDExibicao, slot.Level, template.str, template.dex, template.con, template.intel, template.wit, template.men, runSpd, walkSpd, swimSpd, hpAtual, hpMaximo, mpAtual, mpMaximo, cpAtual, cpMaximo)
	escritor.escreverC(0x04)
	escreverLoc(escritor, slot.X, slot.Y, slot.Z)
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.ObjID))
	escritor.escreverS(slot.CharName)
	escritor.escreverD(uint32(slot.Race))
	escritor.escreverD(uint32(slot.Sex))
	escritor.escreverD(uint32(classIDExibicao))
	escritor.escreverD(uint32(slot.Level))
	escritor.escreverQ(uint64(slot.Exp))
	escritor.escreverD(uint32(template.str))
	escritor.escreverD(uint32(template.dex))
	escritor.escreverD(uint32(template.con))
	escritor.escreverD(uint32(template.intel))
	escritor.escreverD(uint32(template.wit))
	escritor.escreverD(uint32(template.men))
	escritor.escreverD(uint32(hpMaximo))
	escritor.escreverD(uint32(hpAtual))
	escritor.escreverD(uint32(mpMaximo))
	escritor.escreverD(uint32(mpAtual))
	escritor.escreverD(uint32(slot.Sp))
	escritor.escreverD(uint32(pesoAtual))
	escritor.escreverD(uint32(limitePeso))
	escritor.escreverD(20)
	escreverPaperdollObjectIDs(escritor, itens)
	escreverPaperdollItemIDs(escritor, itens)
	escreverZerosH(escritor, 14)
	escritor.escreverD(0)
	escreverZerosH(escritor, 12)
	escritor.escreverD(0)
	escreverZerosH(escritor, 4)
	escritor.escreverD(uint32(template.pAtk))
	escritor.escreverD(uint32(template.pAtkSpd))
	escritor.escreverD(uint32(template.pDef))
	escritor.escreverD(uint32(evasao))
	escritor.escreverD(uint32(precisao))
	escritor.escreverD(uint32(critico))
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
	escritor.escreverF(1.0)
	escritor.escreverF(1.0)
	escritor.escreverF(radiusColisao)
	escritor.escreverF(heightColisao)
	escritor.escreverD(uint32(slot.HairStyle))
	escritor.escreverD(uint32(slot.HairColor))
	escritor.escreverD(uint32(slot.Face))
	escritor.escreverD(0)
	escritor.escreverS(slot.Title)
	escritor.escreverD(uint32(slot.ClanID))
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(uint32(slot.PkKills))
	escritor.escreverD(uint32(slot.PvpKills))
	escritor.escreverH(0)
	escritor.escreverC(0)
	escritor.escreverD(0)
	escritor.escreverC(0)
	escritor.escreverD(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverD(0)
	escritor.escreverH(uint16(limiteInventario))
	escritor.escreverD(uint32(slot.ClassID))
	escritor.escreverD(0)
	escritor.escreverD(uint32(cpMaximo))
	escritor.escreverD(uint32(cpAtual))
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverD(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escreverLoc(escritor, 0, 0, 0)
	escritor.escreverD(0)
	escritor.escreverC(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	pacote := escritor.bytes()
	logger.Infof("UserInfo tamanho=%d hex=%s", len(pacote), resumirHexGameServer(pacote, 256))
	return pacote
}
