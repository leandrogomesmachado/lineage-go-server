package network

import (
	"context"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func (n *npcGlobalRuntime) registrarDanoRecebido(alvoObjID int32, dano int32) {
	if n == nil {
		return
	}
	if alvoObjID <= 0 {
		return
	}
	if dano <= 0 {
		return
	}
	if n.danoPorAlvo == nil {
		n.danoPorAlvo = map[int32]int64{}
	}
	n.danoPorAlvo[alvoObjID] += int64(dano)
}

func (n *npcGlobalRuntime) limparDadosReward() {
	if n == nil {
		return
	}
	n.danoPorAlvo = map[int32]int64{}
}

func (g *gameServer) distribuirRewardMorteNpcGlobal(npc *npcGlobalRuntime, matador *gameClient) {
	if g == nil {
		return
	}
	if npc == nil {
		return
	}
	template, ok := obterTemplateNpc(npc.npcID)
	if !ok {
		return
	}
	totalDano := int64(0)
	var principal *gameClient
	maiorDano := int64(0)
	for objID, dano := range npc.danoPorAlvo {
		if dano <= 1 {
			continue
		}
		cliente := g.mundo.obterPorObjID(objID)
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		if cliente.playerAtivo.hpAtual <= 0 {
			continue
		}
		totalDano += dano
		if dano <= maiorDano {
			continue
		}
		maiorDano = dano
		principal = cliente
	}
	if totalDano <= 0 {
		npc.limparDadosReward()
		return
	}
	for objID, dano := range npc.danoPorAlvo {
		cliente := g.mundo.obterPorObjID(objID)
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		if cliente.personagemAtual == nil {
			continue
		}
		if cliente.playerAtivo.hpAtual <= 0 {
			continue
		}
		expBase := int64(float64(template.exp) * (float64(dano) / float64(totalDano)))
		spBase := int32(float64(template.sp) * (float64(dano) / float64(totalDano)))
		expGanha, spGanho := calcularRewardExpSpMob(cliente.playerAtivo.nivel, npc.nivel, expBase, spBase)
		cliente.aplicarRewardExpSp(expGanha, spGanho)
	}
	if principal != nil {
		g.rolarDropNpcGlobal(template, principal)
	}
	_ = matador
	npc.limparDadosReward()
}

func (g *gameServer) rolarDropNpcGlobal(template npcTemplate, vencedor *gameClient) {
	if g == nil {
		return
	}
	if vencedor == nil {
		return
	}
	if vencedor.playerAtivo == nil {
		return
	}
	if vencedor.personagemAtual == nil {
		return
	}
	for _, categoria := range template.drops {
		if categoria.tipo != "DROP" && categoria.tipo != "CURRENCY" {
			continue
		}
		if categoria.tipo == "SPOIL" {
			continue
		}
		if !rolagemPercentualNpc(vencedor.playerAtivo.objID+template.npcID, categoria.chance) {
			continue
		}
		for _, drop := range categoria.drops {
			if !rolagemPercentualNpc(vencedor.playerAtivo.objID+drop.itemID, drop.chance) {
				continue
			}
			quantidade := drop.min
			if drop.max > drop.min {
				quantidade = drop.min + int64((vencedor.playerAtivo.objID+drop.itemID)%int32(drop.max-drop.min+1))
			}
			if quantidade <= 0 {
				quantidade = 1
			}
			vencedor.adicionarItemReward(drop.itemID, quantidade)
			return
		}
	}
}

func calcularRewardExpSpMob(nivelPlayer int32, nivelMob int32, expBase int64, spBase int32) (int64, int32) {
	if expBase < 0 {
		expBase = 0
	}
	if spBase < 0 {
		spBase = 0
	}
	diff := nivelPlayer - nivelMob
	if diff <= 5 {
		return expBase, spBase
	}
	pow := 1.0
	for i := int32(0); i < diff-5; i++ {
		pow *= 5.0 / 6.0
	}
	expFinal := int64(float64(expBase) * pow)
	spFinal := int32(float64(spBase) * pow)
	if expFinal < 0 {
		expFinal = 0
	}
	if spFinal < 0 {
		spFinal = 0
	}
	return expFinal, spFinal
}

func rolagemPercentualNpc(semente int32, chance float64) bool {
	if chance <= 0 {
		return false
	}
	if chance >= 100 {
		return true
	}
	rolagem := float64((normalizarSementeCombate(semente)*37)%10000) / 100.0
	return rolagem <= chance
}

func (g *gameClient) aplicarRewardExpSp(expGanha int64, spGanho int32) {
	if g == nil {
		return
	}
	if g.playerAtivo == nil {
		return
	}
	if g.personagemAtual == nil {
		return
	}
	if expGanha < 0 {
		expGanha = 0
	}
	if spGanho < 0 {
		spGanho = 0
	}
	if expGanha == 0 && spGanho == 0 {
		return
	}
	nivelAntes := g.playerAtivo.nivel
	g.playerAtivo.exp += expGanha
	g.playerAtivo.sp += spGanho
	novo := calcularNivelPorExp(g.playerAtivo.exp)
	subiuNivel := novo > nivelAntes && novo <= nivelMaximoTabela()
	if subiuNivel {
		g.playerAtivo.nivel = novo
	}
	g.personagemAtual.Exp = g.playerAtivo.exp
	g.personagemAtual.Sp = g.playerAtivo.sp
	g.personagemAtual.Level = g.playerAtivo.nivel
	if g.server != nil && g.server.characterRepo != nil {
		err := g.server.characterRepo.AtualizarExpSp(context.Background(), g.playerAtivo.objID, g.playerAtivo.exp, g.playerAtivo.sp, g.playerAtivo.nivel)
		if err != nil {
			logger.Warnf("Falha ao persistir exp/sp do personagem %s objID=%d: %v", g.playerAtivo.nome, g.playerAtivo.objID, err)
		}
	}
	if expGanha > 0 && spGanho > 0 {
		_ = g.enviarPacket(montarSystemMessageDoisNumeros(msgIDYouEarnedS1ExpS2Sp, int32(expGanha), spGanho))
	} else if expGanha > 0 {
		_ = g.enviarPacket(montarSystemMessageNumero(msgIDEarnedS1Experience, int32(expGanha)))
	}
	if subiuNivel {
		_ = g.enviarPacket(montarSystemMessageSimples(msgIDYouIncreasedYourLevel))
		logger.Infof("Level up personagem=%s objID=%d nivel=%d->%d", g.playerAtivo.nome, g.playerAtivo.objID, nivelAntes, novo)
	}
	_ = g.enviarUserInfoAtualizado()
	g.broadcastCharInfoAtualizado()
}

func (g *gameClient) adicionarItemReward(itemID int32, quantidade int64) {
	if g == nil {
		return
	}
	if g.playerAtivo == nil {
		return
	}
	if g.server == nil {
		return
	}
	if g.server.repositorios == nil {
		return
	}
	if g.server.repositorios.CharacterItems == nil {
		return
	}
	item, err := g.server.repositorios.CharacterItems.InserirOuSomarItem(context.Background(), g.playerAtivo.objID, itemID, quantidade)
	if err != nil {
		logger.Warnf("Falha ao inserir drop itemID=%d qtd=%d personagem=%s objID=%d: %v", itemID, quantidade, g.playerAtivo.nome, g.playerAtivo.objID, err)
		return
	}
	if item != nil {
		g.atualizarItemAtivoReward(*item)
	}
	logger.Infof("Reward drop personagem=%s objID=%d itemID=%d qtd=%d", g.playerAtivo.nome, g.playerAtivo.objID, itemID, quantidade)
	_ = g.enviarPacket(montarItemListPacket(g.itensAtivos, g.augmentacoesAtivas))
}

func (g *gameClient) atualizarItemAtivoReward(item gsdb.CharacterItem) {
	if g == nil {
		return
	}
	for indice := range g.itensAtivos {
		if g.itensAtivos[indice].ObjectID != item.ObjectID {
			continue
		}
		g.itensAtivos[indice] = item
		return
	}
	g.itensAtivos = append(g.itensAtivos, item)
}
