package network

import (
	"math"
	"time"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func (g *gameClient) processarRequestMagicSkillUse(packet *requestMagicSkillUsePacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.playerAtivo.removerProtecaoSpawn()
	if !g.personagemTemSkill(packet.skillID) {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if skillEhPassiva(packet.skillID, g.obterNivelSkill(packet.skillID)) {
		return g.enviarPacket(montarActionFailedPacket())
	}
	templateCubic, ok := obterTemplateSkillCubic(packet.skillID)
	if ok {
		if g.playerAtivo.cubicsEstaoCheios(g.obterNivelSkill(143)) {
			return g.enviarPacket(montarActionFailedPacket())
		}
		logger.Infof("RequestMagicSkillUse cubic conta=%s personagem=%s skillID=%d npcID=%d", g.conta, g.playerAtivo.nome, packet.skillID, templateCubic.npcID)
		g.playerAtivo.adicionarOuRenovarCubic(templateCubic.npcID)
		return g.reenviarEstadoVisualPersonagem()
	}
	templateAbnormal, okAbnormal := obterTemplateSkillAbnormal(packet.skillID)
	if okAbnormal {
		logger.Infof("RequestMagicSkillUse abnormal conta=%s personagem=%s skillID=%d mascara=0x%X", g.conta, g.playerAtivo.nome, packet.skillID, templateAbnormal.abnormalMask)
		g.playerAtivo.adicionarAbnormalEffect(templateAbnormal.abnormalMask)
		return g.reenviarEstadoVisualPersonagem()
	}
	templateFishing, okFishing := obterTemplateSkillFishing(packet.skillID)
	if okFishing {
		logger.Infof("RequestMagicSkillUse fishing conta=%s personagem=%s skillID=%d nome=%s", g.conta, g.playerAtivo.nome, packet.skillID, templateFishing.nome)
		if g.playerAtivo.estaPescando() {
			g.playerAtivo.pararPesca()
			return g.reenviarEstadoVisualPersonagem()
		}
		g.playerAtivo.iniciarPesca(g.playerAtivo.x, g.playerAtivo.y+200, g.playerAtivo.z)
		return g.reenviarEstadoVisualPersonagem()
	}
	tmplAtiva, okAtiva := obterTemplateSkillAtiva(packet.skillID)
	if okAtiva {
		return g.processarSkillAtiva(packet, tmplAtiva)
	}
	return g.enviarPacket(montarActionFailedPacket())
}

func (g *gameClient) processarSkillAtiva(packet *requestMagicSkillUsePacket, tmpl templateSkillAtiva) error {
	skillID := packet.skillID
	skillLevel := g.obterNivelSkill(skillID)
	agora := time.Now().UnixMilli()
	if g.cooldownsSkill != nil {
		if pronto, existe := g.cooldownsSkill[skillID]; existe && agora < pronto {
			return g.enviarPacket(montarActionFailedPacket())
		}
	}
	hitTime := tmpl.hitTime
	if hitTime <= 0 {
		hitTime = 600
	}
	reuseDelay := tmpl.reuseDelay
	if reuseDelay <= 0 {
		reuseDelay = 3000
	}
	tipoSkill := tmpl.skillType
	precisaAlvo := tipoSkill == "PDAM" || tipoSkill == "MDAM"
	alvoObjID := g.playerAtivo.alvoObjID
	if precisaAlvo && alvoObjID <= 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	npcAlvo := g.server.mundo.obterNpcPorObjID(alvoObjID)
	if precisaAlvo && npcAlvo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if precisaAlvo && npcAlvo != nil {
		castRange := tmpl.castRange
		if castRange <= 0 {
			castRange = 600
		}
		distAlvo := distancia3D(g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z, npcAlvo.x, npcAlvo.y, npcAlvo.z)
		if distAlvo > float64(castRange+100) {
			return g.enviarPacket(montarActionFailedPacket())
		}
	}
	if g.cooldownsSkill != nil {
		g.cooldownsSkill[skillID] = agora + int64(reuseDelay)
	}
	targetX, targetY, targetZ := g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z
	if npcAlvo != nil {
		targetX, targetY, targetZ = npcAlvo.x, npcAlvo.y, npcAlvo.z
	}
	pacoteUso := montarMagicSkillUsePacket(g.playerAtivo.objID, alvoObjID, skillID, int32(skillLevel), hitTime, reuseDelay, g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z, targetX, targetY, targetZ)
	if err := g.enviarPacket(pacoteUso); err != nil {
		return err
	}
	g.broadcastPacoteParaVisiveis(pacoteUso)
	logger.Infof("Skill ativa iniciada conta=%s personagem=%s skillID=%d tipo=%s alvo=%d hitTime=%dms", g.conta, g.playerAtivo.nome, skillID, tipoSkill, alvoObjID, hitTime)
	poder := tmpl.obterPoderPorNivel(skillLevel)
	go func() {
		time.Sleep(time.Duration(hitTime) * time.Millisecond)
		g.aplicarEfeitoSkillAtiva(skillID, int32(skillLevel), tipoSkill, poder, alvoObjID, npcAlvo)
	}()
	return nil
}

func (g *gameClient) aplicarEfeitoSkillAtiva(skillID int32, skillLevel int32, tipoSkill string, poder int32, alvoObjID int32, npcAlvo *npcGlobalRuntime) {
	if g.playerAtivo == nil {
		return
	}
	switch tipoSkill {
	case "PDAM", "MDAM":
		g.aplicarDanoSkill(skillID, skillLevel, tipoSkill, poder, alvoObjID, npcAlvo)
	case "HEAL":
		g.aplicarCuraSkill(skillID, skillLevel, poder)
	default:
		pacoteLancado := montarMagicSkillLaunchedPacket(g.playerAtivo.objID, skillID, skillLevel, []int32{g.playerAtivo.objID})
		_ = g.enviarPacket(pacoteLancado)
		g.broadcastPacoteParaVisiveis(pacoteLancado)
	}
}

func (g *gameClient) aplicarDanoSkill(skillID int32, skillLevel int32, tipoSkill string, poder int32, alvoObjID int32, npcAlvo *npcGlobalRuntime) {
	if npcAlvo == nil || !npcAlvo.estaVivo() {
		pacoteLancado := montarMagicSkillLaunchedPacket(g.playerAtivo.objID, skillID, skillLevel, []int32{})
		_ = g.enviarPacket(pacoteLancado)
		return
	}
	var dano int32
	if tipoSkill == "PDAM" {
		dano = int32(math.Round(float64(poder) * 0.77 * (0.9 + float64(geradorCombate.Intn(21))/100.0)))
	} else {
		dano = int32(math.Round(float64(poder) * (0.9 + float64(geradorCombate.Intn(21))/100.0)))
	}
	if dano < 1 {
		dano = 1
	}
	npcAlvo.registrarDanoRecebido(g.playerAtivo.objID, dano)
	morreu := npcAlvo.aplicarDano(dano)
	npcAlvo.registrarAggro(g.playerAtivo.objID, maximoInt32(dano, 1))
	npcAlvo.notificarEventoAi()
	pacoteLancado := montarMagicSkillLaunchedPacket(g.playerAtivo.objID, skillID, skillLevel, []int32{alvoObjID})
	_ = g.enviarPacket(pacoteLancado)
	g.broadcastPacoteParaVisiveis(pacoteLancado)
	for _, cliente := range g.server.mundo.listarPlayersVisiveisParaNpc(npcAlvo) {
		if cliente == nil {
			continue
		}
		_ = cliente.enviarPacket(pacoteLancado)
	}
	_ = g.enviarPacket(montarSystemMessageNumero(msgIDYouDidS1Dano, dano))
	if morreu || npcAlvo.deveBroadcastarStatusHp() {
		statusNpc := montarStatusUpdatePacket(npcAlvo.objID, [][2]int32{
			{statusAttrCurHp, npcAlvo.hpAtual},
			{statusAttrMaxHp, npcAlvo.hpMaximo},
		})
		_ = g.enviarPacket(statusNpc)
		for _, cliente := range g.server.mundo.listarPlayersVisiveisParaNpc(npcAlvo) {
			if cliente == nil {
				continue
			}
			_ = cliente.enviarPacket(statusNpc)
		}
	}
	if !morreu {
		return
	}
	g.server.processarMorteNpcGlobal(npcAlvo)
}

func (g *gameClient) aplicarCuraSkill(skillID int32, skillLevel int32, poder int32) {
	if g.playerAtivo == nil {
		return
	}
	cura := poder
	if cura < 1 {
		cura = 1
	}
	g.playerAtivo.hpAtual += cura
	if g.playerAtivo.hpAtual > g.playerAtivo.hpMaximo {
		g.playerAtivo.hpAtual = g.playerAtivo.hpMaximo
	}
	pacoteLancado := montarMagicSkillLaunchedPacket(g.playerAtivo.objID, skillID, skillLevel, []int32{g.playerAtivo.objID})
	_ = g.enviarPacket(pacoteLancado)
	g.broadcastPacoteParaVisiveis(pacoteLancado)
	statusUpdate := montarStatusUpdatePacket(g.playerAtivo.objID, [][2]int32{
		{statusAttrCurHp, g.playerAtivo.hpAtual},
		{statusAttrMaxHp, g.playerAtivo.hpMaximo},
	})
	_ = g.enviarPacket(statusUpdate)
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
