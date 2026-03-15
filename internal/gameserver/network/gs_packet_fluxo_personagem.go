package network

import gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"

func montarAuthLoginFailPacket(motivo uint32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x14)
	escritor.escreverD(motivo)
	return escritor.bytes()
}

func montarNewCharacterSuccessPacket() []byte {
	escritor := novoEscritorPacket()
	templates := listarTemplatesPersonagemInicial()
	escritor.escreverC(0x17)
	escritor.escreverD(uint32(len(templates)))
	for _, template := range templates {
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

func consolidarSkillsParaSkillList(skills []gsdb.CharacterSkill) []gsdb.CharacterSkill {
	if len(skills) == 0 {
		return []gsdb.CharacterSkill{}
	}
	maioresPorSkillID := make(map[int32]gsdb.CharacterSkill, len(skills))
	for _, skill := range skills {
		atual, existe := maioresPorSkillID[skill.SkillID]
		if !existe {
			maioresPorSkillID[skill.SkillID] = skill
			continue
		}
		if skill.SkillLevel <= atual.SkillLevel {
			continue
		}
		maioresPorSkillID[skill.SkillID] = skill
	}
	resultado := make([]gsdb.CharacterSkill, 0, len(maioresPorSkillID))
	for _, skill := range maioresPorSkillID {
		resultado = append(resultado, skill)
	}
	ordenarSkillsAtivas(resultado)
	return resultado
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

func montarSkillListPacket(skills []gsdb.CharacterSkill) []byte {
	skillsConsolidadas := consolidarSkillsParaSkillList(skills)
	escritor := novoEscritorPacket()
	escritor.escreverC(0x58)
	escritor.escreverD(uint32(len(skillsConsolidadas)))
	for _, skill := range skillsConsolidadas {
		passiva := uint32(0)
		if skillEhPassiva(skill.SkillID, skill.SkillLevel) {
			passiva = 1
		}
		escritor.escreverD(passiva)
		escritor.escreverD(uint32(skill.SkillLevel))
		escritor.escreverD(uint32(skill.SkillID))
		escritor.escreverC(0)
	}
	return escritor.bytes()
}

func montarItemListPacketComJanela(itens []gsdb.CharacterItem, augmentacoes []gsdb.CharacterAugmentation, mostrarJanela bool) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x1b)
	if mostrarJanela {
		escritor.escreverH(1)
	}
	if !mostrarJanela {
		escritor.escreverH(0)
	}
	escritor.escreverH(uint16(len(itens)))
	for _, item := range itens {
		augmentationID := int32(0)
		augmentation := obterAugmentationPorItem(augmentacoes, item.ObjectID)
		if augmentation != nil && augmentation.Attributes >= 0 {
			augmentationID = augmentation.Attributes
		}
		escritor.escreverH(0)
		escritor.escreverD(uint32(item.ObjectID))
		escritor.escreverD(uint32(item.ItemID))
		escritor.escreverD(uint32(item.Count))
		escritor.escreverH(0)
		escritor.escreverH(uint16(item.CustomType1))
		equipado := uint16(0)
		if item.Loc == "PAPERDOLL" {
			equipado = 1
		}
		escritor.escreverH(equipado)
		escritor.escreverD(0)
		escritor.escreverH(uint16(item.EnchantLevel))
		escritor.escreverH(uint16(item.CustomType2))
		escritor.escreverD(uint32(augmentationID))
		escritor.escreverD(uint32(item.ManaLeft))
	}
	return escritor.bytes()
}

func montarItemListPacket(itens []gsdb.CharacterItem, augmentacoes []gsdb.CharacterAugmentation) []byte {
	return montarItemListPacketComJanela(itens, augmentacoes, false)
}

func montarShortCutInitPacket(atalhos []gsdb.CharacterShortcut, itens []gsdb.CharacterItem, augmentacoes []gsdb.CharacterAugmentation) []byte {
	_ = itens
	escritor := novoEscritorPacket()
	escritor.escreverC(0x45)
	escritor.escreverD(uint32(len(atalhos)))
	for _, atalho := range atalhos {
		tipoAtalho := uint32(2)
		if atalho.Type == "ITEM" {
			tipoAtalho = 1
		}
		if atalho.Type == "ACTION" {
			tipoAtalho = 3
		}
		escritor.escreverD(tipoAtalho)
		escritor.escreverD(uint32(atalho.Slot + (atalho.Page * 12)))
		if atalho.Type == "ITEM" {
			augmentationID := int32(0)
			augmentationItemObjectID := atalho.ID
			augmentation := obterAugmentationPorItem(augmentacoes, augmentationItemObjectID)
			if augmentation != nil && augmentation.Attributes >= 0 {
				augmentationID = augmentation.Attributes
			}
			escritor.escreverD(uint32(atalho.ID))
			escritor.escreverD(0)
			escritor.escreverD(0)
			escritor.escreverD(uint32(augmentationID))
			escritor.escreverD(0)
			escritor.escreverD(0)
			continue
		}
		if atalho.Type == "SKILL" {
			escritor.escreverD(uint32(atalho.ID))
			escritor.escreverD(uint32(atalho.Level))
			escritor.escreverC(0)
			escritor.escreverD(0)
			continue
		}
		escritor.escreverD(uint32(atalho.ID))
		escritor.escreverD(0)
	}
	return escritor.bytes()
}

func montarSkillCoolTimePacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0xc1)
	escritor.escreverD(0)
	return escritor.bytes()
}
