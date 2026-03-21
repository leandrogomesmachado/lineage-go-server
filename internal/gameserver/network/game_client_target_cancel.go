package network

func (g *gameClient) processarRequestTargetCancel(packet *requestTargetCancelPacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	_ = packet.unselect
	g.pararAutoAtaque()
	if g.playerAtivo.alvoObjID == 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.playerAtivo.limparAlvo()
	pacote := montarTargetUnselectedPacket(g.playerAtivo.objID, g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z)
	if err := g.enviarPacket(pacote); err != nil {
		return err
	}
	g.broadcastPacoteParaVisiveis(pacote)
	return g.enviarPacket(montarActionFailedPacket())
}
