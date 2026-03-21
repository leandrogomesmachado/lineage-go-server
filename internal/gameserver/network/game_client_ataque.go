package network

import "github.com/leandrogomesmachado/l2raptors-go/pkg/logger"

const distanciaAtaqueBasico = 150.0

func (g *gameClient) processarAttackRequest(packet *attackRequestPacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.playerAtivo.estaSentado() {
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.playerAtivo.removerProtecaoSpawn()
	_ = packet.originX
	_ = packet.originY
	_ = packet.originZ
	_ = packet.shiftPressed
	npcGlobal := g.server.mundo.obterNpcPorObjID(packet.objID)
	if npcGlobal != nil {
		if !npcGlobal.ehMonster {
			return g.enviarPacket(montarActionFailedPacket())
		}
		if !npcGlobal.canBeAttacked {
			return g.enviarPacket(montarActionFailedPacket())
		}
		if !npcGlobal.estaVivo() {
			return g.enviarPacket(montarActionFailedPacket())
		}
		if distancia3D(g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z, npcGlobal.x, npcGlobal.y, npcGlobal.z) > distanciaAtaqueBasico {
			return g.enviarPacket(montarActionFailedPacket())
		}
		g.playerAtivo.definirAlvo(npcGlobal.objID)
		g.iniciarAutoAtaqueNpc(npcGlobal.objID)
		return nil
	}
	alvoCliente := g.server.mundo.obterPorObjID(packet.objID)
	if alvoCliente == nil {
		g.playerAtivo.limparAlvo()
		return g.enviarPacket(montarActionFailedPacket())
	}
	if alvoCliente.playerAtivo == nil {
		g.playerAtivo.limparAlvo()
		return g.enviarPacket(montarActionFailedPacket())
	}
	if alvoCliente.playerAtivo.objID == g.playerAtivo.objID {
		g.playerAtivo.limparAlvo()
		return g.enviarPacket(montarActionFailedPacket())
	}
	if !posicaoNoRaioVisivel(g.playerAtivo, alvoCliente.playerAtivo) {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if distancia3D(g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z, alvoCliente.playerAtivo.x, alvoCliente.playerAtivo.y, alvoCliente.playerAtivo.z) > distanciaAtaqueBasico {
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.playerAtivo.definirAlvo(alvoCliente.playerAtivo.objID)
	dano := g.calcularDanoBasico()
	alvoCliente.aplicarDanoRuntime(dano)
	logger.Infof("AttackRequest recebido conta=%s atacante=%s alvo=%s dano=%d", g.conta, g.playerAtivo.nome, alvoCliente.playerAtivo.nome, dano)
	pacoteInicio := montarAutoAttackStartPacket(g.playerAtivo.objID)
	if err := g.enviarPacket(pacoteInicio); err != nil {
		return err
	}
	g.broadcastPacoteParaVisiveis(pacoteInicio)
	pacoteAtaque := montarAttackPacket(g.playerAtivo.objID, alvoCliente.playerAtivo.objID, dano, g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z)
	if err := g.enviarPacket(pacoteAtaque); err != nil {
		return err
	}
	g.broadcastPacoteParaVisiveis(pacoteAtaque)
	pacoteFim := montarAutoAttackStopPacket(g.playerAtivo.objID)
	if err := g.enviarPacket(pacoteFim); err != nil {
		return err
	}
	g.broadcastPacoteParaVisiveis(pacoteFim)
	if err := alvoCliente.enviarUserInfoAtualizado(); err != nil {
		return err
	}
	alvoCliente.broadcastCharInfoAtualizado()
	if alvoCliente.playerAtivo.hpAtual > 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	alvoCliente.playerAtivo.limparAlvo()
	if err := alvoCliente.enviarPacket(montarTargetUnselectedPacket(alvoCliente.playerAtivo.objID, alvoCliente.playerAtivo.x, alvoCliente.playerAtivo.y, alvoCliente.playerAtivo.z)); err != nil {
		return err
	}
	alvoCliente.broadcastPacoteParaVisiveis(montarTargetUnselectedPacket(alvoCliente.playerAtivo.objID, alvoCliente.playerAtivo.x, alvoCliente.playerAtivo.y, alvoCliente.playerAtivo.z))
	return g.enviarPacket(montarActionFailedPacket())
}

func (g *gameClient) enviarMensagensCombatePlayer(resultado resultadoAtaqueFisico) {
	if g == nil {
		return
	}
	if resultado.errou {
		_ = g.enviarPacket(montarSystemMessageSimples(msgIDMissedTarget))
		return
	}
	if resultado.critico {
		_ = g.enviarPacket(montarSystemMessageSimples(msgIDCriticalHit))
	}
	if resultado.defesaEscudo == "perfect" {
		return
	}
	_ = g.enviarPacket(montarSystemMessageNumero(msgIDYouDidS1Dano, resultado.dano))
}

func (g *gameClient) calcularDanoBasico() int32 {
	if g == nil {
		return 1
	}
	if g.playerAtivo == nil {
		return 1
	}
	dano := g.playerAtivo.nivel * 5
	if dano < 1 {
		return 1
	}
	return dano
}
