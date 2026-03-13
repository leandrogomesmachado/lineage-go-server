package network

import (
	"context"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func (g *gameClient) processarRequestGameStart(packet *requestGameStartPacket) error {
	logger.Infof("RequestGameStart recebido para conta %s slot=%d", g.conta, packet.slot)
	slot, err := g.carregarSlotSelecionado(packet.slot)
	if err != nil {
		logger.Errorf("Erro ao carregar slot %d da conta %s: %v", packet.slot, g.conta, err)
		return g.enviarPacket(montarActionFailedPacket())
	}
	if slot == nil {
		logger.Warnf("Slot %d invalido para conta %s", packet.slot, g.conta)
		return g.enviarPacket(montarActionFailedPacket())
	}
	if slot.AccessLevel < 0 {
		logger.Warnf("Slot %d inacessivel para conta %s", packet.slot, g.conta)
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.personagemAtual = slot
	g.playerAtivo = nil
	g.slotSelecionado = packet.slot
	g.estado = estadoEntering
	if err = g.enviarPacket(montarSSQInfoPacket()); err != nil {
		return err
	}
	return g.enviarPacket(montarCharSelectedPacket(g.sessionKey.PlayOkID1, *slot))
}

func (g *gameClient) processarEnterWorld(packet *enterWorldPacket) error {
	_ = packet
	if g.personagemAtual == nil {
		logger.Warnf("EnterWorld recebido sem personagem selecionado para conta %s", g.conta)
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.garantirSpawnSeguro()
	g.playerAtivo = novoPlayerAtivo(g.conta, *g.personagemAtual)
	if err := g.server.characterRepo.AtualizarLastAccess(context.Background(), g.personagemAtual.ObjID); err != nil {
		logger.Warnf("Falha ao atualizar lastAccess do personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, err)
	}
	g.estado = estadoInGame
	g.server.mundo.registrar(g)
	logger.Infof("EnterWorld recebido para conta %s personagem=%s objID=%d", g.conta, g.personagemAtual.CharName, g.personagemAtual.ObjID)
	if err := g.enviarPacket(montarSkillListPacket()); err != nil {
		return err
	}
	if err := g.enviarPacket(montarUserInfoPacket(*g.personagemAtual)); err != nil {
		return err
	}
	if err := g.enviarPacket(montarItemListPacket()); err != nil {
		return err
	}
	if err := g.enviarPacket(montarShortCutInitPacket()); err != nil {
		return err
	}
	if err := g.enviarPacket(montarSkillCoolTimePacket()); err != nil {
		return err
	}
	if err := g.sincronizarVisibilidadeAoEntrarNoMundo(); err != nil {
		return err
	}
	return g.enviarPacket(montarActionFailedPacket())
}

func (g *gameClient) carregarSlotSelecionado(indice int32) (*gsdb.CharacterSlot, error) {
	slots, err := g.server.characterRepo.FindByAccount(context.Background(), g.conta)
	if err != nil {
		return nil, err
	}
	if indice < 0 {
		return nil, nil
	}
	if int(indice) >= len(slots) {
		return nil, nil
	}
	slot := slots[indice]
	return &slot, nil
}

func (g *gameClient) garantirSpawnSeguro() {
	if g.personagemAtual == nil {
		return
	}
	template, ok := obterTemplatePersonagemInicial(g.personagemAtual.ClassID)
	if !ok {
		return
	}
	spawnInicial := template.obterSpawnInicial(g.personagemAtual.ObjID)
	xAjustado, yAjustado, zAjustado := normalizarPosicaoMundo(g.personagemAtual.X, g.personagemAtual.Y, g.personagemAtual.Z)
	if !g.posicaoPareceSegura(xAjustado, yAjustado, zAjustado, template) {
		xAjustado = spawnInicial.x
		yAjustado = spawnInicial.y
		zAjustado = spawnInicial.z
	}
	g.personagemAtual.X = xAjustado
	g.personagemAtual.Y = yAjustado
	g.personagemAtual.Z = zAjustado
	err := g.server.characterRepo.AtualizarPosicao(context.Background(), g.personagemAtual.ObjID, xAjustado, yAjustado, zAjustado)
	if err != nil {
		logger.Warnf("Falha ao persistir spawn seguro para personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, err)
		return
	}
	logger.Infof("Spawn inicial validado para personagem %s objID=%d pos=(%d,%d,%d)", g.personagemAtual.CharName, g.personagemAtual.ObjID, xAjustado, yAjustado, zAjustado)
}

func (g *gameClient) posicaoPareceSegura(x int32, y int32, z int32, template templatePersonagemInicial) bool {
	if z <= -3200 {
		return false
	}
	if x == 0 && y == 0 {
		return false
	}
	if distancia3D(x, y, z, template.x, template.y, template.z) > 3000 {
		return false
	}
	return true
}

func (g *gameClient) sincronizarVisibilidadeAoEntrarNoMundo() error {
	if g.playerAtivo == nil {
		return nil
	}
	if err := g.enviarPacket(montarCharInfoPacket(g.playerAtivo)); err != nil {
		return err
	}
	visiveis := g.server.mundo.listarVisiveisPara(g)
	for _, outroCliente := range visiveis {
		if outroCliente == nil {
			continue
		}
		if outroCliente.playerAtivo == nil {
			continue
		}
		if err := g.enviarPacket(montarCharInfoPacket(outroCliente.playerAtivo)); err != nil {
			return err
		}
		err := outroCliente.enviarPacket(montarCharInfoPacket(g.playerAtivo))
		if err != nil {
			logger.Warnf("Falha ao enviar CharInfo de %s para %s: %v", g.playerAtivo.nome, outroCliente.playerAtivo.nome, err)
		}
	}
	return nil
}
