package network

import "time"

const periodoRegenPlayerMs = 3000
const tempoSemCombateRegenMs = int64(15000)

func (g *gameClient) iniciarRegenPlayer() {
	go func() {
		ticker := time.NewTicker(periodoRegenPlayerMs * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-g.canalEncerramento:
				return
			case <-ticker.C:
				g.processarRegenPlayer()
			}
		}
	}()
}

func (g *gameClient) processarRegenPlayer() {
	if g.playerAtivo == nil {
		return
	}
	agora := time.Now().UnixMilli()
	if g.playerAtivo.ultimoAtaqueMs > 0 && agora-g.playerAtivo.ultimoAtaqueMs < tempoSemCombateRegenMs {
		return
	}
	alterou := false
	if g.playerAtivo.hpAtual > 0 && g.playerAtivo.hpAtual < g.playerAtivo.hpMaximo {
		regenHp := maximoInt32(1, g.playerAtivo.hpMaximo/20)
		g.playerAtivo.hpAtual += regenHp
		if g.playerAtivo.hpAtual > g.playerAtivo.hpMaximo {
			g.playerAtivo.hpAtual = g.playerAtivo.hpMaximo
		}
		alterou = true
	}
	if g.playerAtivo.mpAtual < g.playerAtivo.mpMaximo {
		regenMp := maximoInt32(1, g.playerAtivo.mpMaximo/20)
		g.playerAtivo.mpAtual += regenMp
		if g.playerAtivo.mpAtual > g.playerAtivo.mpMaximo {
			g.playerAtivo.mpAtual = g.playerAtivo.mpMaximo
		}
		alterou = true
	}
	if !alterou {
		return
	}
	statusUpdate := montarStatusUpdatePacket(g.playerAtivo.objID, [][2]int32{
		{statusAttrCurHp, g.playerAtivo.hpAtual},
		{statusAttrMaxHp, g.playerAtivo.hpMaximo},
		{statusAttrCurMp, g.playerAtivo.mpAtual},
		{statusAttrMaxMp, g.playerAtivo.mpMaximo},
	})
	_ = g.enviarPacket(statusUpdate)
}
