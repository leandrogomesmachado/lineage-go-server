package network

import (
	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
)

type inventarioEstadoItem uint16

const (
	estadoItemInalterado inventarioEstadoItem = 0
	estadoItemAdicionado inventarioEstadoItem = 1
	estadoItemModificado inventarioEstadoItem = 2
	estadoItemRemovido   inventarioEstadoItem = 3
)

type itemAtualizacaoInventario struct {
	item   gsdb.CharacterItem
	estado inventarioEstadoItem
}

func montarInventoryUpdatePacket(atualizacoes []itemAtualizacaoInventario, augmentacoes []gsdb.CharacterAugmentation) []byte {
	escritor := novoEscritorPacket()
	escritor.escreverC(0x27)
	escritor.escreverH(uint16(len(atualizacoes)))
	for _, atualizacao := range atualizacoes {
		item := atualizacao.item
		escritor.escreverH(uint16(atualizacao.estado))
		escritor.escreverH(0)
		escritor.escreverD(uint32(item.ObjectID))
		escritor.escreverD(uint32(item.ItemID))
		escritor.escreverD(uint32(item.Count))
		escritor.escreverH(0)
		escritor.escreverH(uint16(item.CustomType1))
		equipado := uint16(0)
		if item.Loc == "PAPERDOLL" {
			equipado = 1
		}
		escritor.escreverH(equipado)
		escritor.escreverD(uint32(item.LocData))
		escritor.escreverH(uint16(item.EnchantLevel))
		escritor.escreverH(uint16(item.CustomType2))
		augmentationID := int32(0)
		augmentation := obterAugmentationPorItem(augmentacoes, item.ObjectID)
		if augmentation != nil && augmentation.Attributes > 0 {
			augmentationID = augmentation.Attributes
		}
		escritor.escreverD(uint32(augmentationID))
		mana := item.ManaLeft
		if mana < 0 {
			mana = 0
		}
		escritor.escreverD(uint32(mana))
	}
	return escritor.bytes()
}

func (g *gameClient) agendarAtualizacaoInventario(item gsdb.CharacterItem, estado inventarioEstadoItem) {
	if g == nil {
		return
	}
	if estado == estadoItemInalterado {
		return
	}
	copia := item
	novaAtualizacao := itemAtualizacaoInventario{item: copia, estado: estado}
	g.substituirOuAppendAtualizacaoInventario(novaAtualizacao)
}

func (g *gameClient) enviarInventoryUpdatePendentes() error {
	if g == nil {
		return nil
	}
	tamanho := len(g.atualizacoesInventarioPendentes)
	if tamanho == 0 {
		return nil
	}
	pacote := montarInventoryUpdatePacket(g.atualizacoesInventarioPendentes, g.augmentacoesAtivas)
	g.atualizacoesInventarioPendentes = g.atualizacoesInventarioPendentes[:0]
	return g.enviarPacket(pacote)
}

func (g *gameClient) agendarRemocaoItemInventario(objectID int32) {
	if g == nil {
		return
	}
	novaAtualizacao := itemAtualizacaoInventario{
		item:   gsdb.CharacterItem{ObjectID: objectID},
		estado: estadoItemRemovido,
	}
	g.substituirOuAppendAtualizacaoInventario(novaAtualizacao)
}

func (g *gameClient) substituirOuAppendAtualizacaoInventario(atualizacao itemAtualizacaoInventario) {
	if g == nil {
		return
	}
	indiceExistente := -1
	for idx, atual := range g.atualizacoesInventarioPendentes {
		if atual.item.ObjectID == atualizacao.item.ObjectID {
			indiceExistente = idx
			break
		}
	}
	if indiceExistente >= 0 {
		g.atualizacoesInventarioPendentes[indiceExistente] = atualizacao
		return
	}
	g.atualizacoesInventarioPendentes = append(g.atualizacoesInventarioPendentes, atualizacao)
}
