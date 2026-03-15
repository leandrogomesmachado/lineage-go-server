package network

import "math"

const maxStatValueFormula = 100

var bonusStrFormula = gerarBonusFormula(1.036, 34.845)
var bonusIntFormula = gerarBonusFormula(1.020, 31.375)
var bonusDexFormula = gerarBonusFormula(1.009, 19.360)
var bonusWitFormula = gerarBonusFormula(1.050, 20.000)
var bonusConFormula = gerarBonusFormula(1.030, 27.632)
var bonusMenFormula = gerarBonusFormula(1.010, -0.060)
var baseEvasaoPrecisaoFormula = gerarBaseEvasaoPrecisaoFormula()

func gerarBonusFormula(base float64, expoenteBase float64) []float64 {
	resultado := make([]float64, maxStatValueFormula)
	for indice := range resultado {
		valor := math.Pow(base, float64(indice)-expoenteBase)
		resultado[indice] = math.Floor(valor*100+0.5) / 100
	}
	return resultado
}

func gerarBaseEvasaoPrecisaoFormula() []float64 {
	resultado := make([]float64, maxStatValueFormula)
	for indice := range resultado {
		resultado[indice] = math.Sqrt(float64(indice)) * 6
	}
	return resultado
}

func obterBonusFormula(tabela []float64, valor int32) float64 {
	if valor < 0 {
		return tabela[0]
	}
	if int(valor) < len(tabela) {
		return tabela[valor]
	}
	return tabela[len(tabela)-1]
}

func obterLevelModFormula(nivel int32) float64 {
	return (100.0 - 11.0 + float64(nivel)) / 100.0
}

type statsCalculadasPersonagem struct {
	hpMaximo int32
	mpMaximo int32
	cpMaximo int32
	pAtk     int32
	mAtk     int32
	pDef     int32
	mDef     int32
	pAtkSpd  int32
	mAtkSpd  int32
	runSpd   int32
	walkSpd  int32
	swimSpd  int32
	evasao   int32
	precisao int32
	critico  int32
}

func calcularStatsPersonagem(template templatePersonagemInicial, nivel int32, itens []itemPapelBoneca) statsCalculadasPersonagem {
	levelMod := obterLevelModFormula(nivel)
	bonusCon := obterBonusFormula(bonusConFormula, template.con)
	bonusMen := obterBonusFormula(bonusMenFormula, template.men)
	bonusStr := obterBonusFormula(bonusStrFormula, template.str)
	bonusInt := obterBonusFormula(bonusIntFormula, template.intel)
	bonusDex := obterBonusFormula(bonusDexFormula, template.dex)
	bonusWit := obterBonusFormula(bonusWitFormula, template.wit)
	basePDef := float64(template.pDef) - obterReducaoPDefEquipado(itens, template)
	if basePDef < 1 {
		basePDef = 1
	}
	baseMDef := float64(template.mDef) - obterReducaoMDefEquipado(itens)
	if baseMDef < 1 {
		baseMDef = 1
	}
	basePAtkSpd := float64(statsCalculadasBasePAtkSpd(template, itens))
	baseMAtkSpd := float64(template.mAtkSpd)
	if baseMAtkSpd < 1 {
		baseMAtkSpd = 333
	}
	resultado := statsCalculadasPersonagem{
		hpMaximo: int32(math.Round(float64(template.obterHpMaximoPorNivel(nivel)) * bonusCon)),
		mpMaximo: int32(math.Round(float64(template.obterMpMaximoPorNivel(nivel)) * bonusMen)),
		cpMaximo: int32(math.Round(float64(template.obterCpMaximoPorNivel(nivel)) * bonusCon)),
		pAtk:     int32(math.Round(float64(template.pAtk) * bonusStr * levelMod)),
		mAtk:     int32(math.Round(float64(template.mAtk) * (levelMod * levelMod) * (bonusInt * bonusInt))),
		pDef:     int32(math.Round(basePDef * levelMod)),
		mDef:     int32(math.Round(baseMDef * bonusMen * levelMod)),
		pAtkSpd:  int32(math.Round(basePAtkSpd * bonusDex)),
		mAtkSpd:  int32(math.Round(baseMAtkSpd * bonusWit)),
		runSpd:   int32(math.Round(float64(template.runSpd) * bonusDex)),
		walkSpd:  int32(math.Round(float64(template.walkSpd) * bonusDex)),
		swimSpd:  int32(math.Round(float64(template.swimSpd) * bonusDex)),
		evasao:   int32(math.Round(baseEvasaoPrecisaoFormula[template.dex])),
		precisao: int32(math.Round(baseEvasaoPrecisaoFormula[template.dex] + float64(nivel))),
		critico:  int32(math.Round(float64(template.baseCrit) * bonusDex * 10.0)),
	}
	if resultado.hpMaximo < 1 {
		resultado.hpMaximo = 1
	}
	if resultado.mpMaximo < 1 {
		resultado.mpMaximo = 1
	}
	if resultado.cpMaximo < 0 {
		resultado.cpMaximo = 0
	}
	if resultado.pAtk < 1 {
		resultado.pAtk = 1
	}
	if resultado.mAtk < 1 {
		resultado.mAtk = 1
	}
	if resultado.pDef < 1 {
		resultado.pDef = 1
	}
	if resultado.mDef < 1 {
		resultado.mDef = 1
	}
	if resultado.pAtkSpd < 1 {
		resultado.pAtkSpd = 1
	}
	if resultado.mAtkSpd < 1 {
		resultado.mAtkSpd = 1
	}
	if resultado.runSpd < 1 {
		resultado.runSpd = 1
	}
	if resultado.walkSpd < 1 {
		resultado.walkSpd = 1
	}
	if resultado.swimSpd < 1 {
		resultado.swimSpd = 1
	}
	if resultado.critico < 1 {
		resultado.critico = 1
	}
	return resultado
}

