package network

import "github.com/leandrogomesmachado/l2raptors-go/pkg/logger"

func (g *gameClient) processarRequestMagicSkillUse(packet *requestMagicSkillUsePacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if !g.personagemTemSkill(packet.skillID) {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if skillEhPassiva(packet.skillID, g.obterNivelSkill(packet.skillID)) {
		return g.enviarPacket(montarActionFailedPacket())
	}
	template, ok := obterTemplateSkillCubic(packet.skillID)
	if ok {
		if g.playerAtivo.cubicsEstaoCheios(g.obterNivelSkill(143)) {
			return g.enviarPacket(montarActionFailedPacket())
		}
		logger.Infof("RequestMagicSkillUse recebido para cubic conta=%s personagem=%s skillID=%d npcID=%d", g.conta, g.playerAtivo.nome, packet.skillID, template.npcID)
		g.playerAtivo.adicionarOuRenovarCubic(template.npcID)
		return g.reenviarEstadoVisualPersonagem()
	}
	templateAbnormal, okAbnormal := obterTemplateSkillAbnormal(packet.skillID)
	if okAbnormal {
		logger.Infof("RequestMagicSkillUse recebido para abnormalEffect conta=%s personagem=%s skillID=%d mascara=0x%X", g.conta, g.playerAtivo.nome, packet.skillID, templateAbnormal.abnormalMask)
		g.playerAtivo.adicionarAbnormalEffect(templateAbnormal.abnormalMask)
		return g.reenviarEstadoVisualPersonagem()
	}
	templateFishing, okFishing := obterTemplateSkillFishing(packet.skillID)
	if okFishing {
		logger.Infof("RequestMagicSkillUse recebido para fishing conta=%s personagem=%s skillID=%d nome=%s", g.conta, g.playerAtivo.nome, packet.skillID, templateFishing.nome)
		if g.playerAtivo.estaPescando() {
			g.playerAtivo.pararPesca()
			return g.reenviarEstadoVisualPersonagem()
		}
		g.playerAtivo.iniciarPesca(g.playerAtivo.x, g.playerAtivo.y+200, g.playerAtivo.z)
		return g.reenviarEstadoVisualPersonagem()
	}
	return g.enviarPacket(montarActionFailedPacket())
}

func (g *gameClient) personagemTemSkill(skillID int32) bool {
	for _, skill := range g.skillsAtivas {
		if skill.SkillID != skillID {
			continue
		}
		return true
	}
	return false
}

func (g *gameClient) obterNivelSkill(skillID int32) int32 {
	for _, skill := range g.skillsAtivas {
		if skill.SkillID != skillID {
			continue
		}
		return skill.SkillLevel
	}
	return 0
}

func (g *gameClient) reenviarEstadoVisualPersonagem() error {
	if g.playerAtivo == nil {
		return nil
	}
	if g.personagemAtual == nil {
		return nil
	}
	err := g.enviarUserInfoAtualizado()
	if err != nil {
		return err
	}
	g.broadcastCharInfoAtualizado()
	return nil
}
