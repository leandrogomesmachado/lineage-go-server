package network

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

var geradorSpawnGlobal = rand.New(rand.NewSource(time.Now().UnixNano()))
var avisosTerritorioSpawnSuspeito = map[string]bool{}

type npcGlobalRuntime struct {
	objID                   int32
	npcID                   int32
	idTemplate              int32
	nome                    string
	titulo                  string
	alias                   string
	tipo                    string
	tipoAI                  string
	scriptAiDescritor       string
	scriptAiBase            string
	scriptAiVariante        string
	makerEvento             string
	makerMaximumNpcs        int32
	spawnDbName             string
	spawnDbSaving           string
	ehMonster               bool
	nivel                   int32
	hpAtual                 int32
	hpMaximo                int32
	mpAtual                 int32
	mpMaximo                int32
	pAtk                    int32
	pDef                    int32
	mAtk                    int32
	mDef                    int32
	crit                    int32
	aggroRange              int32
	origemX                 int32
	origemY                 int32
	origemZ                 int32
	alvoObjID               int32
	x                       int32
	y                       int32
	z                       int32
	heading                 int32
	regiaoX                 int32
	regiaoY                 int32
	ultimoMoveX             int32
	ultimoMoveY             int32
	ultimoMoveZ             int32
	radiusColisao           float64
	heightColisao           float64
	runSpd                  int32
	walkSpd                 int32
	pAtkSpd                 int32
	mAtkSpd                 int32
	rHand                   int32
	lHand                   int32
	canMove                 bool
	canBeAttacked           bool
	estaMorto               bool
	hatePorAlvo             map[int32]int32
	danoPorAlvo             map[int32]int64
	ultimoAggroMs           int64
	ultimoAtaqueMs          int64
	ultimoRegenMs           int64
	retornandoSpawn         bool
	respawnDelayMs          int64
	respawnAteMs            int64
	spawnTerritorio         territorioSpawn
	spawnPosFixa            string
	aiParams                npcAiParametros
	aiLifeTime              int32
	aiStep                  int32
	aiUltimoProcessamentoMs int64
	aiSeenPlayers           map[int32]int64
	aiTopDesireTargetObjID  int32
	aiHateList              npcHateList
	aiUltimoDesire          npcDesire
	aiProximoDesire         npcDesire
	aiDesireQueue           npcDesireQueue
}

func (n *npcGlobalRuntime) atualizarRegiao() {
	if n == nil {
		return
	}
	n.regiaoX = calcularRegiaoX(n.x)
	n.regiaoY = calcularRegiaoY(n.y)
}

func (n *npcGlobalRuntime) ehAgressivo() bool {
	if n == nil {
		return false
	}
	return n.aggroRange > 0
}