type itemPapelBoneca struct {
	slotPaperdoll int32
	itemID        int32
}

func obterReducaoPDefEquipado(itens []itemPapelBoneca, template templatePersonagemInicial) float64 {
	temCabeca := false
	temPeitoral := false
	temPernas := false
	temLuvas := false
	temBotas := false
	for _, item := range itens {
		if item.slotPaperdoll == 6 {
			temCabeca = true
			continue
		}
		if item.slotPaperdoll == 10 {
			temPeitoral = true
			continue
		}
		if item.slotPaperdoll == 11 {
			temPernas = true
			continue
		}
		if item.slotPaperdoll == 9 {
			temLuvas = true
			continue
		}
		if item.slotPaperdoll == 12 {
			temBotas = true
		}
	}
	reducao := 0.0
	if temCabeca {
		reducao += 12
	}
	if temPeitoral {
		reducao += 31
	}
	if temPernas {
		reducao += 18
	}
	if temLuvas {
		reducao += 8
	}
	if temBotas {
		reducao += 7
	}
	return reducao
}

func obterReducaoMDefEquipado(itens []itemPapelBoneca) float64 {
	temAnelEsquerdo := false
	temAnelDireito := false
	temBrincoEsquerdo := false
	temBrincoDireito := false
	temColar := false
	for _, item := range itens {
		if item.slotPaperdoll == 5 {
			temAnelDireito = true
			continue
		}
		if item.slotPaperdoll == 4 {
			temAnelEsquerdo = true
			continue
		}
		if item.slotPaperdoll == 2 {
			temBrincoDireito = true
			continue
		}
		if item.slotPaperdoll == 1 {
			temBrincoEsquerdo = true
			continue
		}
		if item.slotPaperdoll == 3 {
			temColar = true
		}
	}
	reducao := 0.0
	if temAnelEsquerdo {
		reducao += 5
	}
	if temAnelDireito {
		reducao += 5
	}
	if temBrincoEsquerdo {
		reducao += 9
	}
	if temBrincoDireito {
		reducao += 9
	}
	if temColar {
		reducao += 13
	}
	return reducao
}

func obterBaseAtkSpdFisico(itens []itemPapelBoneca) int32 {
	for _, item := range itens {
		if item.slotPaperdoll != 7 {
			continue
		}
		if item.itemID == 1147 {
			return 379
		}
	}
	return 300
}
