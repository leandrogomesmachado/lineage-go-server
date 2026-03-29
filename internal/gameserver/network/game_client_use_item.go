package network

import (
	"context"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func (g *gameClient) processarRequestUseItem(packet *requestUseItemPacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	itemIdx := g.buscarIndiceItemPorObjID(packet.objectID)
	if itemIdx < 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	item := g.itensAtivos[itemIdx]
	slots := resolverSlotsEquipamento(item.ItemID)
	if len(slots) == 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if item.Loc == "PAPERDOLL" {
		return g.desequiparItem(itemIdx)
	}
	return g.equiparItem(itemIdx)
}

func (g *gameClient) processarRequestUnEquipItem(packet *requestUnEquipItemPacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	itemIdx := g.buscarIndiceItemPorSlot(packet.slot)
	if itemIdx < 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	return g.desequiparItem(itemIdx)
}

func (g *gameClient) equiparItem(itemIdx int) error {
	ctx := context.Background()
	item := g.itensAtivos[itemIdx]
	bodypart, ok := obterBodypartEquip(item.ItemID)
	if !ok {
		logger.Warnf("Bodypart nao encontrado para item=%d", item.ItemID)
		return g.enviarPacket(montarActionFailedPacket())
	}
	slots := resolverSlotsEquipamento(item.ItemID)
	if len(slots) == 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	slotDestino := g.selecionarSlotEquipamento(itemIdx, slots)
	if slotDestino < 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	slotsParaLiberar := g.identificarSlotsParaLiberar(bodypart, slotDestino, slots)
	for _, slot := range slotsParaLiberar {
		indice := g.buscarIndiceItemPorSlot(slot)
		if indice < 0 {
			continue
		}
		if indice == itemIdx {
			continue
		}
		err := g.moverItemParaInventario(ctx, indice)
		if err != nil {
			return g.enviarPacket(montarActionFailedPacket())
		}
	}
	err := g.moverItemParaSlot(ctx, itemIdx, slotDestino)
	if err != nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	return g.enviarAtualizacaoEquipamento()
}

func (g *gameClient) desequiparItem(itemIdx int) error {
	ctx := context.Background()
	err := g.moverItemParaInventario(ctx, itemIdx)
	if err != nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	return g.enviarAtualizacaoEquipamento()
}

func (g *gameClient) enviarAtualizacaoEquipamento() error {
	g.sincronizarPersonagemAtualComPlayerAtivo()
	g.recalcularStatsComEquipamentos()
	if err := g.enviarStatusAtualizado(); err != nil {
		return err
	}
	if err := g.enviarInventoryUpdatePendentes(); err != nil {
		return err
	}
	if err := g.enviarUserInfoAtualizado(); err != nil {
		return err
	}
	g.broadcastCharInfoAtualizado()
	return nil
}

func (g *gameClient) selecionarSlotEquipamento(itemIdx int, slots []int32) int32 {
	item := g.itensAtivos[itemIdx]
	bodypart, ok := obterBodypartEquip(item.ItemID)
	if !ok {
		return -1
	}
	if len(slots) == 1 {
		return slots[0]
	}
	if len(slots) != 2 {
		return slots[0]
	}
	primeiro := slots[0]
	segundo := slots[1]
	indicePrimeiro := g.buscarIndiceItemPorSlot(primeiro)
	if indicePrimeiro < 0 {
		return primeiro
	}
	indiceSegundo := g.buscarIndiceItemPorSlot(segundo)
	if indiceSegundo < 0 {
		return segundo
	}
	itemPrimeiro := g.itensAtivos[indicePrimeiro]
	itemSegundo := g.itensAtivos[indiceSegundo]
	if itemSegundo.ItemID == item.ItemID {
		return primeiro
	}
	if itemPrimeiro.ItemID == item.ItemID {
		return segundo
	}
	if bodypart == "rear;lear" {
		return primeiro
	}
	if bodypart == "rfinger;lfinger" {
		return primeiro
	}
	return primeiro
}

func (g *gameClient) identificarSlotsParaLiberar(bodypart string, slotDestino int32, slotsPossiveis []int32) []int32 {
	resultado := make([]int32, 0, 4)
	resultado = adicionarSlotSeNaoExiste(resultado, slotDestino)
	if bodypart == "lrhand" {
		resultado = adicionarSlotSeNaoExiste(resultado, 8)
	}
	if bodypart == "fullarmor" {
		resultado = adicionarSlotSeNaoExiste(resultado, 11)
	}
	if bodypart == "hairall" {
		resultado = adicionarSlotSeNaoExiste(resultado, 14)
	}
	if bodypart == "lhand" {
		indiceDireita := g.buscarIndiceItemPorSlot(7)
		if indiceDireita >= 0 {
			bodypartDireita, ok := obterBodypartEquip(g.itensAtivos[indiceDireita].ItemID)
			if ok && bodypartDireita == "lrhand" {
				resultado = adicionarSlotSeNaoExiste(resultado, 7)
			}
		}
	}
	if bodypart == "face" {
		indiceHair := g.buscarIndiceItemPorSlot(15)
		if indiceHair >= 0 {
			bodypartHair, ok := obterBodypartEquip(g.itensAtivos[indiceHair].ItemID)
			if ok && bodypartHair == "hairall" {
				resultado = adicionarSlotSeNaoExiste(resultado, 15)
			}
		}
	}
	if bodypart == "hair" {
		indiceFace := g.buscarIndiceItemPorSlot(14)
		if indiceFace >= 0 {
			bodypartFace, ok := obterBodypartEquip(g.itensAtivos[indiceFace].ItemID)
			if ok && bodypartFace == "hairall" {
				resultado = adicionarSlotSeNaoExiste(resultado, 14)
			}
		}
	}
	if bodypart == "alldress" {
		resultado = adicionarSlotSeNaoExiste(resultado, 11)
		resultado = adicionarSlotSeNaoExiste(resultado, 8)
		resultado = adicionarSlotSeNaoExiste(resultado, 7)
		resultado = adicionarSlotSeNaoExiste(resultado, 6)
		resultado = adicionarSlotSeNaoExiste(resultado, 12)
		resultado = adicionarSlotSeNaoExiste(resultado, 9)
	}
	return resultado
}

func adicionarSlotSeNaoExiste(slots []int32, slot int32) []int32 {
	for _, existente := range slots {
		if existente == slot {
			return slots
		}
	}
	return append(slots, slot)
}

func (g *gameClient) moverItemParaInventario(ctx context.Context, idx int) error {
	itens := g.itensAtivos
	item := &itens[idx]
	if item.Loc == "INVENTORY" {
		return nil
	}
	err := g.server.repositorios.CharacterItems.AtualizarLocItem(ctx, item.ObjectID, "INVENTORY", 0)
	if err != nil {
		logger.Warnf("Erro ao mover item para INVENTORY objID=%d: %v", item.ObjectID, err)
		return err
	}
	item.Loc = "INVENTORY"
	item.LocData = 0
	g.itensAtivos[idx] = *item
	g.agendarAtualizacaoInventario(*item, estadoItemModificado)
	return nil
}

func (g *gameClient) moverItemParaSlot(ctx context.Context, idx int, slot int32) error {
	itens := g.itensAtivos
	item := &itens[idx]
	if item.Loc == "PAPERDOLL" && item.LocData == slot {
		return nil
	}
	err := g.server.repositorios.CharacterItems.AtualizarLocItem(ctx, item.ObjectID, "PAPERDOLL", slot)
	if err != nil {
		logger.Warnf("Erro ao atualizar slot para item objID=%d slot=%d: %v", item.ObjectID, slot, err)
		return err
	}
	item.Loc = "PAPERDOLL"
	item.LocData = slot
	g.itensAtivos[idx] = *item
	g.agendarAtualizacaoInventario(*item, estadoItemModificado)
	return nil
}

func (g *gameClient) recalcularStatsComEquipamentos() {
	if g.personagemAtual == nil || g.playerAtivo == nil {
		return
	}
	template, ok := obterTemplatePersonagemInicial(g.playerAtivo.classID)
	if !ok {
		return
	}
	itensPapelBoneca := listarItensPapelBoneca(g.itensAtivos)
	stats := calcularStatsPersonagem(template, g.playerAtivo.nivel, itensPapelBoneca)
	g.personagemAtual.MaxHp = stats.hpMaximo
	g.personagemAtual.MaxMp = stats.mpMaximo
	g.personagemAtual.MaxCp = stats.cpMaximo
	g.playerAtivo.hpMaximo = stats.hpMaximo
	g.playerAtivo.mpMaximo = stats.mpMaximo
	g.playerAtivo.cpMaximo = stats.cpMaximo
	if g.playerAtivo.hpAtual > g.playerAtivo.hpMaximo {
		g.playerAtivo.hpAtual = g.playerAtivo.hpMaximo
	}
	if g.playerAtivo.mpAtual > g.playerAtivo.mpMaximo {
		g.playerAtivo.mpAtual = g.playerAtivo.mpMaximo
	}
	if g.playerAtivo.cpAtual > g.playerAtivo.cpMaximo {
		g.playerAtivo.cpAtual = g.playerAtivo.cpMaximo
	}
}

func (g *gameClient) enviarStatusAtualizado() error {
	if g == nil {
		return nil
	}
	if g.playerAtivo == nil {
		return nil
	}
	atributos := make([][2]int32, 0, 6)
	atributos = append(atributos, [2]int32{statusAttrCurHp, g.playerAtivo.hpAtual})
	atributos = append(atributos, [2]int32{statusAttrMaxHp, g.playerAtivo.hpMaximo})
	atributos = append(atributos, [2]int32{statusAttrCurMp, g.playerAtivo.mpAtual})
	atributos = append(atributos, [2]int32{statusAttrMaxMp, g.playerAtivo.mpMaximo})
	atributos = append(atributos, [2]int32{statusAttrCurCp, g.playerAtivo.cpAtual})
	atributos = append(atributos, [2]int32{statusAttrMaxCp, g.playerAtivo.cpMaximo})
	pacote := montarStatusUpdatePacket(g.playerAtivo.objID, atributos)
	return g.enviarPacket(pacote)
}

func (g *gameClient) buscarIndiceItemPorObjID(objectID int32) int {
	for i := range g.itensAtivos {
		if g.itensAtivos[i].ObjectID == objectID {
			return i
		}
	}
	return -1
}

func (g *gameClient) buscarIndiceItemPorSlot(slot int32) int {
	for i := range g.itensAtivos {
		item := g.itensAtivos[i]
		if item.Loc == "PAPERDOLL" && item.LocData == slot {
			return i
		}
	}
	return -1
}

func obterSlotEquipadoItem(itens []gsdb.CharacterItem, objectID int32) int32 {
	for _, item := range itens {
		if item.ObjectID == objectID && item.Loc == "PAPERDOLL" {
			return item.LocData
		}
	}
	return -1
}
