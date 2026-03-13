package network

import (
	"context"
	"math"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func (g *gameClient) processarMoveBackwardToLocation(packet *moveBackwardToLocationPacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	logger.Infof("MoveBackwardToLocation recebido para conta %s personagem=%s origem=(%d,%d,%d) destino=(%d,%d,%d) tipo=%d", g.conta, g.playerAtivo.nome, packet.originX, packet.originY, packet.originZ, packet.targetX, packet.targetY, packet.targetZ, packet.tipoMovimento)
	if packet.tipoMovimento == 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if distancia3D(packet.originX, packet.originY, packet.originZ, packet.targetX, packet.targetY, packet.targetZ) > 9900 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	origemX := g.playerAtivo.x
	origemY := g.playerAtivo.y
	origemZ := g.playerAtivo.z
	destinoX, destinoY, destinoZ := corrigirPosicaoPorGeodataInicial(origemX, origemY, origemZ, packet.targetX, packet.targetY, packet.targetZ)
	g.playerAtivo.aplicarPosicao(destinoX, destinoY, destinoZ, calcularHeading(origemX, origemY, destinoX, destinoY))
	g.sincronizarPersonagemAtualComPlayerAtivo()
	g.persistirPosicaoPlayerAtivo()
	if err := g.enviarPacket(montarMoveToLocationPacketComOrigem(g.playerAtivo, destinoX, destinoY, destinoZ, origemX, origemY, origemZ)); err != nil {
		return err
	}
	g.broadcastPacoteParaVisiveis(montarMoveToLocationPacketComOrigem(g.playerAtivo, destinoX, destinoY, destinoZ, origemX, origemY, origemZ))
	return nil
}

func (g *gameClient) processarValidatePosition(packet *validatePositionPacket) error {
	if g.playerAtivo == nil {
		return nil
	}
	_ = packet.boatID
	logger.Infof("ValidatePosition recebido para conta %s personagem=%s posCliente=(%d,%d,%d) heading=%d", g.conta, g.playerAtivo.nome, packet.x, packet.y, packet.z, packet.heading)
	if distancia3D(g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z, packet.x, packet.y, packet.z) > desyncMaximoValidate {
		return g.enviarPacket(montarValidateLocationPacket(g.playerAtivo))
	}
	xAjustado, yAjustado, zAjustado := corrigirPosicaoPorGeodataInicial(g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z, packet.x, packet.y, packet.z)
	g.playerAtivo.aplicarPosicao(xAjustado, yAjustado, zAjustado, packet.heading)
	g.sincronizarPersonagemAtualComPlayerAtivo()
	g.persistirPosicaoPlayerAtivo()
	g.broadcastPacoteParaVisiveis(montarValidateLocationPacket(g.playerAtivo))
	return nil
}

func (g *gameClient) processarRequestRestart(packet *requestRestartPacket) error {
	_ = packet
	if g.playerAtivo == nil {
		return g.enviarPacket(montarRestartResponsePacket(false))
	}
	logger.Infof("RequestRestart recebido para conta %s personagem=%s", g.conta, g.playerAtivo.nome)
	if err := g.enviarPacket(montarRestartResponsePacket(true)); err != nil {
		return err
	}
	slots, err := g.server.characterRepo.FindByAccount(context.Background(), g.conta)
	if err != nil {
		return err
	}
	g.notificarRemocaoParaVisiveis()
	g.server.mundo.remover(g.playerAtivo.objID)
	g.playerAtivo = nil
	g.personagemAtual = nil
	g.slotSelecionado = 0
	g.estado = estadoAuthed
	return g.enviarPacket(montarCharSelectInfoPacket(g.conta, g.sessionKey.PlayOkID1, slots))
}

func (g *gameClient) processarLogout(packet *logoutPacket) error {
	_ = packet
	logger.Infof("Logout recebido para conta %s estado=%d", g.conta, g.estado)
	if g.playerAtivo != nil {
		g.notificarRemocaoParaVisiveis()
		g.server.mundo.remover(g.playerAtivo.objID)
	}
	g.playerAtivo = nil
	g.personagemAtual = nil
	g.server.removerCliente(g)
	return g.conn.Close()
}

func (g *gameClient) sincronizarPersonagemAtualComPlayerAtivo() {
	if g.personagemAtual == nil {
		return
	}
	if g.playerAtivo == nil {
		return
	}
	g.personagemAtual.X = g.playerAtivo.x
	g.personagemAtual.Y = g.playerAtivo.y
	g.personagemAtual.Z = g.playerAtivo.z
	g.personagemAtual.CurHp = g.playerAtivo.hpAtual
	g.personagemAtual.CurMp = g.playerAtivo.mpAtual
}

func (g *gameClient) persistirPosicaoPlayerAtivo() {
	if g.playerAtivo == nil {
		return
	}
	err := g.server.characterRepo.AtualizarPosicao(context.Background(), g.playerAtivo.objID, g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z)
	if err != nil {
		logger.Warnf("Falha ao persistir posicao do personagem %s objID=%d: %v", g.playerAtivo.nome, g.playerAtivo.objID, err)
	}
}

func (g *gameClient) broadcastPacoteParaVisiveis(pacote []byte) {
	if g.playerAtivo == nil {
		return
	}
	for _, outroCliente := range g.server.mundo.listarVisiveisPara(g) {
		if outroCliente == nil {
			continue
		}
		err := outroCliente.enviarPacket(pacote)
		if err != nil {
			logger.Warnf("Falha ao enviar pacote de visibilidade de %s: %v", g.playerAtivo.nome, err)
		}
	}
}

func (g *gameClient) notificarRemocaoParaVisiveis() {
	if g.playerAtivo == nil {
		return
	}
	g.broadcastPacoteParaVisiveis(montarDeleteObjectPacket(g.playerAtivo.objID))
}

func distancia3D(origemX int32, origemY int32, origemZ int32, destinoX int32, destinoY int32, destinoZ int32) float64 {
	deltaX := float64(destinoX - origemX)
	deltaY := float64(destinoY - origemY)
	deltaZ := float64(destinoZ - origemZ)
	return math.Sqrt(deltaX*deltaX + deltaY*deltaY + deltaZ*deltaZ)
}

func calcularHeading(origemX int32, origemY int32, destinoX int32, destinoY int32) int32 {
	angulo := math.Atan2(float64(destinoY-origemY), float64(destinoX-origemX))
	if angulo < 0 {
		angulo += 2 * math.Pi
	}
	return int32(angulo * 10430.378350470453)
}
