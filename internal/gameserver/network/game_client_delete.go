package network

import (
	"context"
	"time"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

const diasExcluirPersonagem = 7

func (g *gameClient) processarRequestCharacterDelete(packet *requestCharacterDeletePacket) error {
	logger.Infof("RequestCharacterDelete recebido para conta %s slot=%d", g.conta, packet.slot)
	resultado, err := g.marcarPersonagemParaExcluir(packet.slot)
	if err != nil {
		logger.Errorf("Erro ao marcar exclusao de personagem da conta %s slot=%d: %v", g.conta, packet.slot, err)
		if errEnvio := g.enviarPacket(montarCharDeleteFailPacket(motivoExclusaoFalhou)); errEnvio != nil {
			return errEnvio
		}
		return g.recarregarCharSelect()
	}
	if resultado == motivoExclusaoFalhou {
		if errEnvio := g.enviarPacket(montarCharDeleteFailPacket(motivoExclusaoFalhou)); errEnvio != nil {
			return errEnvio
		}
		return g.recarregarCharSelect()
	}
	if resultado == motivoMembroDeClanNaoPode {
		if errEnvio := g.enviarPacket(montarCharDeleteFailPacket(motivoMembroDeClanNaoPode)); errEnvio != nil {
			return errEnvio
		}
		return g.recarregarCharSelect()
	}
	if resultado == motivoLiderClanNaoPode {
		if errEnvio := g.enviarPacket(montarCharDeleteFailPacket(motivoLiderClanNaoPode)); errEnvio != nil {
			return errEnvio
		}
		return g.recarregarCharSelect()
	}
	if err = g.enviarPacket(montarCharDeleteOkPacket()); err != nil {
		return err
	}
	return g.recarregarCharSelect()
}

func (g *gameClient) processarCharacterRestore(packet *characterRestorePacket) error {
	logger.Infof("CharacterRestore recebido para conta %s slot=%d", g.conta, packet.slot)
	slot, err := g.carregarSlotSelecionado(packet.slot)
	if err != nil {
		return err
	}
	if slot == nil {
		return g.recarregarCharSelect()
	}
	if err = g.server.characterRepo.RestaurarExclusao(context.Background(), slot.ObjID); err != nil {
		logger.Errorf("Erro ao restaurar personagem da conta %s slot=%d objID=%d: %v", g.conta, packet.slot, slot.ObjID, err)
	}
	return g.recarregarCharSelect()
}

func (g *gameClient) marcarPersonagemParaExcluir(indice int32) (uint32, error) {
	slot, err := g.carregarSlotSelecionado(indice)
	if err != nil {
		return motivoExclusaoFalhou, err
	}
	if slot == nil {
		return motivoExclusaoFalhou, nil
	}
	if slot.ClanID != 0 {
		return motivoMembroDeClanNaoPode, nil
	}
	if diasExcluirPersonagem <= 0 {
		err = g.server.characterRepo.DeletarPorObjID(context.Background(), slot.ObjID)
		if err != nil {
			return motivoExclusaoFalhou, err
		}
		return 0, nil
	}
	deleteTime := time.Now().Add(time.Hour * 24 * diasExcluirPersonagem).UnixMilli()
	err = g.server.characterRepo.MarcarParaExcluir(context.Background(), slot.ObjID, deleteTime)
	if err != nil {
		return motivoExclusaoFalhou, err
	}
	return 0, nil
}

func (g *gameClient) recarregarCharSelect() error {
	slots, err := g.server.characterRepo.FindByAccount(context.Background(), g.conta)
	if err != nil {
		return err
	}
	return g.enviarPacket(montarCharSelectInfoPacket(g.conta, g.sessionKey.PlayOkID1, slots))
}