func construirNpcsGlobaisDoSpawn() []*npcGlobalRuntime {
	templates := obterTemplatesSpawnGlobal()
	resultado := make([]*npcGlobalRuntime, 0)
	proximoObjID := int32(700000000)
	totalMakersProcessados := 0
	totalMakersSemTerritorio := 0
	totalMakersFixosSemTerritorio := 0
	totalNpcsTemplatesAusentes := 0
	totalNpcsInstanciados := 0
	quantidadePorNpcID := map[int32]int32{}
	totalMakersEvento := 0
	for _, maker := range templates.makers {
		totalMakersProcessados++
		if maker.evento != "" {
			totalMakersEvento++
			if totalMakersEvento <= 5 {
				logger.Infof("Maker de evento ignorado no spawn automatico arquivo=%s maker=%s evento=%s (sera controlado por sistema de eventos)", maker.arquivoOrigem, maker.nome, maker.evento)
			}
			continue
		}
		territorio, ok := templates.territorios[maker.territorioNome]
		makerTemPosFixa := makerSpawnTemPosFixa(maker)
		if maker.territorioNome != "" && !ok {
			totalMakersSemTerritorio++
			logger.Warnf("Maker de spawn descartado por territorio ausente arquivo=%s maker=%s territorio=%s maximumNpcs=%d npcs=%d", maker.arquivoOrigem, maker.nome, maker.territorioNome, maker.maximumNpcs, len(maker.npcs))
			continue
		}
		if maker.territorioNome == "" && makerTemPosFixa {
			totalMakersFixosSemTerritorio++
			logger.Infof("Maker de spawn fixo processado sem territorio arquivo=%s maker=%s maximumNpcs=%d npcs=%d", maker.arquivoOrigem, maker.nome, maker.maximumNpcs, len(maker.npcs))
		}
		if maker.territorioNome == "" && !makerTemPosFixa {
			totalMakersSemTerritorio++
			logger.Warnf("Maker de spawn descartado sem territorio e sem pos fixa arquivo=%s maker=%s maximumNpcs=%d npcs=%d", maker.arquivoOrigem, maker.nome, maker.maximumNpcs, len(maker.npcs))
			continue
		}
		if strings.Contains(strings.ToLower(maker.nome), "gludio") || strings.Contains(strings.ToLower(maker.territorioNome), "gludio") {
			logger.Infof("Maker Gludin processado arquivo=%s maker=%s territorio=%s maximumNpcs=%d npcs=%d", maker.arquivoOrigem, maker.nome, maker.territorioNome, maker.maximumNpcs, len(maker.npcs))
		}
		for _, npcMaker := range maker.npcs {
			templateNpc, okNpc := obterTemplateNpc(npcMaker.npcID)
			if !okNpc {
				totalNpcsTemplatesAusentes++
				logger.Warnf("NPC de spawn descartado por template ausente arquivo=%s maker=%s territorio=%s npcID=%d total=%d dbName=%s dbSaving=%s", npcMaker.arquivoOrigem, maker.nome, maker.territorioNome, npcMaker.npcID, npcMaker.total, npcMaker.dbName, npcMaker.dbSaving)
				continue
			}
			scriptAiInfo := obterScriptAiMonsterPorNpcID(npcMaker.npcID)
			aiParamsMesclados := novoNpcAiParametros()
			aiParamsMesclados.mesclar(templateNpc.aiParams)
			aiParamsMesclados.mesclar(maker.aiParams)
			aiParamsMesclados.mesclar(npcMaker.aiParams)
			respawnDelayMs := resolverRespawnDelaySpawnGlobal(npcMaker)
			total := npcMaker.total
			if total <= 0 {
				total = 1
			}
			for indice := int32(0); indice < total; indice++ {
				x, y, z, heading, okPosicao := resolverPosicaoSpawnGlobal(territorio, npcMaker.pos, indice, templateNpc.ehMonster(), ok)
				if !okPosicao {
					logger.Warnf("NPC de spawn descartado por posicao invalida arquivo=%s maker=%s territorio=%s npcID=%d indice=%d pos=%s", npcMaker.arquivoOrigem, maker.nome, maker.territorioNome, npcMaker.npcID, indice, npcMaker.pos)
					continue
				}
				npc := &npcGlobalRuntime{objID: proximoObjID, npcID: npcMaker.npcID, idTemplate: templateNpc.idTemplate, nome: templateNpc.nome, titulo: templateNpc.titulo, alias: templateNpc.alias, tipo: templateNpc.tipo, tipoAI: maker.tipoAI, scriptAiDescritor: scriptAiInfo.descritor, scriptAiBase: scriptAiInfo.base, scriptAiVariante: scriptAiInfo.variante, makerEvento: maker.evento, makerMaximumNpcs: maker.maximumNpcs, spawnDbName: npcMaker.dbName, spawnDbSaving: npcMaker.dbSaving, ehMonster: templateNpc.ehMonster(), nivel: templateNpc.nivel, hpAtual: templateNpc.hp, hpMaximo: templateNpc.hp, mpAtual: templateNpc.mp, mpMaximo: templateNpc.mp, pAtk: templateNpc.pAtk, pDef: templateNpc.pDef, mAtk: templateNpc.mAtk, mDef: templateNpc.mDef, crit: templateNpc.crit, aggroRange: templateNpc.aggroRange, origemX: x, origemY: y, origemZ: z, x: x, y: y, z: z, heading: heading, ultimoMoveX: x, ultimoMoveY: y, ultimoMoveZ: z, radiusColisao: templateNpc.radius, heightColisao: templateNpc.height, runSpd: templateNpc.runSpd, walkSpd: templateNpc.walkSpd, pAtkSpd: templateNpc.pAtkSpd, mAtkSpd: templateNpc.mAtkSpd, rHand: templateNpc.rHand, lHand: templateNpc.lHand, canMove: templateNpc.canMove, canBeAttacked: templateNpc.canBeAttacked, hatePorAlvo: map[int32]int32{}, danoPorAlvo: map[int32]int64{}, respawnDelayMs: respawnDelayMs, spawnTerritorio: territorio, spawnPosFixa: npcMaker.pos, aiParams: aiParamsMesclados, aiSeenPlayers: map[int32]int64{}, aiHateList: novoNpcHateList(), aiDesireQueue: novoNpcDesireQueue()}
				npc.atualizarRegiao()
				resultado = append(resultado, npc)
				totalNpcsInstanciados++
				quantidadePorNpcID[npc.npcID]++
				proximoObjID++
			}
		}
	}
	logger.Infof("Resumo construcao spawn global: makersProcessados=%d makersEvento=%d makersSemTerritorio=%d makersFixosSemTerritorio=%d npcsTemplatesAusentes=%d npcsInstanciados=%d", totalMakersProcessados, totalMakersEvento, totalMakersSemTerritorio, totalMakersFixosSemTerritorio, totalNpcsTemplatesAusentes, totalNpcsInstanciados)
	logger.Infof("Resumo mobs alvo spawn global: direWolf20205=%d blackWolf20317=%d kashaDireWolf21121=%d grayWolf20525=%d whiteWolf20527=%d", quantidadePorNpcID[20205], quantidadePorNpcID[20317], quantidadePorNpcID[21121], quantidadePorNpcID[20525], quantidadePorNpcID[20527])
	return resultado
}

