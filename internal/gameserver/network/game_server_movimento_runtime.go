package network

import (
	"context"
	"math"
	"time"
)

const intervaloMovimentoRuntime = 100 * time.Millisecond

func (g *gameServer) loopMovimentoRuntime() {
	if g == nil {
		return
	}
	ticker := time.NewTicker(intervaloMovimentoRuntime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			g.processarTickMovimentoRuntime()
		case <-g.canalParada:
			return
		}
	}
}

func (g *gameServer) processarTickMovimentoRuntime() {
	if g == nil {
		return
	}
	if g.mundo == nil {
		return
	}
	clientes := g.mundo.listarPlayersAtivos()
	for _, cliente := range clientes {
		if cliente == nil {
			continue
		}
		cliente.processarTickMovimentoRuntime(intervaloMovimentoRuntime)
	}
	g.processarTickNpcGlobal(intervaloMovimentoRuntime)
}

func (g *gameClient) processarTickMovimentoRuntime(intervalo time.Duration) {
	if g == nil {
		return
	}
	if g.playerAtivo == nil {
		return
	}
	if !g.playerAtivo.estaMovendo() {
		return
	}
	destinoX := g.playerAtivo.destinoX
	destinoY := g.playerAtivo.destinoY
	destinoZ := g.playerAtivo.destinoZ
	origemX := g.playerAtivo.x
	origemY := g.playerAtivo.y
	origemZ := g.playerAtivo.z
	distancia := distancia3D(origemX, origemY, origemZ, destinoX, destinoY, destinoZ)
	if distancia <= 1 {
		g.playerAtivo.aplicarPosicao(destinoX, destinoY, destinoZ, g.playerAtivo.heading)
		g.playerAtivo.pararMovimento()
		g.sincronizarPersonagemAtualComPlayerAtivo()
		g.playerAtivo.ultimoMoveX = g.playerAtivo.x
		g.playerAtivo.ultimoMoveY = g.playerAtivo.y
		g.playerAtivo.ultimoMoveZ = g.playerAtivo.z
		g.broadcastPacoteParaVisiveis(montarValidateLocationPacket(g.playerAtivo))
		return
	}
	velocidade := g.obterVelocidadeMovimentoRuntime()
	if velocidade <= 0 {
		velocidade = 120
	}
	passo := velocidade * intervalo.Seconds()
	if passo >= distancia {
		g.playerAtivo.aplicarPosicao(destinoX, destinoY, destinoZ, g.playerAtivo.heading)
		g.playerAtivo.pararMovimento()
		g.sincronizarPersonagemAtualComPlayerAtivo()
		g.playerAtivo.ultimoMoveX = g.playerAtivo.x
		g.playerAtivo.ultimoMoveY = g.playerAtivo.y
		g.playerAtivo.ultimoMoveZ = g.playerAtivo.z
		g.broadcastPacoteParaVisiveis(montarValidateLocationPacket(g.playerAtivo))
		return
	}
	ratio := passo / distancia
	novoX := origemX + int32(math.Round(float64(destinoX-origemX)*ratio))
	novoY := origemY + int32(math.Round(float64(destinoY-origemY)*ratio))
	novoZ := origemZ + int32(math.Round(float64(destinoZ-origemZ)*ratio))
	g.playerAtivo.aplicarPosicao(novoX, novoY, novoZ, g.playerAtivo.heading)
	g.sincronizarPersonagemAtualComPlayerAtivo()
	if distancia3D(g.playerAtivo.ultimoMoveX, g.playerAtivo.ultimoMoveY, g.playerAtivo.ultimoMoveZ, g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z) < 16 {
		return
	}
	g.playerAtivo.ultimoMoveX = g.playerAtivo.x
	g.playerAtivo.ultimoMoveY = g.playerAtivo.y
	g.playerAtivo.ultimoMoveZ = g.playerAtivo.z
	g.persistirPosicaoPlayerAtivoSeNecessario()
	g.broadcastPacoteParaVisiveis(montarValidateLocationPacket(g.playerAtivo))
}

func (g *gameClient) obterVelocidadeMovimentoRuntime() float64 {
	if g == nil {
		return 0
	}
	if g.playerAtivo == nil {
		return 0
	}
	template, ok := obterTemplatePersonagemInicial(g.playerAtivo.classID)
	if !ok {
		if g.playerAtivo.correndo {
			return 120
		}
		return 80
	}
	itensPapelBoneca := listarItensPapelBoneca(g.itensAtivos)
	statsCalculadas := calcularStatsPersonagem(template, g.playerAtivo.nivel, itensPapelBoneca)
	if g.playerAtivo.correndo {
		if statsCalculadas.runSpd > 0 {
			return float64(statsCalculadas.runSpd)
		}
		return float64(template.runSpd)
	}
	if statsCalculadas.walkSpd > 0 {
		return float64(statsCalculadas.walkSpd)
	}
	return float64(template.walkSpd)
}

func (g *gameClient) persistirPosicaoPlayerAtivoSeNecessario() {
	if g == nil {
		return
	}
	if g.playerAtivo == nil {
		return
	}
	if g.server == nil {
		return
	}
	if g.server.characterRepo == nil {
		return
	}
	agoraMs := time.Now().UnixMilli()
	if g.playerAtivo.ultimoPersistMs > 0 && agoraMs-g.playerAtivo.ultimoPersistMs < 3000 {
		return
	}
	err := g.server.characterRepo.AtualizarPosicao(context.Background(), g.playerAtivo.objID, g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z)
	if err != nil {
		return
	}
	g.playerAtivo.ultimoPersistMs = agoraMs
}

