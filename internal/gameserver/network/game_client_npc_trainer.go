package network

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
)

func (g *gameClient) inicializarTrainerPessoal() {
	if g.playerAtivo == nil {
		return
	}
	g.trainerPessoal = novoTrainerPessoal(g.playerAtivo)
}

func (g *gameClient) enviarTrainerPessoal() error {
	if g.trainerPessoal == nil {
		return nil
	}
	return g.enviarPacket(montarNpcInfoPacket(g.trainerPessoal))
}

func (g *gameClient) processarRequestBypassToServer(packet *requestBypassToServerPacket) error {
	if g.playerAtivo == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	comando := strings.TrimSpace(packet.comando)
	if strings.HasPrefix(comando, "npc_") {
		return g.processarBypassNpcGlobal(comando)
	}
	return g.enviarPacket(montarActionFailedPacket())
}

func (g *gameClient) processarBypassNpcGlobal(comando string) error {
	partes := strings.SplitN(strings.TrimSpace(comando), "_", 3)
	if len(partes) < 3 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	objID64, err := strconv.ParseInt(strings.TrimSpace(partes[1]), 10, 32)
	if err != nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	objID := int32(objID64)
	npcGlobal := g.server.mundo.obterNpcPorObjID(objID)
	if npcGlobal == nil {
		if g.trainerPessoal == nil {
			return g.enviarPacket(montarActionFailedPacket())
		}
		if g.trainerPessoal.objID != objID {
			return g.enviarPacket(montarActionFailedPacket())
		}
		acaoTrainer := strings.TrimSpace(partes[2])
		if acaoTrainer == "SkillList" {
			return g.enviarSkillListTrainer()
		}
		if acaoTrainer == "Quest" {
			return g.enviarPacket(montarNpcHtmlMessagePacket(g.trainerPessoal.objID, "<html><body>Quest indisponivel.</body></html>"))
		}
		return g.enviarPacket(montarActionFailedPacket())
	}
	if npcGlobal.ehMonster {
		return g.enviarPacket(montarActionFailedPacket())
	}
	acao := strings.TrimSpace(partes[2])
	if acao == "Quest" {
		html := montarHtmlQuestNpcGlobal(npcGlobal)
		return g.enviarPacket(montarNpcHtmlMessagePacket(npcGlobal.objID, html))
	}
	return g.enviarHtmlNpcGlobal(npcGlobal)
}

func (g *gameClient) enviarHtmlNpcGlobal(npc *npcGlobalRuntime) error {
	if npc == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	html := g.carregarHtmlNpcGlobal(npc)
	return g.enviarPacket(montarNpcHtmlMessagePacket(npc.objID, html))
}

func (g *gameClient) carregarHtmlNpcGlobal(npc *npcGlobalRuntime) string {
	if npc == nil {
		return "<html><body>NPC indisponivel.</body></html>"
	}
	if g == nil || g.server == nil || g.server.config == nil {
		return "<html><body>NPC indisponivel.</body></html>"
	}
	caminhos := resolverCaminhosHtmlNpc(g.server.config.Datapack.Path, npc, 0)
	for _, caminho := range caminhos {
		dados, err := os.ReadFile(caminho)
		if err != nil {
			continue
		}
		html := string(dados)
		html = strings.ReplaceAll(html, "%objectId%", strconv.FormatInt(int64(npc.objID), 10))
		if strings.TrimSpace(html) == "" {
			continue
		}
		return html
	}
	return "<html><body>NPC indisponivel.</body></html>"
}

