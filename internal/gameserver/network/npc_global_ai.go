package network

import (
	"math/rand"
	"strings"
	"time"
)

const intervaloAiNpcMs int64 = 1000

func (n *npcGlobalRuntime) limparEstadoAi() {
	if n == nil {
		return
	}
	n.aiLifeTime = 0
	n.aiStep = 0
	n.aiUltimoProcessamentoMs = 0
	n.aiSeenPlayers = map[int32]int64{}
	n.aiTopDesireTargetObjID = 0
	n.aiHateList = novoNpcHateList()
	n.aiUltimoDesire = npcDesire{}
	n.aiProximoDesire = npcDesire{}
	n.aiDesireQueue = novoNpcDesireQueue()
}

func (n *npcGlobalRuntime) podeProcessarAiAgora() bool {
	if n == nil {
		return false
	}
	agoraMs := time.Now().UnixMilli()
	if n.aiUltimoProcessamentoMs <= 0 {
		n.aiUltimoProcessamentoMs = agoraMs
		return true
	}
	if agoraMs-n.aiUltimoProcessamentoMs < intervaloAiNpcMs {
		return false
	}
	n.aiUltimoProcessamentoMs = agoraMs
	return true
}

func (n *npcGlobalRuntime) adicionarDesireAtaque(alvoObjID int32, peso float64, atualizarAggro bool) {
	if n == nil {
		return
	}
	if alvoObjID <= 0 {
		return
	}
	n.aiDesireQueue.adicionarOuAtualizar(novoNpcDesireAttack(alvoObjID, peso, true))
	if !atualizarAggro {
		return
	}
	danoBase := int32(peso)
	if danoBase < 1 {
		danoBase = 1
	}
	n.registrarAggro(alvoObjID, danoBase)
	n.aiHateList.adicionar(alvoObjID, peso)
}

func (n *npcGlobalRuntime) limparDesiresInvalidos(players []*gameClient) {
	if n == nil {
		return
	}
	if n.aiDesireQueue.estaVazia() {
		return
	}
	filtrados := n.aiDesireQueue.desires[:0]
	for _, desire := range n.aiDesireQueue.desires {
		if desire.alvoObjID <= 0 {
			filtrados = append(filtrados, desire)
			continue
		}
		cliente := localizarPlayerPorObjID(players, desire.alvoObjID)
		if cliente == nil {
			n.aiHateList.remover(desire.alvoObjID)
			delete(n.hatePorAlvo, desire.alvoObjID)
			continue
		}
		if cliente.playerAtivo == nil {
			n.aiHateList.remover(desire.alvoObjID)
			delete(n.hatePorAlvo, desire.alvoObjID)
			continue
		}
		if cliente.playerAtivo.hpAtual <= 0 {
			n.aiHateList.remover(desire.alvoObjID)
			delete(n.hatePorAlvo, desire.alvoObjID)
			continue
		}
		if cliente.playerAtivo.estaProtegidoSpawn() {
			n.aiHateList.remover(desire.alvoObjID)
			delete(n.hatePorAlvo, desire.alvoObjID)
			continue
		}
		filtrados = append(filtrados, desire)
	}
	n.aiDesireQueue.desires = filtrados
	if n.aiTopDesireTargetObjID > 0 {
		clienteTop := localizarPlayerPorObjID(players, n.aiTopDesireTargetObjID)
		if clienteTop == nil || clienteTop.playerAtivo == nil {
			n.aiTopDesireTargetObjID = 0
			n.alvoObjID = 0
		}
		if clienteTop != nil && clienteTop.playerAtivo != nil {
			if clienteTop.playerAtivo.hpAtual <= 0 || clienteTop.playerAtivo.estaProtegidoSpawn() {
				n.aiTopDesireTargetObjID = 0
				n.alvoObjID = 0
			}
		}
	}
}

func (n *npcGlobalRuntime) voltarParaPazSeSemAggroOuHate() {
	if n == nil {
		return
	}
	if len(n.hatePorAlvo) > 0 {
		return
	}
	if n.aiHateList.obterMaisOdiado() > 0 {
		return
	}
	if !n.aiDesireQueue.estaVazia() {
		return
	}
	n.alvoObjID = 0
	n.aiTopDesireTargetObjID = 0
	n.retornandoSpawn = true
}