func makerSpawnTemPosFixa(maker makerSpawnGlobalTemplate) bool {
	for _, npcMaker := range maker.npcs {
		if strings.TrimSpace(npcMaker.pos) == "" {
			continue
		}
		return true
	}
	return false
}

func resolverRespawnDelaySpawnGlobal(template npcSpawnGlobalTemplate) int64 {
	baseSegundos := parseRespawnTextoSegundos(template.respawn)
	randomSegundos := parseRespawnTextoSegundos(template.respawnRand)
	if randomSegundos > 0 {
		variacao := geradorSpawnGlobal.Int63n((randomSegundos * 2) + 1)
		baseSegundos += variacao - randomSegundos
	}
	if baseSegundos < 0 {
		baseSegundos = 0
	}
	return baseSegundos * 1000
}

func parseRespawnTextoSegundos(valor string) int64 {
	valorLimpo := strings.TrimSpace(strings.ToLower(valor))
	if valorLimpo == "" {
		return 0
	}
	if strings.HasSuffix(valorLimpo, "sec") {
		valorLimpo = strings.TrimSpace(strings.TrimSuffix(valorLimpo, "sec"))
	}
	if strings.HasSuffix(valorLimpo, "s") {
		valorLimpo = strings.TrimSpace(strings.TrimSuffix(valorLimpo, "s"))
	}
	if strings.HasSuffix(valorLimpo, "min") {
		minutos := int64(parseInt32Seguro(strings.TrimSpace(strings.TrimSuffix(valorLimpo, "min"))))
		if minutos < 0 {
			return 0
		}
		return minutos * 60
	}
	if strings.HasSuffix(valorLimpo, "m") {
		minutos := int64(parseInt32Seguro(strings.TrimSpace(strings.TrimSuffix(valorLimpo, "m"))))
		if minutos < 0 {
			return 0
		}
		return minutos * 60
	}
	segundos := int64(parseInt32Seguro(valorLimpo))
	if segundos < 0 {
		return 0
	}
	return segundos
}