func resolverCaminhosHtmlNpc(datapackPath string, npc *npcGlobalRuntime, pagina int32) []string {
	if npc == nil {
		return []string{}
	}
	npcIDTexto := strconv.FormatInt(int64(npc.npcID), 10)
	nomeArquivo := npcIDTexto + ".htm"
	if pagina > 0 {
		nomeArquivo = npcIDTexto + "-" + strconv.FormatInt(int64(pagina), 10) + ".htm"
	}
	baseHtml := filepath.Join(datapackPath, "data", "html")
	tipoNormalizado := normalizarTipoHtmlNpc(npc.tipo)
	candidatos := make([]string, 0, 8)
	if tipoNormalizado != "" {
		candidatos = append(candidatos, filepath.Join(baseHtml, tipoNormalizado, nomeArquivo))
	}
	candidatos = append(candidatos, filepath.Join(baseHtml, "default", nomeArquivo))
	if pagina == 0 {
		candidatos = append(candidatos, filepath.Join(baseHtml, "npcdefault.htm"))
	}
	return removerDuplicadosCaminho(candidatos)
}

func normalizarTipoHtmlNpc(tipo string) string {
	tipoNormalizado := strings.ToLower(strings.TrimSpace(tipo))
	if tipoNormalizado == "" {
		return ""
	}
	if tipoNormalizado == "folk" {
		return "default"
	}
	if strings.Contains(tipoNormalizado, "village") && strings.Contains(tipoNormalizado, "master") {
		return "villagemaster"
	}
	if strings.Contains(tipoNormalizado, "merchant") {
		return "merchant"
	}
	if strings.Contains(tipoNormalizado, "teleporter") {
		return "teleporter"
	}
	if strings.Contains(tipoNormalizado, "warehouse") {
		return "warehouse"
	}
	if strings.Contains(tipoNormalizado, "trainer") {
		return "trainer"
	}
	return tipoNormalizado
}

func removerDuplicadosCaminho(caminhos []string) []string {
	if len(caminhos) == 0 {
		return []string{}
	}
	mapa := make(map[string]struct{}, len(caminhos))
	resultado := make([]string, 0, len(caminhos))
	for _, caminho := range caminhos {
		caminhoLimpo := strings.TrimSpace(caminho)
		if caminhoLimpo == "" {
			continue
		}
		if _, existe := mapa[caminhoLimpo]; existe {
			continue
		}
		mapa[caminhoLimpo] = struct{}{}
		resultado = append(resultado, caminhoLimpo)
	}
	return resultado
}

func montarHtmlQuestNpcGlobal(npc *npcGlobalRuntime) string {
	if npc == nil {
		return "<html><body>Quest indisponivel.</body></html>"
	}
	return "<html><body>Quest indisponivel para " + npc.nome + ".</body></html>"
}

func (g *gameClient) enviarHtmlTrainer() error {
	if g.trainerPessoal == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.personagemAtual == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if !g.trainerPessoal.podeEnsinar(g.personagemAtual.ClassID) {
		html := g.trainerPessoal.carregarHtmlSemSkills(g.server.config.Datapack.Path)
		return g.enviarPacket(montarNpcHtmlMessagePacket(g.trainerPessoal.objID, html))
	}
	html := g.trainerPessoal.carregarHtmlBase(g.server.config.Datapack.Path)
	return g.enviarPacket(montarNpcHtmlMessagePacket(g.trainerPessoal.objID, html))
}

func (g *gameClient) enviarSkillListTrainer() error {
	if g.personagemAtual == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.trainerPessoal == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if !g.trainerPessoal.podeEnsinar(g.personagemAtual.ClassID) {
		html := g.trainerPessoal.carregarHtmlSemSkills(g.server.config.Datapack.Path)
		if err := g.enviarPacket(montarNpcHtmlMessagePacket(g.trainerPessoal.objID, html)); err != nil {
			return err
		}
		return g.enviarPacket(montarActionFailedPacket())
	}
	pacote := montarAcquireSkillListPacket(g.skillsAtivas, g.personagemAtual.ClassID, g.personagemAtual.Level)
	if err := g.enviarPacket(pacote); err != nil {
		return err
	}
	return g.enviarPacket(montarActionFailedPacket())
}

