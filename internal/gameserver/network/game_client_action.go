package network

func (g *gameClient) processarAction(packet *actionPacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	_ = packet.originX
	_ = packet.originY
	_ = packet.originZ
	_ = packet.shiftPressed
	if g.trainerPessoal != nil && packet.objID == g.trainerPessoal.objID {
		g.playerAtivo.definirAlvo(g.trainerPessoal.objID)
		if err := g.enviarPacket(montarMyTargetSelectedPacket(g.trainerPessoal.objID, 0)); err != nil {
			return err
		}
		return g.enviarHtmlTrainer()
	}
	npcGlobal := g.server.mundo.obterNpcPorObjID(packet.objID)
	if npcGlobal != nil {
		g.playerAtivo.definirAlvo(npcGlobal.objID)
		if err := g.enviarPacket(montarMyTargetSelectedPacket(npcGlobal.objID, 0)); err != nil {
			return err
		}
		if !npcGlobal.ehMonster {
			return g.enviarPacket(montarActionFailedPacket())
		}
		return g.enviarPacket(montarActionFailedPacket())
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
	g.playerAtivo.definirAlvo(alvoCliente.playerAtivo.objID)
	if err := g.enviarPacket(montarMyTargetSelectedPacket(alvoCliente.playerAtivo.objID, 0)); err != nil {
		return err
	}
	pacoteAlvo := montarTargetSelectedPacket(g.playerAtivo.objID, alvoCliente.playerAtivo.objID, g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z)
	g.broadcastPacoteParaVisiveis(pacoteAlvo)
	return g.enviarPacket(montarActionFailedPacket())
}

func (g *gameClient) processarRequestActionUse(packet *requestActionUsePacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	_ = packet.ctrlPressed
	_ = packet.shiftPressed
	if packet.actionID == acaoProximoAlvo {
		return g.processarNextTarget()
	}
	if packet.actionID != acaoSentarLevantar {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.playerAtivo.estaSentado() {
		g.playerAtivo.levantar()
		pacote := montarChangeWaitTypePacket(g.playerAtivo, false)
		if err := g.enviarPacket(pacote); err != nil {
			return err
		}
		g.broadcastPacoteParaVisiveis(pacote)
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.playerAtivo.sentar()
	pacote := montarChangeWaitTypePacket(g.playerAtivo, true)
	if err := g.enviarPacket(pacote); err != nil {
		return err
	}
	g.broadcastPacoteParaVisiveis(pacote)
	return g.enviarPacket(montarActionFailedPacket())
}
