package network

import "strings"

type npcAiParametros struct {
	setAggressiveTime int32
	halfAggressive    int32
	randomAggressive  int32
	attackLowLevel    int32
	isVs              int32
	attackLowHp       int32
	daggerBackAttack  int32
	canMoveSuperior   int32
	shoutMsg1         int32
	shoutMsg2         int32
	shoutMsg3         int32
	shoutMsg4         int32
	canSeeThrough     bool
	valoresBrutos     map[string]string
}

func novoNpcAiParametros() npcAiParametros {
	return npcAiParametros{setAggressiveTime: 0, valoresBrutos: make(map[string]string)}
}

func (p *npcAiParametros) garantirMapa() {
	if p == nil {
		return
	}
	if p.valoresBrutos != nil {
		return
	}
	p.valoresBrutos = make(map[string]string)
}

func (p *npcAiParametros) mesclar(origem npcAiParametros) {
	if p == nil {
		return
	}
	p.garantirMapa()
	for nome, valor := range origem.valoresBrutos {
		p.aplicar(nome, valor)
	}
}

func (p *npcAiParametros) obterValorBruto(nome string) string {
	if p == nil {
		return ""
	}
	nomeNormalizado := strings.TrimSpace(strings.ToLower(nome))
	if nomeNormalizado == "" {
		return ""
	}
	if p.valoresBrutos == nil {
		return ""
	}
	return p.valoresBrutos[nomeNormalizado]
}

func (p *npcAiParametros) obterStringOuPadrao(nome string, valorPadrao string) string {
	if p == nil {
		return valorPadrao
	}
	valor := strings.TrimSpace(p.obterValorBruto(nome))
	if valor == "" {
		return valorPadrao
	}
	return valor
}

func (p *npcAiParametros) obterInteiroOuPadrao(nome string, valorPadrao int32) int32 {
	if p == nil {
		return valorPadrao
	}
	valor := strings.TrimSpace(p.obterValorBruto(nome))
	if valor == "" {
		return valorPadrao
	}
	return parseInt32Seguro(valor)
}

func (p *npcAiParametros) aplicar(nome string, valor string) {
	if p == nil {
		return
	}
	p.garantirMapa()
	nomeNormalizado := strings.TrimSpace(strings.ToLower(nome))
	valorNormalizado := strings.TrimSpace(valor)
	p.valoresBrutos[nomeNormalizado] = valorNormalizado
	if nomeNormalizado == "setaggressivetime" {
		p.setAggressiveTime = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "halfaggressive" {
		p.halfAggressive = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "randomaggressive" {
		p.randomAggressive = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "attacklowlevel" {
		p.attackLowLevel = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "isvs" {
		p.isVs = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "attacklowhp" {
		p.attackLowHp = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "daggerbackattack" {
		p.daggerBackAttack = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "canmovesuperior" {
		p.canMoveSuperior = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "shoutmsg1" {
		p.shoutMsg1 = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "shoutmsg2" {
		p.shoutMsg2 = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "shoutmsg3" {
		p.shoutMsg3 = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "shoutmsg4" {
		p.shoutMsg4 = parseInt32Seguro(valorNormalizado)
		return
	}
	if nomeNormalizado == "canseethrough" {
		p.canSeeThrough = strings.EqualFold(valorNormalizado, "true")
	}
}