func (n *npcGlobalRuntime) reconsiderarAlvo(players []*gameClient) int32 {
	if n == nil {
		return 0
	}
	maiorHate := int32(0)
	melhorObjID := int32(0)
	for alvoObjID, hate := range n.hatePorAlvo {
		cliente := localizarPlayerPorObjID(players, alvoObjID)
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		if cliente.playerAtivo.hpAtual <= 0 {
			continue
		}
		if cliente.playerAtivo.estaProtegidoSpawn() {
			continue
		}
		if hate <= maiorHate {
			continue
		}
		if !n.estaAlvoEmRangePerseguicao(cliente) {
			continue
		}
		melhorObjID = alvoObjID
		maiorHate = hate
	}
	if melhorObjID > 0 {
		return melhorObjID
	}
	if !n.ehScriptWarriorBase() && n.ehAgressivo() {
		for _, cliente := range players {
			if cliente == nil {
				continue
			}
			if cliente.playerAtivo == nil {
				continue
			}
			if cliente.playerAtivo.hpAtual <= 0 {
				continue
			}
			if cliente.playerAtivo.estaProtegidoSpawn() {
				continue
			}
			if !n.estaAlvoEmRangeAggro(cliente) {
				continue
			}
			n.adicionarDesireAtaque(cliente.playerAtivo.objID, 1, true)
			return cliente.playerAtivo.objID
		}
	}
	return 0
}

func (n *npcGlobalRuntime) estaAlvoEmRangePerseguicao(cliente *gameClient) bool {
	if n == nil || cliente == nil || cliente.playerAtivo == nil {
		return false
	}
	limitePerseguicao := float64(1200)
	if n.ehAgressivo() {
		limitePerseguicao = float64(n.aggroRange * 3)
		if limitePerseguicao < 600 {
			limitePerseguicao = 600
		}
	}
	distancia := distancia3D(n.x, n.y, n.z, cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z)
	if distancia > limitePerseguicao {
		return false
	}
	return true
}

func (n *npcGlobalRuntime) estaAlvoEmRangeAggro(cliente *gameClient) bool {
	if n == nil || cliente == nil || cliente.playerAtivo == nil {
		return false
	}
	distancia := distancia3D(n.x, n.y, n.z, cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z)
	if distancia > float64(n.aggroRange) {
		return false
	}
	distanciaOrigem := distancia3D(n.origemX, n.origemY, n.origemZ, cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z)
	if distanciaOrigem > float64(n.aggroRange) {
		return false
	}
	return true
}

func (n *npcGlobalRuntime) selecionarAlvoPorHateAi(players []*gameClient) int32 {
	if n == nil {
		return 0
	}
	alvoObjID := n.aiHateList.obterMaisOdiado()
	if alvoObjID <= 0 {
		return 0
	}
	cliente := localizarPlayerPorObjID(players, alvoObjID)
	if cliente == nil {
		n.aiHateList.remover(alvoObjID)
		return 0
	}
	if cliente.playerAtivo == nil {
		n.aiHateList.remover(alvoObjID)
		return 0
	}
	if cliente.playerAtivo.hpAtual <= 0 {
		n.aiHateList.remover(alvoObjID)
		return 0
	}
	if cliente.playerAtivo.estaProtegidoSpawn() {
		n.aiHateList.remover(alvoObjID)
		return 0
	}
	return alvoObjID
}

func (n *npcGlobalRuntime) processarSeeCreature(players []*gameClient) {
	if n == nil {
		return
	}
	for _, cliente := range players {
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		player := cliente.playerAtivo
		if player.hpAtual <= 0 {
			continue
		}
		if player.estaProtegidoSpawn() {
			continue
		}
		if !n.aiParams.canSeeThrough && player.movendo == false {
			// Placeholder de regra leve para manter o escopo minimo desta fase.
		}
		if distancia3D(n.x, n.y, n.z, player.x, player.y, player.z) > float64(n.obterSeeRange()) {
			delete(n.aiSeenPlayers, player.objID)
			continue
		}
		n.aiSeenPlayers[player.objID] = time.Now().UnixMilli()
		n.processarSeeCreatureWarriorBase(cliente)
	}
}