func (g *gameServer) processarTickNpcGlobal(intervalo time.Duration) {
	if g == nil {
		return
	}
	if g.mundo == nil {
		return
	}
	npcs := g.mundo.listarNpcsGlobais()
	players := g.mundo.listarPlayersAtivos()
	for _, npc := range npcs {
		if npc == nil {
			continue
		}
		if !npc.ehMonster {
			continue
		}
		if !npc.canMove {
			continue
		}
		if npc.aggroRange <= 0 && distanciaAtePlayerMaisProximo(npc, players) > raioVisibilidade {
			continue
		}
		g.processarMovimentoNpcGlobal(npc, players, intervalo)
	}
}

func (g *gameServer) processarMovimentoNpcGlobal(npc *npcGlobalRuntime, players []*gameClient, intervalo time.Duration) {
	if npc == nil {
		return
	}
	alvo := selecionarAlvoNpcGlobal(npc, players)
	if alvo != nil {
		mudouAlvo := npc.alvoObjID != alvo.playerAtivo.objID
		npc.alvoObjID = alvo.playerAtivo.objID
		moverNpcGlobalPasso(npc, alvo.playerAtivo.x, alvo.playerAtivo.y, alvo.playerAtivo.z, intervalo)
		g.broadcastNpcGlobalAtualizado(npc, mudouAlvo)
		return
	}
	tinhaAlvo := npc.alvoObjID > 0
	npc.alvoObjID = 0
	if distancia3D(npc.x, npc.y, npc.z, npc.origemX, npc.origemY, npc.origemZ) <= 8 {
		if tinhaAlvo {
			g.broadcastNpcGlobalAtualizado(npc, true)
		}
		return
	}
	moverNpcGlobalPasso(npc, npc.origemX, npc.origemY, npc.origemZ, intervalo)
	g.broadcastNpcGlobalAtualizado(npc, tinhaAlvo)
}

func selecionarAlvoNpcGlobal(npc *npcGlobalRuntime, players []*gameClient) *gameClient {
	if npc == nil {
		return nil
	}
	melhorDistancia := 0.0
	var melhor *gameClient
	for _, cliente := range players {
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		distancia := distancia3D(npc.x, npc.y, npc.z, cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z)
		if npc.aggroRange > 0 && distancia > float64(npc.aggroRange) {
			continue
		}
		distanciaOrigem := distancia3D(npc.origemX, npc.origemY, npc.origemZ, cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z)
		if npc.aggroRange > 0 && distanciaOrigem > float64(npc.aggroRange) {
			continue
		}
		if melhor == nil || distancia < melhorDistancia {
			melhor = cliente
			melhorDistancia = distancia
		}
	}
	return melhor
}

func distanciaAtePlayerMaisProximo(npc *npcGlobalRuntime, players []*gameClient) float64 {
	if npc == nil {
		return 999999
	}
	melhor := 999999.0
	for _, cliente := range players {
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		distancia := distancia3D(npc.x, npc.y, npc.z, cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z)
		if distancia < melhor {
			melhor = distancia
		}
	}
	return melhor
}

func moverNpcGlobalPasso(npc *npcGlobalRuntime, destinoX int32, destinoY int32, destinoZ int32, intervalo time.Duration) {
	if npc == nil {
		return
	}
	distancia := distancia3D(npc.x, npc.y, npc.z, destinoX, destinoY, destinoZ)
	heading := calcularHeading(npc.x, npc.y, destinoX, destinoY)
	if distancia <= 1 {
		npc.aplicarPosicaoComHeading(destinoX, destinoY, destinoZ, heading)
		return
	}
	velocidade := float64(npc.runSpd)
	if velocidade <= 0 {
		velocidade = 120
	}
	passo := velocidade * intervalo.Seconds()
	if passo >= distancia {
		npc.aplicarPosicaoComHeading(destinoX, destinoY, destinoZ, heading)
		return
	}
	ratio := passo / distancia
	novoX := npc.x + int32(math.Round(float64(destinoX-npc.x)*ratio))
	novoY := npc.y + int32(math.Round(float64(destinoY-npc.y)*ratio))
	novoZ := npc.z + int32(math.Round(float64(destinoZ-npc.z)*ratio))
	npc.aplicarPosicaoComHeading(novoX, novoY, novoZ, heading)
}

func (g *gameServer) broadcastNpcGlobalAtualizado(npc *npcGlobalRuntime, forcar bool) {
	if g == nil {
		return
	}
	if g.mundo == nil {
		return
	}
	if npc == nil {
		return
	}
	if !forcar && distancia3D(npc.ultimoMoveX, npc.ultimoMoveY, npc.ultimoMoveZ, npc.x, npc.y, npc.z) < 32 {
		return
	}
	npc.ultimoMoveX = npc.x
	npc.ultimoMoveY = npc.y
	npc.ultimoMoveZ = npc.z
	pacote := montarNpcGlobalInfoPacket(npc)
	for _, cliente := range g.mundo.listarPlayersVisiveisParaNpc(npc) {
		if cliente == nil {
			continue
		}
		_ = cliente.enviarPacket(pacote)
	}
}
