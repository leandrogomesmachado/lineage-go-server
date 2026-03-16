package network

func (g *gameClient) processarAction(packet *actionPacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.playerAtivo.removerProtecaoSpawn()
	_ = packet.originX
	_ = packet.originY
	_ = packet.originZ
	_ = packet.shiftPressed
	npcGlobal := g.server.mundo.obterNpcPorObjID(packet.objID)
	if npcGlobal != nil {
		if npcGlobal.ehMonster && g.playerAtivo.alvoObjID == npcGlobal.objID {
			return g.processarAttackRequest(&attackRequestPacket{objID: npcGlobal.objID, originX: packet.originX, originY: packet.originY, originZ: packet.originZ, shiftPressed: packet.shiftPressed})
		}
		g.playerAtivo.definirAlvo(npcGlobal.objID)
		if err := g.enviarPacket(montarMyTargetSelectedPacket(npcGlobal.objID, 0)); err != nil {
			return err
		}
		if !npcGlobal.ehMonster {
			return g.enviarHtmlNpcGlobal(npcGlobal)
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
	g.playerAtivo.removerProtecaoSpawn()
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
