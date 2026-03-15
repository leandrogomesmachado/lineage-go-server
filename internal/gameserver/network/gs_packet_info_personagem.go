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

func obterAugmentationPorItem(augmentacoes []gsdb.CharacterAugmentation, objectID int32) *gsdb.CharacterAugmentation {
	for i := range augmentacoes {
		item := &augmentacoes[i]
		if item.ItemOID != objectID {
			continue
		}
		return item
	}
	return nil
}

func obterAugmentationPorSlotPaperdoll(itens []gsdb.CharacterItem, augmentacoes []gsdb.CharacterAugmentation, slotPaperdoll int32) int32 {
	item := obterItemEquipadoPorSlot(itens, slotPaperdoll)
	if item == nil {
		return 0
	}
	augmentation := obterAugmentationPorItem(augmentacoes, item.ObjectID)
	if augmentation == nil {
		return 0
	}
	if augmentation.Attributes < 0 {
		return 0
	}
	return augmentation.Attributes
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

func montarCharInfoPacket(player *playerAtivo, itens []gsdb.CharacterItem, augmentacoes []gsdb.CharacterAugmentation, cubicIDs []int32) []byte {
	escritor := novoEscritorPacket()
	template, ok := obterTemplatePersonagemInicial(player.classID)
	if !ok {
		template = templatePersonagemInicial{classID: player.classID, race: player.race, str: 1, dex: 1, con: 1, intel: 1, wit: 1, men: 1, runSpd: 120, walkSpd: 80, swimSpd: 50, baseAtkSpd: 300, baseCrit: 4, pAtk: 1, pDef: 1, mAtk: 1, mDef: 1, radiusMasculino: 8, radiusFeminino: 8, heightMasculino: 23, heightFeminino: 23}
	}
	itensPapelBoneca := listarItensPapelBoneca(itens)
	statsCalculadas := calcularStatsPersonagem(template, player.nivel, itensPapelBoneca)
	augmentationMaoDireita := obterAugmentationPorSlotPaperdoll(itens, augmentacoes, 7)
	augmentationMaoEsquerda := obterAugmentationPorSlotPaperdoll(itens, augmentacoes, 8)
	radiusColisao, heightColisao := template.obterColisao(player.sexo)
	runSpd := statsCalculadas.runSpd
	walkSpd := statsCalculadas.walkSpd
	swimSpd := statsCalculadas.swimSpd
	hpMaximo := player.hpMaximo
	if hpMaximo <= 0 {
		hpMaximo = statsCalculadas.hpMaximo
	}
	mpMaximo := player.mpMaximo
	if mpMaximo <= 0 {
		mpMaximo = statsCalculadas.mpMaximo
	}
	cpMaximo := player.cpMaximo
	if cpMaximo <= 0 {
		cpMaximo = statsCalculadas.cpMaximo
	}
	hpAtual := player.hpAtual
	if hpAtual <= 0 || hpAtual > hpMaximo {
		hpAtual = hpMaximo
	}
	mpAtual := player.mpAtual
	if mpAtual <= 0 || mpAtual > mpMaximo {
		mpAtual = mpMaximo
	}
	cpAtual := player.cpAtual
	if cpAtual < 0 || cpAtual > cpMaximo {
		cpAtual = cpMaximo
	}
	movMultiplier := 1.0
	if template.runSpd > 0 {
		movMultiplier = float64(runSpd) / float64(template.runSpd)
	}
	atkMultiplier := 1.0
	basePAtkSpd := statsCalculadasBasePAtkSpd(template, itensPapelBoneca)
	if basePAtkSpd > 0 {
		atkMultiplier = float64(statsCalculadas.pAtkSpd) / float64(basePAtkSpd)
	}
	logger.Infof("CharInfo nome=%s classID=%d nivel=%d str=%d dex=%d con=%d int=%d wit=%d men=%d run=%d walk=%d swim=%d hpAtual=%d hpMax=%d mpAtual=%d mpMax=%d cpAtual=%d cpMax=%d", player.nome, player.classID, player.nivel, template.str, template.dex, template.con, template.intel, template.wit, template.men, runSpd, walkSpd, swimSpd, hpAtual, hpMaximo, mpAtual, mpMaximo, cpAtual, cpMaximo)
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
	escritor.escreverD(uint32(augmentationMaoDireita))
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
	escritor.escreverD(uint32(augmentationMaoEsquerda))
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverH(0)
	escritor.escreverD(uint32(player.pvpKills))
	escritor.escreverD(uint32(player.karma))
	escritor.escreverD(uint32(statsCalculadas.mAtkSpd))
	escritor.escreverD(uint32(statsCalculadas.pAtkSpd))
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
	escritor.escreverF(movMultiplier)
	escritor.escreverF(atkMultiplier)
	escritor.escreverF(radiusColisao)
	escritor.escreverF(heightColisao)
	escritor.escreverD(uint32(player.hairStyle))
	escritor.escreverD(uint32(player.hairColor))
	escritor.escreverD(uint32(player.face))
	escritor.escreverS(player.titulo)
	escritor.escreverD(uint32(player.clanID))
	escritor.escreverD(uint32(player.clanCrestID))
	escritor.escreverD(uint32(player.allyID))
	escritor.escreverD(uint32(player.allyCrestID))
	escritor.escreverD(0)
	escritor.escreverD(uint32(player.relation))
	escritor.escreverC(1)
	escritor.escreverC(1)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(uint8(player.mountType))
	escritor.escreverC(uint8(player.operateType))
	escritor.escreverC(0)
	escreverCubics(escritor, cubicIDs)
	escritor.escreverD(uint32(player.abnormalEffect))
	escritor.escreverC(uint8(player.recLeft))
	escritor.escreverH(uint16(player.recHave))
	escritor.escreverD(uint32(player.classID))
	escritor.escreverD(uint32(cpMaximo))
	escritor.escreverD(uint32(cpAtual))
	escritor.escreverC(0)
	escritor.escreverC(uint8(player.team))
	escritor.escreverD(uint32(player.clanCrestLargeID))
	escritor.escreverC(uint8(player.nobless))
	escritor.escreverC(uint8(player.hero))
	escritor.escreverC(uint8(player.fishing))
	escreverLoc(escritor, player.fishingX, player.fishingY, player.fishingZ)
	escritor.escreverD(uint32(player.nameColor))
	escritor.escreverD(uint32(player.heading))
	escritor.escreverD(uint32(player.pledgeClass))
	escritor.escreverD(uint32(player.pledgeType))
	escritor.escreverD(uint32(player.titleColor))
	escritor.escreverD(0)
	pacote := escritor.bytes()
	logger.Infof("CharInfo tamanho=%d hex=%s", len(pacote), resumirHexGameServer(pacote, 256))
	return pacote
}

func montarUserInfoPacket(slot gsdb.CharacterSlot, itens []gsdb.CharacterItem, augmentacoes []gsdb.CharacterAugmentation, cubicIDs []int32) []byte {
	escritor := novoEscritorPacket()
	template, ok := obterTemplatePersonagemInicial(slot.ClassID)
	if !ok {
		template = templatePersonagemInicial{classID: slot.ClassID, race: slot.Race, str: 1, dex: 1, con: 1, intel: 1, wit: 1, men: 1, runSpd: 120, walkSpd: 80, swimSpd: 50, baseAtkSpd: 300, baseCrit: 4, pAtk: 1, pDef: 1, mAtk: 1, mDef: 1, radiusMasculino: 8, radiusFeminino: 8, heightMasculino: 23, heightFeminino: 23}
	}
	itensPapelBoneca := listarItensPapelBoneca(itens)
	statsCalculadas := calcularStatsPersonagem(template, slot.Level, itensPapelBoneca)
	augmentationMaoDireita := obterAugmentationPorSlotPaperdoll(itens, augmentacoes, 7)
	augmentationMaoEsquerda := obterAugmentationPorSlotPaperdoll(itens, augmentacoes, 8)
	limiteInventario, _, _, _, _, _, _ := obterLimitesArmazenamento()
	radiusColisao, heightColisao := template.obterColisao(slot.Sex)
	runSpd := statsCalculadas.runSpd
	walkSpd := statsCalculadas.walkSpd
	swimSpd := statsCalculadas.swimSpd
	movMultiplier := 1.0
	if template.runSpd > 0 {
		movMultiplier = float64(runSpd) / float64(template.runSpd)
	}
	atkMultiplier := 1.0
	basePAtkSpd := statsCalculadasBasePAtkSpd(template, itensPapelBoneca)
	if basePAtkSpd > 0 {
		atkMultiplier = float64(statsCalculadas.pAtkSpd) / float64(basePAtkSpd)
	}
	classIDExibicao := slot.ClassID
	if slot.BaseClass > 0 && slot.BaseClass != slot.ClassID {
		classIDExibicao = slot.BaseClass
	}
	hpMaximo := slot.MaxHp
	if hpMaximo <= 0 {
		hpMaximo = statsCalculadas.hpMaximo
	}
	mpMaximo := slot.MaxMp
	if mpMaximo <= 0 {
		mpMaximo = statsCalculadas.mpMaximo
	}
	cpMaximo := slot.MaxCp
	if cpMaximo <= 0 {
		cpMaximo = statsCalculadas.cpMaximo
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
	if hpMaximo < 1 {
		hpMaximo = 1
	}
	if mpMaximo < 1 {
		mpMaximo = 1
	}
	if cpMaximo < 0 {
		cpMaximo = 0
	}
	pesoAtual := int32(1000)
	limitePeso := int32(10000)
	evasao := statsCalculadas.evasao
	precisao := statsCalculadas.precisao
	critico := statsCalculadas.critico
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
	escritor.escreverD(uint32(augmentationMaoDireita))
	escreverZerosH(escritor, 12)
	escritor.escreverD(uint32(augmentationMaoEsquerda))
	escreverZerosH(escritor, 4)
	escritor.escreverD(uint32(statsCalculadas.pAtk))
	escritor.escreverD(uint32(statsCalculadas.pAtkSpd))
	escritor.escreverD(uint32(statsCalculadas.pDef))
	escritor.escreverD(uint32(evasao))
	escritor.escreverD(uint32(precisao))
	escritor.escreverD(uint32(critico))
	escritor.escreverD(uint32(statsCalculadas.mAtk))
	escritor.escreverD(uint32(statsCalculadas.mAtkSpd))
	escritor.escreverD(uint32(statsCalculadas.pAtkSpd))
	escritor.escreverD(uint32(statsCalculadas.mDef))
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
	escritor.escreverF(movMultiplier)
	escritor.escreverF(atkMultiplier)
	escritor.escreverF(radiusColisao)
	escritor.escreverF(heightColisao)
	escritor.escreverD(uint32(slot.HairStyle))
	escritor.escreverD(uint32(slot.HairColor))
	escritor.escreverD(uint32(slot.Face))
	escritor.escreverD(0)
	escritor.escreverS(slot.Title)
	escritor.escreverD(uint32(slot.ClanID))
	escritor.escreverD(uint32(slot.ClanCrestID))
	escritor.escreverD(uint32(slot.AllyID))
	escritor.escreverD(uint32(slot.AllyCrestID))
	escritor.escreverD(uint32(slot.Relation))
	escritor.escreverC(uint8(slot.MountType))
	escritor.escreverC(uint8(slot.OperateType))
	escritor.escreverC(0)
	escritor.escreverD(uint32(slot.PkKills))
	escritor.escreverD(uint32(slot.PvpKills))
	escreverCubics(escritor, cubicIDs)
	escritor.escreverC(0)
	escritor.escreverD(uint32(slot.AbnormalEffect))
	escritor.escreverC(0)
	escritor.escreverD(uint32(slot.ClanPrivileges))
	escritor.escreverH(uint16(slot.RecLeft))
	escritor.escreverH(uint16(slot.RecHave))
	escritor.escreverD(uint32(slot.MountNpcID))
	escritor.escreverH(uint16(limiteInventario))
	escritor.escreverD(uint32(slot.ClassID))
	escritor.escreverD(0)
	escritor.escreverD(uint32(cpMaximo))
	escritor.escreverD(uint32(cpAtual))
	escritor.escreverC(0)
	escritor.escreverC(uint8(slot.Team))
	escritor.escreverD(uint32(slot.ClanCrestLargeID))
	escritor.escreverC(uint8(slot.Nobless))
	escritor.escreverC(uint8(slot.Hero))
	escritor.escreverC(uint8(slot.Fishing))
	escreverLoc(escritor, slot.FishingX, slot.FishingY, slot.FishingZ)
	escritor.escreverD(uint32(slot.NameColor))
	escritor.escreverC(uint8(boolParaByte(slot.Running)))
	escritor.escreverD(uint32(slot.PledgeClass))
	escritor.escreverD(uint32(slot.PledgeType))
	escritor.escreverD(uint32(slot.TitleColor))
	escritor.escreverD(0)
	pacote := escritor.bytes()
	logger.Infof("UserInfo tamanho=%d hex=%s", len(pacote), resumirHexGameServer(pacote, 256))
	return pacote
}

func listarItensPapelBoneca(itens []gsdb.CharacterItem) []itemPapelBoneca {
	resultado := make([]itemPapelBoneca, 0, len(itens))
	for _, item := range itens {
		if item.Loc != "PAPERDOLL" {
			continue
		}
		resultado = append(resultado, itemPapelBoneca{slotPaperdoll: item.LocData, itemID: item.ItemID})
	}
	return resultado
}

func statsCalculadasBasePAtkSpd(template templatePersonagemInicial, itens []itemPapelBoneca) int32 {
	base := obterBaseAtkSpdFisico(itens)
	if base > 0 {
		return base
	}
	if template.baseAtkSpd > 0 {
		return template.baseAtkSpd
	}
	return 300
}
