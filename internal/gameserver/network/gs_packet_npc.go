package network

import (
	"strconv"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
)

func montarNpcInfoPacket(npc *npcTrainerRuntime) []byte {
	if npc == nil {
		return montarActionFailedPacket()
	}
	escritor := novoEscritorPacket()
	escritor.escreverC(0x16)
	escritor.escreverD(uint32(npc.objID))
	escritor.escreverD(uint32(npc.npcID + 1000000))
	escritor.escreverD(0)
	escritor.escreverD(uint32(npc.x))
	escritor.escreverD(uint32(npc.y))
	escritor.escreverD(uint32(npc.z))
	escritor.escreverD(uint32(npc.heading))
	escritor.escreverD(0)
	escritor.escreverD(253)
	escritor.escreverD(333)
	escritor.escreverD(120)
	escritor.escreverD(40)
	escritor.escreverD(120)
	escritor.escreverD(40)
	escritor.escreverD(120)
	escritor.escreverD(40)
	escritor.escreverD(120)
	escritor.escreverD(40)
	escritor.escreverF(1)
	escritor.escreverF(1)
	escritor.escreverF(8)
	escritor.escreverF(23)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverC(1)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverC(2)
	escritor.escreverS(npc.nome)
	escritor.escreverS(npc.titulo)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverF(8)
	escritor.escreverF(23)
	escritor.escreverD(0)
	escritor.escreverD(0)
	return escritor.bytes()
}

func montarNpcGlobalInfoPacket(npc *npcGlobalRuntime) []byte {
	if npc == nil {
		return montarActionFailedPacket()
	}
	atacavel := uint32(0)
	if npc.canBeAttacked {
		atacavel = 1
	}
	abnormalEffect := uint32(0)
	if npc.alvoObjID > 0 {
		abnormalEffect = 0
	}
	movMultiplier := 1.0
	atkMultiplier := 1.0
	if npc.runSpd > 0 {
		movMultiplier = 1.0
	}
	if npc.pAtkSpd > 0 {
		atkMultiplier = 1.0
	}
	correndo := uint8(0)
	if npc.ehMonster && npc.canMove && distancia3D(npc.ultimoMoveX, npc.ultimoMoveY, npc.ultimoMoveZ, npc.x, npc.y, npc.z) > 0 {
		correndo = 1
	}
	emCombate := uint8(0)
	if npc.alvoObjID > 0 {
		emCombate = 1
	}
	escritor := novoEscritorPacket()
	escritor.escreverC(0x16)
	escritor.escreverD(uint32(npc.objID))
	idTemplate := npc.idTemplate
	if idTemplate <= 0 {
		idTemplate = npc.npcID
	}
	escritor.escreverD(uint32(idTemplate + 1000000))
	escritor.escreverD(atacavel)
	escritor.escreverD(uint32(npc.x))
	escritor.escreverD(uint32(npc.y))
	escritor.escreverD(uint32(npc.z))
	escritor.escreverD(uint32(npc.heading))
	escritor.escreverD(0)
	escritor.escreverD(uint32(npc.mAtkSpd))
	escritor.escreverD(uint32(npc.pAtkSpd))
	escritor.escreverD(uint32(npc.runSpd))
	escritor.escreverD(uint32(npc.walkSpd))
	escritor.escreverD(uint32(npc.runSpd))
	escritor.escreverD(uint32(npc.walkSpd))
	escritor.escreverD(uint32(npc.runSpd))
	escritor.escreverD(uint32(npc.walkSpd))
	escritor.escreverD(uint32(npc.runSpd))
	escritor.escreverD(uint32(npc.walkSpd))
	escritor.escreverF(movMultiplier)
	escritor.escreverF(atkMultiplier)
	escritor.escreverF(npc.radiusColisao)
	escritor.escreverF(npc.heightColisao)
	escritor.escreverD(uint32(npc.rHand))
	escritor.escreverD(0)
	escritor.escreverD(uint32(npc.lHand))
	escritor.escreverC(1)
	escritor.escreverC(correndo)
	escritor.escreverC(emCombate)
	escritor.escreverC(0)
	escritor.escreverC(2)
	escritor.escreverS(npc.nome)
	titulo := npc.titulo
	if npc.ehMonster {
		titulo = "Lv " + formatarNivelNpc(npc.nivel, npc.aggroRange, titulo)
	}
	escritor.escreverS(titulo)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(abnormalEffect)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverD(0)
	escritor.escreverC(0)
	escritor.escreverC(0)
	escritor.escreverF(npc.radiusColisao)
	escritor.escreverF(npc.heightColisao)
	escritor.escreverD(0)
	escritor.escreverD(0)
	return escritor.bytes()
}

func formatarNivelNpc(nivel int32, aggroRange int32, titulo string) string {
	prefixo := ""
	if nivel > 0 {
		prefixo = "" + strconv.FormatInt(int64(nivel), 10)
	}
	if aggroRange > 0 {
		prefixo += "*"
	}
	if titulo == "" {
		return prefixo
	}
	if prefixo == "" {
		return titulo
	}
	return prefixo + " " + titulo
}

func montarNpcHtmlMessagePacket(objID int32, html string) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x0f)
	escritor.escreverD(uint32(objID))
	escritor.escreverS(html)
	escritor.escreverD(0)
	return escritor.bytes()
}

func montarAcquireSkillListPacket(skills []gsdb.CharacterSkill, classID int32, nivel int32) []byte {
	_ = classID
	_ = nivel
	disponiveis := listarProximasSkillsAprendiveis(skills, classID, nivel)
	if len(disponiveis) == 0 {
		return montarAcquireSkillDonePacket()
	}
	escritor := novoEscritorPacket()
	escritor.escreverC(0x8a)
	escritor.escreverD(0)
	escritor.escreverD(uint32(len(disponiveis)))
	for _, skill := range disponiveis {
		escritor.escreverD(uint32(skill.skillID))
		escritor.escreverD(uint32(skill.skillLevel))
		escritor.escreverD(uint32(skill.skillLevel))
		escritor.escreverD(uint32(skill.cost))
		escritor.escreverD(0)
	}
	return escritor.bytes()
}

func montarAcquireSkillInfoPacket(skill templateSkillClasse) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x8b)
	escritor.escreverD(uint32(skill.skillID))
	escritor.escreverD(uint32(skill.skillLevel))
	escritor.escreverD(uint32(skill.cost))
	escritor.escreverD(0)
	escritor.escreverD(0)
	return escritor.bytes()
}

func montarAcquireSkillDonePacket() []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x8e)
	return escritor.bytes()
}
