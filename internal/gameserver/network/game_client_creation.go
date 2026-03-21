package network

import (
	"context"
	"regexp"
	"strings"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

var regexNomePersonagem = regexp.MustCompile(`^[A-Za-z0-9]{1,16}$`)

func (g *gameClient) processarRequestNewCharacter(packet *requestNewCharacterPacket) error {
	_ = packet
	logger.Infof("RequestNewCharacter recebido para conta %s", g.conta)
	return g.enviarPacket(montarNewCharacterSuccessPacket())
}

func (g *gameClient) processarRequestCharacterCreate(packet *requestCharacterCreatePacket) error {
	logger.Infof("RequestCharacterCreate recebido para conta %s com nome=%s classID=%d race=%d sexo=%d", g.conta, packet.nome, packet.classID, packet.race, packet.sexo)
	motivoFalha := g.validarCriacaoPersonagem(packet)
	if motivoFalha != 0xFFFFFFFF {
		return g.enviarPacket(montarCharCreateFailPacket(motivoFalha))
	}
	template, ok := obterTemplatePersonagemInicial(packet.classID)
	if !ok {
		return g.enviarPacket(montarCharCreateFailPacket(motivoCriacaoFalhou))
	}
	err := g.criarPersonagem(packet, template)
	if err != nil {
		logger.Errorf("Erro ao criar personagem da conta %s: %v", g.conta, err)
		return g.enviarPacket(montarCharCreateFailPacket(motivoCriacaoFalhou))
	}
	slots, err := g.server.characterRepo.FindByAccount(context.Background(), g.conta)
	if err != nil {
		logger.Errorf("Erro ao recarregar slots apos criar personagem da conta %s: %v", g.conta, err)
		return g.enviarPacket(montarCharCreateFailPacket(motivoCriacaoFalhou))
	}
	if err = g.enviarPacket(montarCharCreateOkPacket()); err != nil {
		return err
	}
	return g.enviarPacket(montarCharSelectInfoPacket(g.conta, g.sessionKey.PlayOkID1, slots))
}

func (g *gameClient) validarCriacaoPersonagem(packet *requestCharacterCreatePacket) uint32 {
	if packet.race < 0 || packet.race > 4 {
		return motivoCriacaoFalhou
	}
	if packet.face < 0 || packet.face > 2 {
		return motivoCriacaoFalhou
	}
	if packet.hairColor < 0 || packet.hairColor > 3 {
		return motivoCriacaoFalhou
	}
	if packet.sexo < 0 || packet.sexo > 1 {
		return motivoCriacaoFalhou
	}
	if packet.sexo == 0 && (packet.hairStyle < 0 || packet.hairStyle > 4) {
		return motivoCriacaoFalhou
	}
	if packet.sexo == 1 && (packet.hairStyle < 0 || packet.hairStyle > 6) {
		return motivoCriacaoFalhou
	}
	if !regexNomePersonagem.MatchString(packet.nome) {
		return motivoNomeIncorreto
	}
	template, ok := obterTemplatePersonagemInicial(packet.classID)
	if !ok {
		return motivoCriacaoFalhou
	}
	if template.race != packet.race {
		return motivoCriacaoFalhou
	}
	quantidade, err := g.server.characterRepo.CountByAccount(context.Background(), g.conta)
	if err != nil {
		logger.Errorf("Erro ao contar personagens da conta %s: %v", g.conta, err)
		return motivoCriacaoFalhou
	}
	if quantidade >= 7 {
		return motivoMuitosPersonagens
	}
	existe, err := g.server.characterRepo.ExistsByName(context.Background(), packet.nome)
	if err != nil {
		logger.Errorf("Erro ao verificar nome de personagem %s: %v", packet.nome, err)
		return motivoCriacaoFalhou
	}
	if existe {
		return motivoNomeJaExiste
	}
	return 0xFFFFFFFF
}

func (g *gameClient) criarPersonagem(packet *requestCharacterCreatePacket, template templatePersonagemInicial) error {
	spawnInicial := template.obterSpawnInicial(int32(len(strings.TrimSpace(packet.nome))))
	statsCalculadas := calcularStatsPersonagem(template, 1, []itemPapelBoneca{})
	cpInicial := int32(0)
	if statsCalculadas.cpMaximo > 0 {
		cpInicial = statsCalculadas.cpMaximo / 2
	}
	entrada := gsdb.CharacterCreateInput{
		AccountName: g.conta,
		CharName:    strings.TrimSpace(packet.nome),
		Race:        packet.race,
		Sex:         packet.sexo,
		ClassID:     packet.classID,
		BaseClass:   packet.classID,
		HairStyle:   packet.hairStyle,
		HairColor:   packet.hairColor,
		Face:        packet.face,
		X:           spawnInicial.x,
		Y:           spawnInicial.y,
		Z:           spawnInicial.z,
		Level:       1,
		MaxHp:       statsCalculadas.hpMaximo,
		CurHp:       statsCalculadas.hpMaximo,
		MaxMp:       statsCalculadas.mpMaximo,
		CurMp:       statsCalculadas.mpMaximo,
		MaxCp:       statsCalculadas.cpMaximo,
		CurCp:       cpInicial,
		Exp:         0,
		Sp:          0,
		Title:       "",
		NameColor:   0xFFFFFF,
		TitleColor:  0xFFFF77,
		AccessLevel: 0,
	}
	slotCriado, err := g.server.characterRepo.Create(context.Background(), entrada)
	if err != nil {
		return err
	}
	if g.server.repositorios == nil {
		return nil
	}
	if g.server.repositorios.CharacterSkills == nil {
		return nil
	}
	skillsIniciais := listarSkillsIniciaisClasse(packet.classID, 1)
	for i := range skillsIniciais {
		skillsIniciais[i].CharObjID = slotCriado.ObjID
	}
	if err := g.server.repositorios.CharacterSkills.InserirLote(context.Background(), skillsIniciais); err != nil {
		return err
	}
	if err := g.inserirItensEAtalhosIniciais(slotCriado.ObjID, template); err != nil {
		return err
	}
	return nil
}

func (g *gameClient) inserirItensEAtalhosIniciais(charObjID int32, template templatePersonagemInicial) error {
	ctx := context.Background()
	tutorialGuideObjID := int32(0)
	for _, itemTemplate := range template.itensIniciais {
		slots := resolverSlotsEquipamento(itemTemplate.itemID)
		equipavel := itemTemplate.estaEquipado && len(slots) > 0
		if !equipavel {
			item, err := g.server.repositorios.CharacterItems.InserirOuSomarItem(ctx, charObjID, itemTemplate.itemID, itemTemplate.count)
			if err != nil {
				return err
			}
			if item != nil && itemTemplate.itemID == 5588 {
				tutorialGuideObjID = item.ObjectID
			}
			continue
		}
		slotPrincipal := slots[0]
		entrada := gsdb.CharacterItem{
			OwnerID:  charObjID,
			ItemID:   itemTemplate.itemID,
			Count:    itemTemplate.count,
			Loc:      "PAPERDOLL",
			LocData:  slotPrincipal,
			ManaLeft: -1,
		}
		item, err := g.server.repositorios.CharacterItems.InserirItemCustom(ctx, entrada)
		if err != nil {
			return err
		}
		if item != nil && itemTemplate.itemID == 5588 {
			tutorialGuideObjID = item.ObjectID
		}
		for _, slotExtra := range slots[1:] {
			entradaExtra := gsdb.CharacterItem{
				OwnerID:  charObjID,
				ItemID:   itemTemplate.itemID,
				Count:    itemTemplate.count,
				Loc:      "PAPERDOLL",
				LocData:  slotExtra,
				ManaLeft: -1,
			}
			_, errExtra := g.server.repositorios.CharacterItems.InserirItemCustom(ctx, entradaExtra)
			if errExtra != nil {
				return errExtra
			}
		}
	}
	atalhos := []gsdb.CharacterShortcut{
		{CharObjID: charObjID, Slot: 0, Page: 0, Type: "ACTION", ID: 2, Level: 0, ClassIndex: 0},
		{CharObjID: charObjID, Slot: 3, Page: 0, Type: "ACTION", ID: 5, Level: 0, ClassIndex: 0},
		{CharObjID: charObjID, Slot: 10, Page: 0, Type: "ACTION", ID: 0, Level: 0, ClassIndex: 0},
	}
	if tutorialGuideObjID > 0 {
		atalhos = append(atalhos, gsdb.CharacterShortcut{CharObjID: charObjID, Slot: 11, Page: 0, Type: "ITEM", ID: tutorialGuideObjID, Level: 0, ClassIndex: 0})
	}
	return g.server.repositorios.CharacterShortcuts.InserirLote(ctx, atalhos)
}
