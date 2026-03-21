package network

import (
	"context"
	"time"
)

func (g *gameClient) iniciarAutoAtaqueNpc(alvoObjID int32) {
	if g.playerAtivo == nil {
		return
	}
	if g.playerAtivo.autoAtaqueAlvoID == alvoObjID && g.playerAtivo.atacandoAgora {
		return
	}
	g.pararAutoAtaque()
	ctx, cancelar := context.WithCancel(context.Background())
	g.playerAtivo.autoAtaqueAlvoID = alvoObjID
	g.playerAtivo.stopAutoAtaque = cancelar
	go g.loopAutoAtaqueNpc(ctx, alvoObjID)
}

func (g *gameClient) pararAutoAtaque() {
	if g.playerAtivo == nil {
		return
	}
	cancelar := g.playerAtivo.stopAutoAtaque
	g.playerAtivo.stopAutoAtaque = nil
	g.playerAtivo.autoAtaqueAlvoID = 0
	if cancelar != nil {
		cancelar()
	}
	if g.playerAtivo.finalizarAutoAtaqueEstado() {
		pacoteFim := montarAutoAttackStopPacket(g.playerAtivo.objID)
		_ = g.enviarPacket(pacoteFim)
		g.broadcastPacoteParaVisiveis(pacoteFim)
	}
}

func (g *gameClient) loopAutoAtaqueNpc(ctx context.Context, alvoObjID int32) {
	inicioEnviado := false
	for {
		if g.playerAtivo == nil {
			g.encerrarLoopAutoAtaque()
			return
		}
		select {
		case <-ctx.Done():
			g.encerrarLoopAutoAtaque()
			return
		default:
		}

		npcGlobal := g.server.mundo.obterNpcPorObjID(alvoObjID)
		if npcGlobal == nil {
			g.encerrarLoopAutoAtaque()
			return
		}
		if !npcGlobal.estaVivo() {
			g.encerrarLoopAutoAtaque()
			return
		}

		distancia := distancia3D(g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z, npcGlobal.x, npcGlobal.y, npcGlobal.z)
		if distancia > distanciaAtaqueBasico {
			if !esperarContextoAtaque(ctx, 500*time.Millisecond) {
				g.encerrarLoopAutoAtaque()
				return
			}
			continue
		}

		template, ok := obterTemplatePersonagemInicial(g.playerAtivo.classID)
		var resultado resultadoAtaqueFisico
		if ok {
			itensPapelBoneca := listarItensPapelBoneca(g.itensAtivos)
			stats := calcularStatsPersonagem(template, g.playerAtivo.nivel, itensPapelBoneca)
			resultado = calcularResultadoAtaquePlayerContraNpc(g.playerAtivo, npcGlobal, stats)
		}
		if !ok {
			resultado = resultadoAtaqueFisico{dano: g.calcularDanoBasico(), errou: false, critico: false, defesaEscudo: "failed", intervaloAtaque: 1200}
		}

		duracaoAtaqueMs := time.Duration(resultado.intervaloAtaque) * time.Millisecond
		preHit := duracaoAtaqueMs / 2
		if preHit <= 0 {
			preHit = 200 * time.Millisecond
		}
		if preHit > duracaoAtaqueMs {
			preHit = duracaoAtaqueMs
		}

		if !inicioEnviado {
			g.playerAtivo.iniciarAutoAtaqueEstado()
			pacoteInicio := montarAutoAttackStartPacket(g.playerAtivo.objID)
			_ = g.enviarPacket(pacoteInicio)
			g.broadcastPacoteParaVisiveis(pacoteInicio)
			inicioEnviado = true
		}

		if !esperarContextoAtaque(ctx, preHit) {
			g.encerrarLoopAutoAtaque()
			return
		}

		npcAtual := g.server.mundo.obterNpcPorObjID(alvoObjID)
		if npcAtual == nil {
			g.encerrarLoopAutoAtaque()
			return
		}
		if !npcAtual.estaVivo() {
			g.encerrarLoopAutoAtaque()
			return
		}
		if distancia3D(g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z, npcAtual.x, npcAtual.y, npcAtual.z) > distanciaAtaqueBasico {
			continue
		}

		dano := resultado.dano
		if resultado.errou {
			dano = 0
		}

		npcAtual.registrarDanoRecebido(g.playerAtivo.objID, dano)
		recompensa := npcAtual.aplicarDano(dano)
		npcAtual.registrarAggro(g.playerAtivo.objID, maximoInt32(dano, 1))
		npcAtual.notificarEventoAi()
		g.playerAtivo.ultimoAtaqueMs = time.Now().UnixMilli()

		pacoteAtaque := montarAttackPacket(g.playerAtivo.objID, npcAtual.objID, dano, g.playerAtivo.x, g.playerAtivo.y, g.playerAtivo.z)
		_ = g.enviarPacket(pacoteAtaque)
		g.broadcastPacoteParaVisiveis(pacoteAtaque)
		for _, cliente := range g.server.mundo.listarPlayersVisiveisParaNpc(npcAtual) {
			if cliente == nil {
				continue
			}
			_ = cliente.enviarPacket(pacoteAtaque)
		}

		if recompensa || npcAtual.deveBroadcastarStatusHp() {
			statusNpc := montarStatusUpdatePacket(npcAtual.objID, [][2]int32{{statusAttrCurHp, npcAtual.hpAtual}, {statusAttrMaxHp, npcAtual.hpMaximo}})
			_ = g.enviarPacket(statusNpc)
			for _, cliente := range g.server.mundo.listarPlayersVisiveisParaNpc(npcAtual) {
				if cliente == nil {
					continue
				}
				_ = cliente.enviarPacket(statusNpc)
			}
		}

		g.enviarMensagensCombatePlayer(resultado)

		if recompensa {
			g.server.distribuirRewardMorteNpcGlobal(npcAtual, g)
			g.server.processarMorteNpcGlobal(npcAtual)
			g.server.mundo.removerNpc(npcAtual.objID)
			g.encerrarLoopAutoAtaque()
			return
		}

		resto := duracaoAtaqueMs - preHit
		if resto < 0 {
			resto = 0
		}
		if !esperarContextoAtaque(ctx, resto) {
			g.encerrarLoopAutoAtaque()
			return
		}
	}
}

func (g *gameClient) encerrarLoopAutoAtaque() {
	if g.playerAtivo == nil {
		return
	}
	g.playerAtivo.stopAutoAtaque = nil
	g.playerAtivo.autoAtaqueAlvoID = 0
	if g.playerAtivo.finalizarAutoAtaqueEstado() {
		pacoteFim := montarAutoAttackStopPacket(g.playerAtivo.objID)
		_ = g.enviarPacket(pacoteFim)
		g.broadcastPacoteParaVisiveis(pacoteFim)
	}
}

func esperarContextoAtaque(ctx context.Context, duracao time.Duration) bool {
	if duracao <= 0 {
		return true
	}
	timer := time.NewTimer(duracao)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
