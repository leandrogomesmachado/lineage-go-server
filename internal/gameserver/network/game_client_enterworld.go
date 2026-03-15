package network

import (
	"context"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func (g *gameClient) processarRequestGameStart(packet *requestGameStartPacket) error {
	logger.Infof("RequestGameStart recebido para conta %s slot=%d", g.conta, packet.slot)
	slot, err := g.carregarSlotSelecionado(packet.slot)
	if err != nil {
		logger.Errorf("Erro ao carregar slot %d da conta %s: %v", packet.slot, g.conta, err)
		return g.enviarPacket(montarActionFailedPacket())
	}
	if slot == nil {
		logger.Warnf("Slot %d invalido para conta %s", packet.slot, g.conta)
		return g.enviarPacket(montarActionFailedPacket())
	}
	if slot.AccessLevel < 0 {
		logger.Warnf("Slot %d inacessivel para conta %s", packet.slot, g.conta)
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.personagemAtual = slot
	g.playerAtivo = nil
	g.slotSelecionado = packet.slot
	g.estado = estadoEntering
	if err = g.enviarPacket(montarSSQInfoPacket()); err != nil {
		return err
	}
	return g.enviarPacket(montarCharSelectedPacket(g.sessionKey.PlayOkID1, *slot))
}

func (g *gameClient) normalizarStatusPersonagemPorTemplate() {
	if g.personagemAtual == nil {
		return
	}
	template, ok := obterTemplatePersonagemInicial(g.personagemAtual.ClassID)
	if !ok {
		return
	}
	nivel := g.personagemAtual.Level
	if nivel <= 0 {
		nivel = 1
		g.personagemAtual.Level = 1
	}
	statsCalculadas := calcularStatsPersonagem(template, nivel, []itemPapelBoneca{})
	hpMaximo := statsCalculadas.hpMaximo
	mpMaximo := statsCalculadas.mpMaximo
	cpMaximo := statsCalculadas.cpMaximo
	if hpMaximo > 0 {
		g.personagemAtual.MaxHp = hpMaximo
	}
	if mpMaximo > 0 {
		g.personagemAtual.MaxMp = mpMaximo
	}
	if cpMaximo > 0 {
		g.personagemAtual.MaxCp = cpMaximo
	}
	if g.personagemAtual.CurHp <= 0 || g.personagemAtual.CurHp > g.personagemAtual.MaxHp {
		g.personagemAtual.CurHp = g.personagemAtual.MaxHp
	}
	if g.personagemAtual.CurMp <= 0 || g.personagemAtual.CurMp > g.personagemAtual.MaxMp {
		g.personagemAtual.CurMp = g.personagemAtual.MaxMp
	}
	if g.personagemAtual.CurCp < 0 || g.personagemAtual.CurCp > g.personagemAtual.MaxCp {
		g.personagemAtual.CurCp = g.personagemAtual.MaxCp
	}
}

func (g *gameClient) carregarDadosAuxiliaresPersonagem() {
	if g.personagemAtual == nil {
		return
	}
	if g.server.repositorios == nil {
		return
	}
	classIndex := int32(0)
	if g.personagemAtual.BaseClass > 0 && g.personagemAtual.BaseClass != g.personagemAtual.ClassID {
		classIndex = 1
	}
	ctx := context.Background()
	hennas, errHennas := g.server.repositorios.CharacterHennas.ListarPorPersonagem(ctx, g.personagemAtual.ObjID, classIndex)
	if errHennas == nil {
		g.hennasAtivas = hennas
	}
	if errHennas != nil {
		logger.Warnf("Falha ao carregar hennas do personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, errHennas)
	}
	skills, errSkills := g.server.repositorios.CharacterSkills.ListarPorPersonagem(ctx, g.personagemAtual.ObjID, classIndex)
	if errSkills == nil {
		g.skillsAtivas = skills
	}
	if errSkills != nil {
		logger.Warnf("Falha ao carregar skills do personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, errSkills)
	}
	if len(g.skillsAtivas) == 0 {
		g.skillsAtivas = listarSkillsIniciaisClasse(g.personagemAtual.ClassID, g.personagemAtual.Level)
		for indice := range g.skillsAtivas {
			g.skillsAtivas[indice].CharObjID = g.personagemAtual.ObjID
		}
	}
	atalhos, errAtalhos := g.server.repositorios.CharacterShortcuts.ListarPorPersonagem(ctx, g.personagemAtual.ObjID, classIndex)
	if errAtalhos == nil {
		g.atalhosAtivos = atalhos
	}
	if errAtalhos != nil {
		logger.Warnf("Falha ao carregar atalhos do personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, errAtalhos)
	}
	subclasses, errSubclasses := g.server.repositorios.CharacterSubclasses.ListarPorPersonagem(ctx, g.personagemAtual.ObjID)
	if errSubclasses == nil {
		g.subclassesAtivas = subclasses
	}
	if errSubclasses != nil {
		logger.Warnf("Falha ao carregar subclasses do personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, errSubclasses)
	}
	itens, errItens := g.server.repositorios.CharacterItems.ListarPorPersonagem(ctx, g.personagemAtual.ObjID)
	if errItens == nil {
		g.itensAtivos = itens
	}
	if errItens != nil {
		logger.Warnf("Falha ao carregar itens do personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, errItens)
	}
	if errItens == nil && g.server.repositorios.CharacterAugments != nil {
		objectIDs := make([]int32, 0, len(itens))
		for _, item := range itens {
			objectIDs = append(objectIDs, item.ObjectID)
		}
		augmentacoes, errAugment := g.server.repositorios.CharacterAugments.ListarPorItens(ctx, objectIDs)
		if errAugment == nil {
			g.augmentacoesAtivas = augmentacoes
		}
		if errAugment != nil {
			logger.Warnf("Falha ao carregar augmentations do personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, errAugment)
		}
	}
}

func (g *gameClient) processarEnterWorld(packet *enterWorldPacket) error {
	_ = packet
	if g.personagemAtual == nil {
		logger.Warnf("EnterWorld recebido sem personagem selecionado para conta %s", g.conta)
		return g.enviarPacket(montarActionFailedPacket())
	}
	g.normalizarStatusPersonagemPorTemplate()
	g.garantirSpawnSeguro()
	g.playerAtivo = novoPlayerAtivo(g.conta, *g.personagemAtual)
	g.inicializarTrainerPessoal()
	g.carregarDadosAuxiliaresPersonagem()
	g.sincronizarSkillsAutoLearn()
	if err := g.server.characterRepo.AtualizarLastAccess(context.Background(), g.personagemAtual.ObjID); err != nil {
		logger.Warnf("Falha ao atualizar lastAccess do personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, err)
	}
	g.estado = estadoInGame
	g.server.mundo.registrar(g)
	logger.Infof("EnterWorld recebido para conta %s personagem=%s objID=%d", g.conta, g.personagemAtual.CharName, g.personagemAtual.ObjID)
	if err := g.enviarPacket(montarSkillListPacket(g.skillsAtivas)); err != nil {
		return err
	}
	if err := g.enviarPacket(montarExStorageMaxCountPacket()); err != nil {
		return err
	}
	if err := g.enviarPacket(montarHennaInfoPacket(*g.personagemAtual, g.hennasAtivas)); err != nil {
		return err
	}
	if err := g.enviarPacket(montarEtcStatusUpdatePacket()); err != nil {
		return err
	}
	if err := g.enviarUserInfoAtualizado(); err != nil {
		return err
	}
	if err := g.enviarPacket(montarItemListPacket(g.itensAtivos, g.augmentacoesAtivas)); err != nil {
		return err
	}
	if err := g.enviarPacket(montarShortCutInitPacket(g.atalhosAtivos, g.itensAtivos, g.augmentacoesAtivas)); err != nil {
		return err
	}
	if err := g.enviarPacket(montarSkillCoolTimePacket()); err != nil {
		return err
	}
	if err := g.sincronizarVisibilidadeAoEntrarNoMundo(); err != nil {
		return err
	}
	if err := g.enviarTrainerPessoal(); err != nil {
		return err
	}
	return g.enviarPacket(montarActionFailedPacket())
}

func (g *gameClient) carregarSlotSelecionado(indice int32) (*gsdb.CharacterSlot, error) {
	slots, err := g.server.characterRepo.FindByAccount(context.Background(), g.conta)
	if err != nil {
		return nil, err
	}
	if indice < 0 {
		return nil, nil
	}
	if int(indice) >= len(slots) {
		return nil, nil
	}
	slot := slots[indice]
	return &slot, nil
}

func (g *gameClient) garantirSpawnSeguro() {
	if g.personagemAtual == nil {
		return
	}
	template, ok := obterTemplatePersonagemInicial(g.personagemAtual.ClassID)
	if !ok {
		return
	}
	spawnInicial := template.obterSpawnInicial(g.personagemAtual.ObjID)
	xAjustado, yAjustado, zAjustado := normalizarPosicaoMundo(g.personagemAtual.X, g.personagemAtual.Y, g.personagemAtual.Z)
	statusInconsistente := g.personagemAtual.MaxHp <= 0 || g.personagemAtual.MaxMp <= 0 || g.personagemAtual.Level <= 0
	if statusInconsistente || !g.posicaoPareceSegura(xAjustado, yAjustado, zAjustado, template) {
		xAjustado = spawnInicial.x
		yAjustado = spawnInicial.y
		zAjustado = spawnInicial.z
	}
	g.personagemAtual.X = xAjustado
	g.personagemAtual.Y = yAjustado
	g.personagemAtual.Z = zAjustado
	err := g.server.characterRepo.AtualizarPosicao(context.Background(), g.personagemAtual.ObjID, xAjustado, yAjustado, zAjustado)
	if err != nil {
		logger.Warnf("Falha ao persistir spawn seguro para personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, err)
		return
	}
	logger.Infof("Spawn inicial validado para personagem %s objID=%d pos=(%d,%d,%d)", g.personagemAtual.CharName, g.personagemAtual.ObjID, xAjustado, yAjustado, zAjustado)
}

func (g *gameClient) posicaoPareceSegura(x int32, y int32, z int32, template templatePersonagemInicial) bool {
	if z <= -5000 {
		return false
	}
	if x == 0 && y == 0 {
		return false
	}
	if distancia3D(x, y, z, template.x, template.y, template.z) > 3000 {
		return false
	}
	return true
}

func (g *gameClient) sincronizarVisibilidadeAoEntrarNoMundo() error {
	if g.playerAtivo == nil {
		return nil
	}
	visiveis := g.server.mundo.listarVisiveisPara(g)
	for _, outroCliente := range visiveis {
		if outroCliente == nil {
			continue
		}
		if outroCliente.playerAtivo == nil {
			continue
		}
		if err := g.enviarPacket(montarCharInfoPacket(outroCliente.playerAtivo, outroCliente.itensAtivos, outroCliente.augmentacoesAtivas, outroCliente.playerAtivo.listarCubics())); err != nil {
			return err
		}
		err := outroCliente.enviarPacket(montarCharInfoPacket(g.playerAtivo, g.itensAtivos, g.augmentacoesAtivas, g.playerAtivo.listarCubics()))
		if err != nil {
			logger.Warnf("Falha ao enviar CharInfo de %s para %s: %v", g.playerAtivo.nome, outroCliente.playerAtivo.nome, err)
		}
	}
	return nil
}
