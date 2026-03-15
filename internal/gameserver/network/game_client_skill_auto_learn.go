package network

import (
	"context"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func (g *gameClient) sincronizarSkillsAutoLearn() {
	if g == nil {
		return
	}
	if g.server == nil {
		return
	}
	if g.server.config == nil {
		return
	}
	if !g.server.config.Skills.AutoLearn {
		return
	}
	if g.personagemAtual == nil {
		return
	}
	if g.server.repositorios == nil {
		return
	}
	if g.server.repositorios.CharacterSkills == nil {
		return
	}
	classIndex := int32(0)
	if g.personagemAtual.BaseClass > 0 && g.personagemAtual.BaseClass != g.personagemAtual.ClassID {
		classIndex = 1
	}
	skillsFaltantes := listarSkillsFaltantesClasse(g.personagemAtual.ClassID, g.personagemAtual.Level, g.skillsAtivas, classIndex, g.personagemAtual.ObjID)
	if len(skillsFaltantes) == 0 {
		return
	}
	ctx := context.Background()
	err := g.server.repositorios.CharacterSkills.InserirLote(ctx, skillsFaltantes)
	if err != nil {
		logger.Warnf("Falha ao aplicar auto learn de skills para personagem %s objID=%d: %v", g.personagemAtual.CharName, g.personagemAtual.ObjID, err)
		return
	}
	g.skillsAtivas = append(g.skillsAtivas, skillsFaltantes...)
	g.skillsAtivas = consolidarSkillsParaSkillList(g.skillsAtivas)
	logger.Infof("Auto learn aplicou %d skills para personagem %s objID=%d", len(skillsFaltantes), g.personagemAtual.CharName, g.personagemAtual.ObjID)
}

func ordenarSkillsAtivas(skills []gsdb.CharacterSkill) {
	if len(skills) <= 1 {
		return
	}
	for i := 0; i < len(skills)-1; i++ {
		for j := i + 1; j < len(skills); j++ {
			if skills[i].SkillID > skills[j].SkillID {
				skills[i], skills[j] = skills[j], skills[i]
				continue
			}
			if skills[i].SkillID < skills[j].SkillID {
				continue
			}
			if skills[i].SkillLevel <= skills[j].SkillLevel {
				continue
			}
			skills[i], skills[j] = skills[j], skills[i]
		}
	}
}