func (n *npcGlobalRuntime) processarSeeCreatureWarriorBase(cliente *gameClient) {
	if n == nil {
		return
	}
	if cliente == nil || cliente.playerAtivo == nil {
		return
	}
	if !n.ehScriptWarriorBase() {
		return
	}
	if n.estaEmCombateOuCastAi() {
		return
	}
	player := cliente.playerAtivo
	if player == nil {
		return
	}
	halfAggressive := n.obterNpcIntAiParamOuPadrao("HalfAggressive", n.aiParams.halfAggressive)
	if halfAggressive == 1 {
		horaJogo := time.Now().Hour()
		if horaJogo < 5 {
			return
		}
		n.tryToAttack(cliente)
		return
	}
	if halfAggressive == 2 {
		horaJogo := time.Now().Hour()
		if horaJogo >= 5 {
			return
		}
		n.tryToAttack(cliente)
		return
	}
	randomAggressive := n.obterNpcIntAiParamOuPadrao("RandomAggressive", n.aiParams.randomAggressive)
	if randomAggressive > 0 {
		if rand.Intn(100) >= int(randomAggressive) {
			n.aiDesireQueue.removerPorTipoEAlvo(tipoNpcDesireAttack, player.objID)
			return
		}
		n.tryToAttack(cliente)
		return
	}
	isVs := n.obterNpcIntAiParamOuPadrao("IsVs", n.aiParams.isVs)
	if isVs == 1 {
		diferencaNivel := n.nivel - player.nivel
		if diferencaNivel < 0 {
			diferencaNivel = -diferencaNivel
		}
		if diferencaNivel <= 2 {
			n.tryToAttack(cliente)
			return
		}
	}
	attackLowLevel := n.obterNpcIntAiParamOuPadrao("AttackLowLevel", n.aiParams.attackLowLevel)
	if attackLowLevel == 1 {
		if player.nivel+15 < n.nivel {
			n.adicionarDesireAtaque(player.objID, 700, true)
			return
		}
	}
	daggerBackAttack := n.obterNpcIntAiParamOuPadrao("DaggerBackAttack", n.aiParams.daggerBackAttack)
	if daggerBackAttack == 1 {
		if rand.Intn(100) < 50 {
			if distancia3D(n.x, n.y, n.z, player.x, player.y, player.z) <= 100 {
				n.tryToAttack(cliente)
				return
			}
		}
	}
}

func (n *npcGlobalRuntime) estaEmCombateOuCastAi() bool {
	if n == nil {
		return false
	}
	if n.alvoObjID > 0 {
		return true
	}
	if n.aiTopDesireTargetObjID > 0 {
		return true
	}
	if !n.aiDesireQueue.estaVazia() {
		return true
	}
	if len(n.hatePorAlvo) > 0 {
		return true
	}
	if n.aiHateList.obterMaisOdiado() > 0 {
		return true
	}
	return false
}

func (n *npcGlobalRuntime) tryToAttack(cliente *gameClient) {
	if n == nil {
		return
	}
	if cliente == nil || cliente.playerAtivo == nil {
		return
	}
	if !n.estaVivo() {
		return
	}
	randomAggressive := n.obterNpcIntAiParamOuPadrao("RandomAggressive", n.aiParams.randomAggressive)
	halfAggressive := n.obterNpcIntAiParamOuPadrao("HalfAggressive", n.aiParams.halfAggressive)
	if !n.ehAgressivo() && randomAggressive <= 0 && halfAggressive <= 0 {
		return
	}
	if !n.estaDentroTerritorioOrigem(cliente.playerAtivo.x, cliente.playerAtivo.y, cliente.playerAtivo.z) {
		return
	}
	if !n.passouSetAggressiveTime() {
		return
	}
	n.adicionarDesireAtaque(cliente.playerAtivo.objID, 200, true)
}

func (n *npcGlobalRuntime) passouSetAggressiveTime() bool {
	if n == nil {
		return false
	}
	setAggressiveTime := n.obterNpcIntAiParamOuPadrao("SetAggressiveTime", n.aiParams.setAggressiveTime)
	if setAggressiveTime == 0 {
		return true
	}
	if setAggressiveTime == -1 {
		limite := int32(rand.Intn(5) + 3)
		return n.aiLifeTime >= limite
	}
	limite := setAggressiveTime + int32(rand.Intn(4))
	return n.aiLifeTime > limite
}

func (n *npcGlobalRuntime) obterSeeRange() int32 {
	if n == nil {
		return 400
	}
	return 400
}

