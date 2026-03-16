package network

import "math"

type resultadoAtaqueFisico struct {
	dano            int32
	errou           bool
	critico         bool
	defesaEscudo    string
	intervaloAtaque int64
}

func calcularResultadoAtaquePlayerContraNpc(player *playerAtivo, npc *npcGlobalRuntime, stats statsCalculadasPersonagem) resultadoAtaqueFisico {
	if player == nil || npc == nil {
		return resultadoAtaqueFisico{dano: 1, errou: false, critico: false, defesaEscudo: "failed", intervaloAtaque: 1200}
	}
	acuracia := stats.precisao
	evasao := calcularEvasaoAproximada(npc.nivel, 30)
	deltaZ := player.z - npc.z
	acertou := calcularAcertoFisico(acuracia, evasao, deltaZ, false, true, player.objID+npc.objID)
	if !acertou {
		return resultadoAtaqueFisico{dano: 0, errou: true, critico: false, defesaEscudo: "failed", intervaloAtaque: calcularIntervaloAtaqueFisico(stats.pAtkSpd)}
	}
	chanceCrit := calcularChanceCriticaAproximada(stats.critico)
	critico := calcularCriticoFisico(chanceCrit, player.objID+npc.objID+13)
	defesaEscudo := "failed"
	dano := calcularDanoFisicoBase(stats.pAtk, npc.pDef, critico, defesaEscudo, player.objID+npc.objID+29)
	return resultadoAtaqueFisico{dano: dano, errou: false, critico: critico, defesaEscudo: defesaEscudo, intervaloAtaque: calcularIntervaloAtaqueFisico(stats.pAtkSpd)}
}

func calcularResultadoAtaqueNpcContraPlayer(npc *npcGlobalRuntime, player *playerAtivo, stats statsCalculadasPersonagem) resultadoAtaqueFisico {
	if npc == nil || player == nil {
		return resultadoAtaqueFisico{dano: 1, errou: false, critico: false, defesaEscudo: "failed", intervaloAtaque: 1200}
	}
	acuracia := calcularPrecisaoAproximada(npc.nivel, 30)
	evasao := stats.evasao
	deltaZ := npc.z - player.z
	acertou := calcularAcertoFisico(acuracia, evasao, deltaZ, false, true, npc.objID+player.objID)
	if !acertou {
		return resultadoAtaqueFisico{dano: 0, errou: true, critico: false, defesaEscudo: "failed", intervaloAtaque: calcularIntervaloAtaqueFisico(npc.pAtkSpd)}
	}
	chanceCrit := calcularChanceCriticaAproximada(npc.crit)
	critico := calcularCriticoFisico(chanceCrit, npc.objID+player.objID+17)
	defesaEscudo := calcularDefesaEscudo(25, critico, npc.objID+player.objID+31)
	dano := calcularDanoFisicoBase(npc.pAtk, stats.pDef, critico, defesaEscudo, npc.objID+player.objID+43)
	return resultadoAtaqueFisico{dano: dano, errou: false, critico: critico, defesaEscudo: defesaEscudo, intervaloAtaque: calcularIntervaloAtaqueFisico(npc.pAtkSpd)}
}

func calcularIntervaloAtaqueFisico(pAtkSpd int32) int64 {
	intervaloMs := int64(1200)
	if pAtkSpd > 0 {
		intervaloMs = int64(500000 / pAtkSpd)
	}
	if intervaloMs < 400 {
		return 400
	}
	return intervaloMs
}

func calcularPrecisaoAproximada(nivel int32, dexBase int32) int32 {
	base := int32(math.Round(math.Sqrt(float64(maximoInt32(dexBase, 1))) * 6))
	precisao := base + nivel
	if precisao < 1 {
		return 1
	}
	return precisao
}

func calcularEvasaoAproximada(nivel int32, dexBase int32) int32 {
	base := int32(math.Round(math.Sqrt(float64(maximoInt32(dexBase, 1))) * 6))
	evasao := base + nivel/3
	if evasao < 1 {
		return 1
	}
	return evasao
}

func calcularChanceCriticaAproximada(critBase int32) int32 {
	chance := critBase * 10
	if chance < 1 {
		return 1
	}
	if chance > 500 {
		return 500
	}
	return chance
}

func calcularAcertoFisico(acuracia int32, evasao int32, deltaZ int32, atacanteAtras bool, atacanteLado bool, semente int32) bool {
	diff := acuracia - evasao
	if deltaZ > 50 {
		diff += 3
	}
	if deltaZ < -50 {
		diff -= 3
	}
	if atacanteAtras {
		diff += 10
	}
	if atacanteLado {
		diff += 5
	}
	chance := (90 + (2 * diff)) * 10
	if chance < 300 {
		chance = 300
	}
	if chance > 980 {
		chance = 980
	}
	rolagem := normalizarSementeCombate(semente) % 1000
	return rolagem < chance
}

func calcularCriticoFisico(chance int32, semente int32) bool {
	if chance < 1 {
		return false
	}
	if chance > 1000 {
		chance = 1000
	}
	rolagem := normalizarSementeCombate(semente) % 1000
	return rolagem < chance
}

func calcularDefesaEscudo(acuraciaEscudo int32, critico bool, semente int32) string {
	chanceBase := acuraciaEscudo
	if chanceBase <= 0 {
		return "failed"
	}
	if critico {
		chanceBase *= 3
	}
	rolagem := normalizarSementeCombate(semente) % 100
	if rolagem < 5 {
		return "perfect"
	}
	if rolagem < chanceBase {
		return "success"
	}
	return "failed"
}

func calcularDanoFisicoBase(pAtk int32, pDef int32, critico bool, defesaEscudo string, variacaoBase int32) int32 {
	ataque := maximoInt32(pAtk, 1)
	defesa := maximoInt32(pDef, 1)
	defesaEfetiva := float64(defesa)
	if defesaEscudo == "success" {
		defesaEfetiva += math.Max(5, float64(defesa)/2)
	}
	if defesaEscudo == "perfect" {
		return 1
	}
	rnd := 0.95 + float64(normalizarSementeCombate(variacaoBase)%11)/100.0
	posMul := 1.0
	if critico {
		posMul = 2.0
	}
	dano := (float64(ataque) * posMul * rnd) * 77.0 / defesaEfetiva
	if dano < 1 {
		return 1
	}
	return int32(math.Round(dano))
}

func normalizarSementeCombate(semente int32) int32 {
	if semente < 0 {
		semente *= -1
	}
	return semente + 1
}

func maximoInt32(a int32, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