func listarProximasSkillsAprendiveis(skillsAtuais []gsdb.CharacterSkill, classID int32, nivel int32) []templateSkillClasse {
	templatesClasseSkillsMu.RLock()
	template, ok := templatesClasseSkills[classID]
	templatesClasseSkillsMu.RUnlock()
	if !ok {
		return []templateSkillClasse{}
	}
	mapaAtual := make(map[int32]int32, len(skillsAtuais))
	for _, skillAtual := range skillsAtuais {
		nivelAtual, existe := mapaAtual[skillAtual.SkillID]
		if existe && nivelAtual >= skillAtual.SkillLevel {
			continue
		}
		mapaAtual[skillAtual.SkillID] = skillAtual.SkillLevel
	}
	resultado := make([]templateSkillClasse, 0)
	for _, skill := range template.skills {
		if skill.minLvl > nivel {
			continue
		}
		nivelAtual := mapaAtual[skill.skillID]
		if nivelAtual >= skill.skillLevel {
			continue
		}
		if nivelAtual != skill.skillLevel-1 {
			continue
		}
		resultado = append(resultado, skill)
	}
	sort.Slice(resultado, func(i int, j int) bool {
		if resultado[i].minLvl != resultado[j].minLvl {
			return resultado[i].minLvl < resultado[j].minLvl
		}
		if resultado[i].skillID != resultado[j].skillID {
			return resultado[i].skillID < resultado[j].skillID
		}
		return resultado[i].skillLevel < resultado[j].skillLevel
	})
	return resultado
}

func (g *gameClient) processarRequestAcquireSkillInfo(packet *requestAcquireSkillInfoPacket) error {
	if packet.skillType != 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.personagemAtual == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.trainerPessoal == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if !g.trainerPessoal.podeEnsinar(g.personagemAtual.ClassID) {
		return g.enviarPacket(montarActionFailedPacket())
	}
	skill, ok := g.obterSkillAprendivel(packet.skillID, packet.skillLevel)
	if !ok {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if err := g.enviarPacket(montarAcquireSkillInfoPacket(skill)); err != nil {
		return err
	}
	return g.enviarPacket(montarActionFailedPacket())
}

func (g *gameClient) processarRequestAcquireSkill(packet *requestAcquireSkillPacket) error {
	if packet.skillType != 0 {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.trainerPessoal == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.personagemAtual == nil {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if !g.trainerPessoal.podeEnsinar(g.personagemAtual.ClassID) {
		return g.enviarPacket(montarActionFailedPacket())
	}
	skill, ok := g.obterSkillAprendivel(packet.skillID, packet.skillLevel)
	if !ok {
		return g.enviarPacket(montarActionFailedPacket())
	}
	if g.personagemAtual.Sp < skill.cost {
		return g.enviarSkillListTrainer()
	}
	novaSkill := gsdb.CharacterSkill{CharObjID: g.personagemAtual.ObjID, SkillID: skill.skillID, SkillLevel: skill.skillLevel, ClassIndex: 0}
	if err := g.server.repositorios.CharacterSkills.InserirLote(context.Background(), []gsdb.CharacterSkill{novaSkill}); err != nil {
		return err
	}
	g.skillsAtivas = append(g.skillsAtivas, novaSkill)
	ordenarSkillsAtivas(g.skillsAtivas)
	g.personagemAtual.Sp -= skill.cost
	if g.playerAtivo != nil {
		g.playerAtivo.sp = g.personagemAtual.Sp
	}
	if err := g.enviarUserInfoAtualizado(); err != nil {
		return err
	}
	g.broadcastCharInfoAtualizado()
	if err := g.enviarPacket(montarSkillListPacket(g.skillsAtivas)); err != nil {
		return err
	}
	return g.enviarSkillListTrainer()
}

func (g *gameClient) obterSkillAprendivel(skillID int32, skillLevel int32) (templateSkillClasse, bool) {
	if g.personagemAtual == nil {
		return templateSkillClasse{}, false
	}
	disponiveis := listarProximasSkillsAprendiveis(g.skillsAtivas, g.personagemAtual.ClassID, g.personagemAtual.Level)
	for _, skill := range disponiveis {
		if skill.skillID != skillID {
			continue
		}
		if skill.skillLevel != skillLevel {
			continue
		}
		return skill, true
	}
	return templateSkillClasse{}, false
}
