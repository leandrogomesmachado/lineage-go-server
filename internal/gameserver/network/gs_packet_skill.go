package network

func montarMagicSkillUsePacket(casterObjID int32, targetObjID int32, skillID int32, skillLevel int32, hitTime int32, reuseDelay int32, cx int32, cy int32, cz int32, tx int32, ty int32, tz int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x48)
	escritor.escreverD(uint32(casterObjID))
	escritor.escreverD(uint32(targetObjID))
	escritor.escreverD(uint32(skillID))
	escritor.escreverD(uint32(skillLevel))
	escritor.escreverD(uint32(hitTime))
	escritor.escreverD(uint32(reuseDelay))
	escritor.escreverD(uint32(cx))
	escritor.escreverD(uint32(cy))
	escritor.escreverD(uint32(cz))
	escritor.escreverD(uint32(tx))
	escritor.escreverD(uint32(ty))
	escritor.escreverD(uint32(tz))
	return escritor.bytes()
}

func montarMagicSkillLaunchedPacket(casterObjID int32, skillID int32, skillLevel int32, alvosObjIDs []int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0xC4)
	escritor.escreverD(uint32(casterObjID))
	escritor.escreverD(uint32(skillID))
	escritor.escreverD(uint32(skillLevel))
	escritor.escreverD(uint32(len(alvosObjIDs)))
	for _, alvoID := range alvosObjIDs {
		escritor.escreverD(uint32(alvoID))
	}
	return escritor.bytes()
}

func montarMagicSkillCanceledPacket(casterObjID int32) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0xCA)
	escritor.escreverD(uint32(casterObjID))
	return escritor.bytes()
}