func resolverPosicaoSpawnGlobal(territorio territorioSpawn, pos string, indice int32, ehMonster bool, possuiTerritorio bool) (int32, int32, int32, int32, bool) {
	if possuiTerritorio {
		logarTerritorioSpawnSuspeitoSeNecessario(territorio)
	}
	if pos != "" {
		partes := strings.Split(pos, ";")
		if len(partes) >= 4 {
			x := parseInt32Seguro(partes[0])
			y := parseInt32Seguro(partes[1])
			z := parseInt32Seguro(partes[2])
			heading := parseInt32Seguro(partes[3])
			xNormalizado, yNormalizado, zNormalizado, headingNormalizado := normalizarPosicaoGlobal(x, y, z, heading)
			return xNormalizado, yNormalizado, zNormalizado, headingNormalizado, true
		}
	}
	if !possuiTerritorio {
		return 0, 0, 0, 0, false
	}
	if len(territorio.nos) < 3 {
		return 0, 0, 0, 0, false
	}
	x, y, z, heading := sortearPosicaoTerritorio(territorio)
	xNormalizado, yNormalizado, zNormalizado, headingNormalizado := normalizarPosicaoGlobal(x, y, z, heading)
	return xNormalizado, yNormalizado, zNormalizado, headingNormalizado, true
}

func sortearPosicaoTerritorio(territorio territorioSpawn) (int32, int32, int32, int32) {
	avgZ := resolverZReferenciaTerritorio(territorio)
	var ultimoX, ultimoY, ultimoZ int32
	ultimoX = territorio.nos[0].x
	ultimoY = territorio.nos[0].y
	ultimoZ = resolverZComGeodataOuTerritorio(ultimoX, ultimoY, avgZ, territorio)
	if territorio.maxX <= territorio.minX || territorio.maxY <= territorio.minY {
		heading := int32(geradorSpawnGlobal.Intn(65536))
		return ultimoX, ultimoY, ultimoZ, heading
	}
	for tentativa := 0; tentativa < 10; tentativa++ {
		x := territorio.minX + int32(geradorSpawnGlobal.Intn(int(territorio.maxX-territorio.minX+1)))
		y := territorio.minY + int32(geradorSpawnGlobal.Intn(int(territorio.maxY-territorio.minY+1)))
		if !pontoDentroTerritorio(territorio, x, y) {
			continue
		}
		z := resolverZComGeodataOuTerritorio(x, y, avgZ, territorio)
		ultimoX = x
		ultimoY = y
		ultimoZ = z
		heading := int32(geradorSpawnGlobal.Intn(65536))
		return x, y, z, heading
	}
	heading := int32(geradorSpawnGlobal.Intn(65536))
	return ultimoX, ultimoY, ultimoZ, heading
}

func resolverZSeguroSpawnGlobal(territorio territorioSpawn, zBase int32) int32 {
	if territorio.maxZ == 0 && territorio.minZ == 0 {
		return zBase
	}
	if territorio.maxZ < territorio.minZ {
		return zBase
	}
	if zBase < territorio.minZ {
		return territorio.maxZ
	}
	if zBase > territorio.maxZ {
		return territorio.maxZ
	}
	faixaZ := territorio.maxZ - territorio.minZ
	if faixaZ <= 512 {
		return zBase
	}
	zLimiteInferior := territorio.maxZ - 512
	if zBase < zLimiteInferior {
		return territorio.maxZ
	}
	return zBase
}

func logarTerritorioSpawnSuspeitoSeNecessario(territorio territorioSpawn) {
	if territorio.nome == "" {
		return
	}
	if avisosTerritorioSpawnSuspeito[territorio.nome] {
		return
	}
	faixaZ := territorio.maxZ - territorio.minZ
	if faixaZ <= 1500 {
		return
	}
	avisosTerritorioSpawnSuspeito[territorio.nome] = true
	logger.Warnf("Territorio de spawn com faixa vertical alta nome=%s minZ=%d maxZ=%d faixaZ=%d minX=%d maxX=%d minY=%d maxY=%d", territorio.nome, territorio.minZ, territorio.maxZ, faixaZ, territorio.minX, territorio.maxX, territorio.minY, territorio.maxY)
}

func pontoDentroTerritorio(territorio territorioSpawn, x int32, y int32) bool {
	quantidade := len(territorio.nos)
	if quantidade < 3 {
		return false
	}
	dentro := false
	for atual, anterior := 0, quantidade-1; atual < quantidade; anterior, atual = atual, atual+1 {
		xAtual := territorio.nos[atual].x
		yAtual := territorio.nos[atual].y
		xAnterior := territorio.nos[anterior].x
		yAnterior := territorio.nos[anterior].y
		intersecta := ((yAtual > y) != (yAnterior > y)) && (float64(x) < (float64(xAnterior-xAtual)*float64(y-yAtual))/float64(yAnterior-yAtual)+float64(xAtual))
		if !intersecta {
			continue
		}
		dentro = !dentro
	}
	return dentro
}