func (n *npcGlobalRuntime) ehScriptWarriorBase() bool {
	if n == nil {
		return false
	}
	if possuiScriptAiMonsterCarregado(n.npcID) {
		if n.ehScriptAiBase("WarriorBase") {
			return true
		}
		if n.ehScriptAiVariante("Warrior") {
			return true
		}
		descritorScript := strings.ToLower(strings.TrimSpace(n.scriptAiDescritor))
		if strings.Contains(descritorScript, "monster/warriorbase") {
			return true
		}
		return false
	}
	if n.ehScriptAiBase("WarriorBase") {
		return true
	}
	if n.ehScriptAiVariante("Warrior") {
		return true
	}
	descritorScript := strings.ToLower(strings.TrimSpace(n.scriptAiDescritor))
	if strings.Contains(descritorScript, "monster/warriorbase") {
		return true
	}
	tipoAi := strings.ToLower(strings.TrimSpace(n.tipoAI))
	if strings.Contains(tipoAi, "warrior") {
		return true
	}
	if strings.Contains(tipoAi, "warriorbase") {
		return true
	}
	return false
}

func (n *npcGlobalRuntime) estaDentroTerritorioOrigem(x int32, y int32, z int32) bool {
	if n == nil {
		return false
	}
	if len(n.spawnTerritorio.nos) == 0 {
		return true
	}
	if z < n.spawnTerritorio.minZ || z > n.spawnTerritorio.maxZ {
		return false
	}
	if x < n.spawnTerritorio.minX || x > n.spawnTerritorio.maxX {
		return false
	}
	if y < n.spawnTerritorio.minY || y > n.spawnTerritorio.maxY {
		return false
	}
	return true
}

func (g *gameServer) processarAiNpcGlobal(npc *npcGlobalRuntime, players []*gameClient) {
	if g == nil || npc == nil {
		return
	}
	if !npc.estaVivo() {
		return
	}
	if !npc.ehMonster {
		return
	}
	if !npc.podeProcessarAiAgora() {
		return
	}
	npc.limparDesiresInvalidos(players)
	npc.processarSeeCreature(players)
	g.executarDesireDominanteNpcGlobal(npc, players)
	if npc.aiTopDesireTargetObjID > 0 {
		npc.alvoObjID = npc.aiTopDesireTargetObjID
	}
	npc.aiLifeTime++
	npc.aiStep++
	if npc.aiStep%3 == 0 {
		npc.hateReduzirPeriodicamente()
		npc.aiHateList.reduzirTodos(6.6)
		npc.aiDesireQueue.diminuirPesoPorTipo(tipoNpcDesireAttack, 6.6)
		npc.aiStep = 0
	}
}

func (g *gameServer) executarDesireDominanteNpcGlobal(npc *npcGlobalRuntime, players []*gameClient) {
	if g == nil || npc == nil {
		return
	}
	desire := npc.aiDesireQueue.obterMaiorPeso()
	if desire == nil {
		npc.aiTopDesireTargetObjID = npc.selecionarAlvoPorHateAi(players)
		if npc.aiTopDesireTargetObjID > 0 {
			npc.alvoObjID = npc.aiTopDesireTargetObjID
			return
		}
		novoAlvoObjID := npc.reconsiderarAlvo(players)
		if novoAlvoObjID > 0 {
			npc.aiTopDesireTargetObjID = novoAlvoObjID
			npc.alvoObjID = novoAlvoObjID
			return
		}
		npc.alvoObjID = 0
		npc.voltarParaPazSeSemAggroOuHate()
		return
	}
	npc.aiTopDesireTargetObjID = desire.alvoObjID
	npc.aiUltimoDesire = *desire
	if desire.tipo != tipoNpcDesireAttack {
		return
	}
	npc.alvoObjID = desire.alvoObjID
}

func (n *npcGlobalRuntime) hateReduzirPeriodicamente() {
	if n == nil {
		return
	}
	if len(n.hatePorAlvo) == 0 {
		return
	}
	for alvoObjID, hate := range n.hatePorAlvo {
		novoHate := hate - 6
		if novoHate > 0 {
			n.hatePorAlvo[alvoObjID] = novoHate
			continue
		}
		delete(n.hatePorAlvo, alvoObjID)
		n.aiHateList.remover(alvoObjID)
		n.aiDesireQueue.removerPorTipoEAlvo(tipoNpcDesireAttack, alvoObjID)
	}
	if len(n.hatePorAlvo) > 0 {
		return
	}
	n.alvoObjID = 0
	n.voltarParaPazSeSemAggroOuHate()
}
