package network

import (
	"time"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

const distanciaAtaqueMob = 90.0
const tamanhoBarraHp = int32(352)

func calcularSegmentoHpBarra(hpAtual int32, hpMaximo int32) int32 {
	if hpMaximo <= 0 {
		return 0
	}
	return hpAtual * tamanhoBarraHp / hpMaximo
}

func (n *npcGlobalRuntime) deveBroadcastarStatusHp() bool {
	if n == nil {
		return true
	}
	segmento := calcularSegmentoHpBarra(n.hpAtual, n.hpMaximo)
	if segmento == n.hpBarSegmentoAnterior && n.hpAtual > 0 {
		return false
	}
	n.hpBarSegmentoAnterior = segmento
	return true
}

func (n *npcGlobalRuntime) notificarEventoAi() {
	if n == nil {
		return
	}
	if n.canalEventoAi == nil {
		return
	}
	select {
	case n.canalEventoAi <- struct{}{}:
	default:
	}
}

func (n *npcGlobalRuntime) registrarAggro(alvoObjID int32, dano int32) {
	if n == nil {
		return
	}
	if alvoObjID <= 0 {
		return
	}
	if n.hatePorAlvo == nil {
		n.hatePorAlvo = map[int32]int32{}
	}
	hateAtual := n.hatePorAlvo[alvoObjID]
	if dano < 1 {
		dano = 1
	}
	n.hatePorAlvo[alvoObjID] = hateAtual + dano
	n.aiHateList.adicionar(alvoObjID, float64(dano))
	n.ultimoAggroMs = time.Now().UnixMilli()
	n.retornandoSpawn = false
}

func (n *npcGlobalRuntime) limparAggro(alvoObjID int32) {
	if n == nil {
		return
	}
	if n.hatePorAlvo == nil {
		return
	}
	delete(n.hatePorAlvo, alvoObjID)
	n.aiHateList.remover(alvoObjID)
	n.aiDesireQueue.removerPorTipoEAlvo(tipoNpcDesireAttack, alvoObjID)
	if n.alvoObjID == alvoObjID {
		n.alvoObjID = 0
	}
}

func (n *npcGlobalRuntime) limparTodoAggro() {
	if n == nil {
		return
	}
	n.hatePorAlvo = map[int32]int32{}
	n.aiHateList.limpar()
	n.aiDesireQueue.limpar()
	n.alvoObjID = 0
}

func (n *npcGlobalRuntime) agendarRespawn() {
	if n == nil {
		return
	}
	if n.respawnDelayMs <= 0 {
		n.respawnAteMs = 0
		return
	}
	n.respawnAteMs = time.Now().UnixMilli() + n.respawnDelayMs
}

func (n *npcGlobalRuntime) prontoParaRespawn() bool {
	if n == nil {
		return false
	}
	if !n.estaMorto {
		return false
	}
	if n.respawnAteMs <= 0 {
		return false
	}
	return time.Now().UnixMilli() >= n.respawnAteMs
}

func (n *npcGlobalRuntime) resetarEstadoRespawn() {
	if n == nil {
		return
	}
	n.hpAtual = n.hpMaximo
	n.mpAtual = n.mpMaximo
	n.estaMorto = false
	n.respawnAteMs = 0
	n.ultimoAggroMs = 0
	n.ultimoAtaqueMs = 0
	n.ultimoRegenMs = 0
	n.retornandoSpawn = false
	n.limparTodoAggro()
	n.limparEstadoAi()
	n.limparDadosReward()
	x, y, z, heading, ok := resolverPosicaoSpawnGlobal(n.spawnTerritorio, n.spawnPosFixa, 0, n.ehMonster, n.spawnTerritorio.nome != "")
	if !ok {
		x = n.origemX
		y = n.origemY
		z = n.origemZ
		heading = n.heading
	}
	n.origemX = x
	n.origemY = y
	n.origemZ = z
	n.aplicarPosicaoComHeading(x, y, z, heading)
	n.ultimoMoveX = x
	n.ultimoMoveY = y
	n.ultimoMoveZ = z
}

func (g *gameServer) processarRespawnNpcGlobal(npc *npcGlobalRuntime) {
	if g == nil {
		return
	}
	if g.mundo == nil {
		return
	}
	if npc == nil {
		return
	}
	if !npc.prontoParaRespawn() {
		return
	}
	npc.resetarEstadoRespawn()
	logger.Infof("Mob respawnado npcID=%d objIDNpc=%d nome=%s pos=(%d,%d,%d)", npc.npcID, npc.objID, npc.nome, npc.x, npc.y, npc.z)
	pacote := montarNpcGlobalInfoPacket(npc)
	for _, cliente := range g.mundo.listarPlayersVisiveisParaNpc(npc) {
		if cliente == nil {
			continue
		}
		_ = cliente.enviarPacket(pacote)
	}
}

func (n *npcGlobalRuntime) podeAtacarAgora() bool {
	if n == nil {
		return false
	}
	agoraMs := time.Now().UnixMilli()
	if n.ultimoAtaqueMs <= 0 {
		n.ultimoAtaqueMs = agoraMs
		return true
	}
	intervaloMs := int64(1200)
	if n.pAtkSpd > 0 {
		intervaloMs = int64(500000 / n.pAtkSpd)
	}
	if intervaloMs < 400 {
		intervaloMs = 400
	}
	if agoraMs-n.ultimoAtaqueMs < intervaloMs {
		return false
	}
	n.ultimoAtaqueMs = agoraMs
	return true
}

func (n *npcGlobalRuntime) regenerarSeNecessario() {
	if n == nil {
		return
	}
	if !n.estaVivo() {
		return
	}
	agoraMs := time.Now().UnixMilli()
	if n.ultimoRegenMs > 0 && agoraMs-n.ultimoRegenMs < 3000 {
		return
	}
	n.ultimoRegenMs = agoraMs
	if n.alvoObjID > 0 {
		return
	}
	if len(n.hatePorAlvo) > 0 {
		return
	}
	if n.hpAtual < n.hpMaximo {
		regenHp := n.hpMaximo / 20
		if regenHp < 1 {
			regenHp = 1
		}
		n.hpAtual += regenHp
		if n.hpAtual > n.hpMaximo {
			n.hpAtual = n.hpMaximo
		}
	}
	if n.mpAtual < n.mpMaximo {
		regenMp := n.mpMaximo / 20
		if regenMp < 1 {
			regenMp = 1
		}
		n.mpAtual += regenMp
		if n.mpAtual > n.mpMaximo {
			n.mpAtual = n.mpMaximo
		}
	}
}

func (n *npcGlobalRuntime) deveRetornarSpawn() bool {
	if n == nil {
		return false
	}
	if !n.estaVivo() {
		return false
	}
	if len(n.hatePorAlvo) > 0 {
		return false
	}
	if n.alvoObjID > 0 {
		return false
	}
	if distancia3D(n.x, n.y, n.z, n.origemX, n.origemY, n.origemZ) <= 8 {
		return false
	}
	return true
}

func selecionarAlvoNpcGlobalPorHate(npc *npcGlobalRuntime, players []*gameClient) *gameClient {
	if npc == nil {
		return nil
	}
	if npc.aiTopDesireTargetObjID > 0 {
		cliente := localizarPlayerPorObjID(players, npc.aiTopDesireTargetObjID)
		if cliente != nil && cliente.playerAtivo != nil && cliente.playerAtivo.hpAtual > 0 && !cliente.playerAtivo.estaProtegidoSpawn() {
			if npc.estaAlvoEmRangePerseguicao(cliente) {
				return cliente
			}
		}
	}
	if len(npc.hatePorAlvo) == 0 {
		alvoObjID := npc.aiHateList.obterMaisOdiado()
		if alvoObjID <= 0 {
			return nil
		}
		cliente := localizarPlayerPorObjID(players, alvoObjID)
		if cliente == nil || cliente.playerAtivo == nil {
			npc.aiHateList.remover(alvoObjID)
			return nil
		}
		if cliente.playerAtivo.hpAtual <= 0 || cliente.playerAtivo.estaProtegidoSpawn() {
			npc.aiHateList.remover(alvoObjID)
			return nil
		}
		if !npc.estaAlvoEmRangePerseguicao(cliente) {
			return nil
		}
		return cliente
	}
	maiorHate := int32(0)
	var melhor *gameClient
	for _, cliente := range players {
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		if cliente.playerAtivo.estaProtegidoSpawn() {
			continue
		}
		if cliente.playerAtivo.hpAtual <= 0 {
			npc.limparAggro(cliente.playerAtivo.objID)
			continue
		}
		hate := npc.hatePorAlvo[cliente.playerAtivo.objID]
		if hate <= 0 {
			continue
		}
		distancia := distancia3D(npc.x, npc.y, npc.z, cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z)
		limitePerseguicao := float64(1200)
		if npc.ehAgressivo() {
			limitePerseguicao = float64(npc.aggroRange * 3)
			if limitePerseguicao < 600 {
				limitePerseguicao = 600
			}
		}
		if distancia > limitePerseguicao {
			continue
		}
		if melhor == nil || hate > maiorHate {
			melhor = cliente
			maiorHate = hate
		}
	}
	if melhor != nil {
		return melhor
	}
	alvoObjID := npc.aiHateList.obterMaisOdiado()
	if alvoObjID <= 0 {
		return nil
	}
	cliente := localizarPlayerPorObjID(players, alvoObjID)
	if cliente == nil || cliente.playerAtivo == nil {
		npc.aiHateList.remover(alvoObjID)
		return nil
	}
	if cliente.playerAtivo.hpAtual <= 0 || cliente.playerAtivo.estaProtegidoSpawn() {
		npc.aiHateList.remover(alvoObjID)
		return nil
	}
	if !npc.estaAlvoEmRangePerseguicao(cliente) {
		return nil
	}
	return cliente
}

func (n *npcGlobalRuntime) refreshAggro(players []*gameClient) {
	if n == nil {
		return
	}
	if len(n.hatePorAlvo) == 0 {
		alvoObjID := n.aiHateList.obterMaisOdiado()
		if alvoObjID <= 0 {
			n.aiTopDesireTargetObjID = 0
			n.alvoObjID = 0
			n.voltarParaPazSeSemAggroOuHate()
			return
		}
		cliente := localizarPlayerPorObjID(players, alvoObjID)
		if cliente == nil || cliente.playerAtivo == nil {
			n.aiHateList.remover(alvoObjID)
			n.aiTopDesireTargetObjID = 0
			n.alvoObjID = 0
			n.voltarParaPazSeSemAggroOuHate()
			return
		}
		if cliente.playerAtivo.hpAtual <= 0 || cliente.playerAtivo.estaProtegidoSpawn() {
			n.aiHateList.remover(alvoObjID)
			n.aiTopDesireTargetObjID = 0
			n.alvoObjID = 0
			n.voltarParaPazSeSemAggroOuHate()
			return
		}
		if !n.estaAlvoEmRangePerseguicao(cliente) {
			n.aiHateList.remover(alvoObjID)
			n.aiTopDesireTargetObjID = 0
			n.alvoObjID = 0
			n.voltarParaPazSeSemAggroOuHate()
			return
		}
		n.aiTopDesireTargetObjID = alvoObjID
		n.alvoObjID = alvoObjID
		return
	}
	for alvoObjID := range n.hatePorAlvo {
		cliente := localizarPlayerPorObjID(players, alvoObjID)
		if cliente == nil {
			n.limparAggro(alvoObjID)
			continue
		}
		if cliente.playerAtivo == nil {
			n.limparAggro(alvoObjID)
			continue
		}
		if cliente.playerAtivo.hpAtual <= 0 {
			n.limparAggro(alvoObjID)
			continue
		}
		if cliente.playerAtivo.estaProtegidoSpawn() {
			n.limparAggro(alvoObjID)
			continue
		}
		if n.aiTopDesireTargetObjID == alvoObjID {
			n.alvoObjID = alvoObjID
		}
		distanciaSpawn := distancia3D(n.origemX, n.origemY, n.origemZ, cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z)
		limiteRefresh := float64(1500)
		if n.ehAgressivo() {
			limiteRefresh = float64(n.aggroRange * 4)
			if limiteRefresh < 800 {
				limiteRefresh = 800
			}
		}
		if distanciaSpawn > limiteRefresh {
			n.limparAggro(alvoObjID)
		}
	}
	if len(n.hatePorAlvo) > 0 {
		return
	}
	n.alvoObjID = 0
	n.aiTopDesireTargetObjID = 0
	n.voltarParaPazSeSemAggroOuHate()
}

func localizarPlayerPorObjID(players []*gameClient, objID int32) *gameClient {
	if objID <= 0 {
		return nil
	}
	for _, cliente := range players {
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		if cliente.playerAtivo.objID != objID {
			continue
		}
		return cliente
	}
	return nil
}

func (g *gameServer) processarPosCombateNpcGlobal(npc *npcGlobalRuntime, players []*gameClient) {
	if npc == nil {
		return
	}
	npc.refreshAggro(players)
	npc.regenerarSeNecessario()
	if !npc.deveRetornarSpawn() {
		return
	}
	npc.retornandoSpawn = true
}

func (n *npcGlobalRuntime) estaVivo() bool {
	if n == nil {
		return false
	}
	if n.estaMorto {
		return false
	}
	return n.hpAtual > 0
}

func (n *npcGlobalRuntime) aplicarDano(dano int32) bool {
	if n == nil {
		return false
	}
	if !n.estaVivo() {
		return false
	}
	if dano <= 0 {
		dano = 1
	}
	n.hpAtual -= dano
	if n.hpAtual > 0 {
		return false
	}
	n.hpAtual = 0
	n.estaMorto = true
	n.alvoObjID = 0
	n.agendarRespawn()
	return true
}

func (g *gameClient) calcularDanoFisicoContraNpc(npc *npcGlobalRuntime) int32 {
	if g == nil {
		return 1
	}
	if g.playerAtivo == nil {
		return 1
	}
	if npc == nil {
		return 1
	}
	template, ok := obterTemplatePersonagemInicial(g.playerAtivo.classID)
	if !ok {
		dano := g.playerAtivo.nivel * 5
		if dano < 1 {
			return 1
		}
		return dano
	}
	itensPapelBoneca := listarItensPapelBoneca(g.itensAtivos)
	stats := calcularStatsPersonagem(template, g.playerAtivo.nivel, itensPapelBoneca)
	resultado := calcularResultadoAtaquePlayerContraNpc(g.playerAtivo, npc, stats)
	return resultado.dano
}

func (g *gameServer) calcularDanoFisicoNpcContraPlayer(npc *npcGlobalRuntime, alvo *gameClient) int32 {
	if npc == nil {
		return 1
	}
	if alvo == nil {
		return 1
	}
	if alvo.playerAtivo == nil {
		return 1
	}
	template, ok := obterTemplatePersonagemInicial(alvo.playerAtivo.classID)
	if !ok {
		dano := npc.pAtk / 2
		if dano < 1 {
			return 1
		}
		return dano
	}
	itensPapelBoneca := listarItensPapelBoneca(alvo.itensAtivos)
	stats := calcularStatsPersonagem(template, alvo.playerAtivo.nivel, itensPapelBoneca)
	resultado := calcularResultadoAtaqueNpcContraPlayer(npc, alvo.playerAtivo, stats)
	return resultado.dano
}

func (g *gameClient) aplicarDanoRuntime(dano int32) {
	if g == nil {
		return
	}
	if g.playerAtivo == nil {
		return
	}
	if dano < 0 {
		return
	}
	hpAtual := g.playerAtivo.hpAtual - dano
	if hpAtual > 0 {
		g.playerAtivo.hpAtual = hpAtual
		g.sincronizarPersonagemAtualComPlayerAtivo()
		return
	}
	g.playerAtivo.hpAtual = 0
	g.playerAtivo.limparAlvo()
	g.sincronizarPersonagemAtualComPlayerAtivo()
}

func (g *gameServer) processarAtaqueNpcContraPlayer(npc *npcGlobalRuntime, alvo *gameClient) {
	if g == nil {
		return
	}
	if npc == nil {
		return
	}
	if alvo == nil {
		return
	}
	if alvo.playerAtivo == nil {
		return
	}
	if !npc.estaVivo() {
		return
	}
	if alvo.playerAtivo.hpAtual <= 0 {
		return
	}
	if distancia3D(npc.x, npc.y, npc.z, alvo.playerAtivo.x, alvo.playerAtivo.y, alvo.playerAtivo.z) > distanciaAtaqueMob {
		return
	}
	if !npc.podeAtacarAgora() {
		return
	}
	template, ok := obterTemplatePersonagemInicial(alvo.playerAtivo.classID)
	var resultado resultadoAtaqueFisico
	if ok {
		itensPapelBoneca := listarItensPapelBoneca(alvo.itensAtivos)
		stats := calcularStatsPersonagem(template, alvo.playerAtivo.nivel, itensPapelBoneca)
		resultado = calcularResultadoAtaqueNpcContraPlayer(npc, alvo.playerAtivo, stats)
	}
	if !ok {
		resultado = resultadoAtaqueFisico{dano: g.calcularDanoFisicoNpcContraPlayer(npc, alvo), errou: false, critico: false, defesaEscudo: "failed", intervaloAtaque: calcularIntervaloAtaqueFisico(npc.pAtkSpd)}
	}
	dano := resultado.dano
	if resultado.errou {
		dano = 0
	}
	alvo.aplicarDanoRuntime(dano)
	npc.registrarAggro(alvo.playerAtivo.objID, 1)
	logger.Infof("Ataque mob conta=%s npcID=%d objIDNpc=%d alvo=%s dano=%d critico=%t errou=%t escudo=%s hpNpc=%d/%d hpPlayer=%d/%d", alvo.conta, npc.npcID, npc.objID, alvo.playerAtivo.nome, dano, resultado.critico, resultado.errou, resultado.defesaEscudo, npc.hpAtual, npc.hpMaximo, alvo.playerAtivo.hpAtual, alvo.playerAtivo.hpMaximo)
	pachoteInicio := montarAutoAttackStartPacket(npc.objID)
	_ = alvo.enviarPacket(pachoteInicio)
	for _, cliente := range g.mundo.listarPlayersVisiveisParaNpc(npc) {
		if cliente == nil {
			continue
		}
		_ = cliente.enviarPacket(pachoteInicio)
	}
	pacoteAtaque := montarAttackPacket(npc.objID, alvo.playerAtivo.objID, dano, npc.x, npc.y, npc.z)
	_ = alvo.enviarPacket(pacoteAtaque)
	for _, cliente := range g.mundo.listarPlayersVisiveisParaNpc(npc) {
		if cliente == nil {
			continue
		}
		_ = cliente.enviarPacket(pacoteAtaque)
	}
	pacoteFim := montarAutoAttackStopPacket(npc.objID)
	_ = alvo.enviarPacket(pacoteFim)
	for _, cliente := range g.mundo.listarPlayersVisiveisParaNpc(npc) {
		if cliente == nil {
			continue
		}
		_ = cliente.enviarPacket(pacoteFim)
	}
	if resultado.errou {
		_ = alvo.enviarPacket(montarSystemMessageNome(msgIDAvoidedS1Attack, npc.nome))
	}
	statusUpdate := montarStatusUpdatePacket(alvo.playerAtivo.objID, [][2]int32{
		{statusAttrCurHp, alvo.playerAtivo.hpAtual},
		{statusAttrMaxHp, alvo.playerAtivo.hpMaximo},
	})
	_ = alvo.enviarPacket(statusUpdate)
	for _, cliente := range g.mundo.listarPlayersVisiveisParaNpc(npc) {
		if cliente == nil {
			continue
		}
		_ = cliente.enviarPacket(statusUpdate)
	}
	_ = alvo.enviarUserInfoAtualizado()
	alvo.broadcastCharInfoAtualizado()
}

func (g *gameServer) processarMorteNpcGlobal(npc *npcGlobalRuntime) {
	if g == nil {
		return
	}
	if g.mundo == nil {
		return
	}
	if npc == nil {
		return
	}
	logger.Infof("Mob morto npcID=%d objIDNpc=%d nome=%s", npc.npcID, npc.objID, npc.nome)
	pacoteDelete := montarDeleteObjectPacket(npc.objID)
	for _, cliente := range g.mundo.listarPlayersVisiveisParaNpc(npc) {
		if cliente == nil {
			continue
		}
		_ = cliente.enviarPacket(pacoteDelete)
	}
}