func normalizarPosicaoGlobal(x int32, y int32, z int32, heading int32) (int32, int32, int32, int32) {
	xAjustado, yAjustado, zAjustado := normalizarPosicaoMundo(x, y, z)
	return xAjustado, yAjustado, zAjustado, heading
}

func (g *gameServer) inicializarNpcsGlobais() {
	if g == nil {
		return
	}
	if g.mundo == nil {
		return
	}
	g.mundo.limparNpcs()
	npcs := construirNpcsGlobaisDoSpawn()
	amostraLogada := 0
	totalComScriptAi := 0
	totalWarriorBase := 0
	totalGremlin := 0
	totalGremlinWarriorBase := 0
	gremlinLogado := false
	for _, npc := range npcs {
		g.mundo.registrarNpc(npc)
		if strings.TrimSpace(npc.scriptAiDescritor) != "" {
			totalComScriptAi++
		}
		if strings.EqualFold(strings.TrimSpace(npc.scriptAiBase), "WarriorBase") {
			totalWarriorBase++
		}
		if npc.npcID == 18342 {
			totalGremlin++
			if npc.ehScriptWarriorBase() {
				totalGremlinWarriorBase++
			}
			if !gremlinLogado {
				logger.Infof("Gremlin carregado npcID=%d scriptAI=%s scriptBase=%s scriptVariante=%s tipoAI=%s makerEvento=%s dbName=%s dbSaving=%s SetAggressiveTime=%d HalfAggressive=%d RandomAggressive=%d AttackLowLevel=%d IsVs=%d", npc.npcID, npc.scriptAiDescritor, npc.scriptAiBase, npc.scriptAiVariante, npc.tipoAI, npc.makerEvento, npc.spawnDbName, npc.spawnDbSaving, npc.obterNpcIntAiParamOuPadrao("SetAggressiveTime", 0), npc.obterNpcIntAiParamOuPadrao("HalfAggressive", 0), npc.obterNpcIntAiParamOuPadrao("RandomAggressive", 0), npc.obterNpcIntAiParamOuPadrao("AttackLowLevel", 0), npc.obterNpcIntAiParamOuPadrao("IsVs", 0))
				gremlinLogado = true
			}
		}
		if amostraLogada >= 5 {
			continue
		}
		logger.Infof("NPC global instanciado objID=%d npcID=%d idTemplate=%d nome=%s tipoAI=%s scriptAI=%s scriptBase=%s scriptVariante=%s makerEvento=%s dbName=%s dbSaving=%s pos=(%d,%d,%d) heading=%d", npc.objID, npc.npcID, npc.idTemplate, npc.nome, npc.tipoAI, npc.scriptAiDescritor, npc.scriptAiBase, npc.scriptAiVariante, npc.makerEvento, npc.spawnDbName, npc.spawnDbSaving, npc.x, npc.y, npc.z, npc.heading)
		amostraLogada++
	}
	logger.Infof("NPCs globais com script AI autoritativo: total=%d warriorBase=%d", totalComScriptAi, totalWarriorBase)
	logger.Infof("Gremlins carregados no spawn global: total=%d warriorBase=%d", totalGremlin, totalGremlinWarriorBase)
	logger.Infof("NPCs globais inicializados: %d", len(npcs))
}

func nomeNpcGlobalLog(npc *npcGlobalRuntime) string {
	if npc == nil {
		return "npc_global_nil"
	}
	return fmt.Sprintf("%s(%d)", npc.nome, npc.npcID)
}

func (n *npcGlobalRuntime) aplicarPosicao(x int32, y int32, z int32) {
	if n == nil {
		return
	}
	n.x = x
	n.y = y
	n.z = z
	n.atualizarRegiao()
}

func (n *npcGlobalRuntime) aplicarPosicaoComHeading(x int32, y int32, z int32, heading int32) {
	if n == nil {
		return
	}
	n.x = x
	n.y = y
	n.z = z
	n.heading = heading
	n.atualizarRegiao()
}
