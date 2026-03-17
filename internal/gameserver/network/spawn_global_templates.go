package network

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type territorioSpawn struct {
	nome string
	minZ int32
	maxZ int32
	minX int32
	maxX int32
	minY int32
	maxY int32
	nos  []pontoTerritorio
}

type pontoTerritorio struct {
	x int32
	y int32
}

type npcSpawnGlobalTemplate struct {
	npcID         int32
	total         int32
	respawn       string
	respawnRand   string
	pos           string
	dbName        string
	dbSaving      string
	arquivoOrigem string
	aiParams      npcAiParametros
}

type makerSpawnGlobalTemplate struct {
	nome           string
	territorioNome string
	tipoAI         string
	evento         string
	maximumNpcs    int32
	arquivoOrigem  string
	aiParams       npcAiParametros
	npcs           []npcSpawnGlobalTemplate
}

type spawnGlobalTemplate struct {
	territorios map[string]territorioSpawn
	makers      []makerSpawnGlobalTemplate
}

type xmlListaSpawnGlobal struct {
	Territorios []xmlTerritorioSpawn `xml:"territory"`
	NpcMakers   []xmlNpcMakerSpawn   `xml:"npcmaker"`
}

type xmlTerritorioSpawn struct {
	Name string                 `xml:"name,attr"`
	MinZ string                 `xml:"minZ,attr"`
	MaxZ string                 `xml:"maxZ,attr"`
	Nos  []xmlNoTerritorioSpawn `xml:"node"`
}

type xmlNoTerritorioSpawn struct {
	X string `xml:"x,attr"`
	Y string `xml:"y,attr"`
}

type xmlNpcMakerSpawn struct {
	Name        string              `xml:"name,attr"`
	Territory   string              `xml:"territory,attr"`
	Event       string              `xml:"event,attr"`
	MaximumNpcs string              `xml:"maximumNpcs,attr"`
	AIs         []xmlAiMakerSpawn   `xml:"ai"`
	Npcs        []xmlNpcSpawnGlobal `xml:"npc"`
}

type xmlAiMakerSpawn struct {
	Type string      `xml:"type,attr"`
	Sets []xmlNpcSet `xml:"set"`
}

type xmlNpcSpawnGlobal struct {
	ID          string      `xml:"id,attr"`
	Total       string      `xml:"total,attr"`
	Respawn     string      `xml:"respawn,attr"`
	RespawnRand string      `xml:"respawnRand,attr"`
	Pos         string      `xml:"pos,attr"`
	DbName      string      `xml:"dbName,attr"`
	DbSaving    string      `xml:"dbSaving,attr"`
	AiSets      []xmlNpcSet `xml:"ai>set"`
}

var spawnGlobalAtual = spawnGlobalTemplate{territorios: map[string]territorioSpawn{}, makers: []makerSpawnGlobalTemplate{}}
var spawnGlobalMu sync.RWMutex

func carregarTemplatesSpawnGlobal(datapackPath string) error {
	spawnPath := filepath.Join(datapackPath, "data", "xml", "spawnlist")
	arquivos, err := filepath.Glob(filepath.Join(spawnPath, "*.xml"))
	if err != nil {
		return err
	}
	template := spawnGlobalTemplate{territorios: make(map[string]territorioSpawn), makers: make([]makerSpawnGlobalTemplate, 0)}
	totalTerritorios := 0
	totalMakersLidos := 0
	totalMakersValidos := 0
	totalNpcsLidos := 0
	totalNpcsValidos := 0
	for _, arquivo := range arquivos {
		dados, errLeitura := os.ReadFile(arquivo)
		if errLeitura != nil {
			logger.Warnf("Falha ao ler XML de spawn %s: %v", arquivo, errLeitura)
			continue
		}
		var lista xmlListaSpawnGlobal
		errXml := xml.Unmarshal(dados, &lista)
		if errXml != nil {
			logger.Warnf("Falha ao parsear XML de spawn %s: %v", arquivo, errXml)
			continue
		}
		territoriosArquivo := 0
		makersLidosArquivo := 0
		makersValidosArquivo := 0
		npcsLidosArquivo := 0
		npcsValidosArquivo := 0
		for _, territorioXml := range lista.Territorios {
			territorio := territorioSpawn{nome: strings.TrimSpace(territorioXml.Name), minZ: parseInt32Seguro(territorioXml.MinZ), maxZ: parseInt32Seguro(territorioXml.MaxZ), nos: make([]pontoTerritorio, 0, len(territorioXml.Nos))}
			for _, no := range territorioXml.Nos {
				ponto := pontoTerritorio{x: parseInt32Seguro(no.X), y: parseInt32Seguro(no.Y)}
				territorio.nos = append(territorio.nos, ponto)
				if len(territorio.nos) == 1 {
					territorio.minX = ponto.x
					territorio.maxX = ponto.x
					territorio.minY = ponto.y
					territorio.maxY = ponto.y
					continue
				}
				if ponto.x < territorio.minX {
					territorio.minX = ponto.x
				}
				if ponto.x > territorio.maxX {
					territorio.maxX = ponto.x
				}
				if ponto.y < territorio.minY {
					territorio.minY = ponto.y
				}
				if ponto.y > territorio.maxY {
					territorio.maxY = ponto.y
				}
			}
			if territorio.nome == "" {
				continue
			}
			if len(territorio.nos) == 0 {
				continue
			}
			template.territorios[territorio.nome] = territorio
			territoriosArquivo++
			totalTerritorios++
		}
		for _, makerXml := range lista.NpcMakers {
			makersLidosArquivo++
			totalMakersLidos++
			maker := makerSpawnGlobalTemplate{nome: strings.TrimSpace(makerXml.Name), territorioNome: strings.TrimSpace(makerXml.Territory), evento: strings.TrimSpace(makerXml.Event), maximumNpcs: parseInt32Seguro(strings.TrimSpace(makerXml.MaximumNpcs)), arquivoOrigem: filepath.Base(arquivo), aiParams: novoNpcAiParametros(), npcs: make([]npcSpawnGlobalTemplate, 0, len(makerXml.Npcs))}
			makerTemPosFixa := false
			for _, ai := range makerXml.AIs {
				maker.tipoAI = strings.TrimSpace(ai.Type)
				for _, aiSet := range ai.Sets {
					maker.aiParams.aplicar(aiSet.Name, aiSet.Val)
				}
				break
			}
			for _, npcXml := range makerXml.Npcs {
				npcsLidosArquivo++
				totalNpcsLidos++
				npcTemplate := npcSpawnGlobalTemplate{npcID: parseInt32Seguro(npcXml.ID), total: parseInt32Seguro(npcXml.Total), respawn: strings.TrimSpace(npcXml.Respawn), respawnRand: strings.TrimSpace(npcXml.RespawnRand), pos: strings.TrimSpace(npcXml.Pos), dbName: strings.TrimSpace(npcXml.DbName), dbSaving: strings.TrimSpace(npcXml.DbSaving), arquivoOrigem: filepath.Base(arquivo), aiParams: novoNpcAiParametros()}
				if npcTemplate.pos != "" {
					makerTemPosFixa = true
				}
				for _, aiSet := range npcXml.AiSets {
					npcTemplate.aiParams.aplicar(aiSet.Name, aiSet.Val)
				}
				maker.npcs = append(maker.npcs, npcTemplate)
				npcsValidosArquivo++
				totalNpcsValidos++
			}
			if maker.nome == "" {
				logger.Warnf("Maker de spawn ignorado por nome vazio arquivo=%s territorio=%s", filepath.Base(arquivo), maker.territorioNome)
				continue
			}
			if maker.territorioNome == "" && !makerTemPosFixa {
				logger.Warnf("Maker de spawn ignorado por territorio vazio sem pos fixa arquivo=%s maker=%s", filepath.Base(arquivo), maker.nome)
				continue
			}
			if maker.territorioNome == "" && makerTemPosFixa {
				logger.Infof("Maker de spawn fixo sem territorio aceito arquivo=%s maker=%s npcs=%d", filepath.Base(arquivo), maker.nome, len(maker.npcs))
			}
			if len(maker.npcs) == 0 {
				logger.Warnf("Maker de spawn ignorado sem NPCs arquivo=%s maker=%s territorio=%s", filepath.Base(arquivo), maker.nome, maker.territorioNome)
				continue
			}
			template.makers = append(template.makers, maker)
			makersValidosArquivo++
			totalMakersValidos++
		}
		logger.Infof("Spawn XML carregado arquivo=%s territorios=%d makersLidos=%d makersValidos=%d npcsLidos=%d npcsValidos=%d", filepath.Base(arquivo), territoriosArquivo, makersLidosArquivo, makersValidosArquivo, npcsLidosArquivo, npcsValidosArquivo)
	}
	spawnGlobalMu.Lock()
	spawnGlobalAtual = template
	spawnGlobalMu.Unlock()
	logger.Infof("Templates globais de spawn carregados: territorios=%d makers=%d", len(template.territorios), len(template.makers))
	logger.Infof("Resumo loading spawn global: territorios=%d makersLidos=%d makersValidos=%d npcsLidos=%d npcsValidos=%d", totalTerritorios, totalMakersLidos, totalMakersValidos, totalNpcsLidos, totalNpcsValidos)
	return nil
}

func obterTemplatesSpawnGlobal() spawnGlobalTemplate {
	spawnGlobalMu.RLock()
	template := spawnGlobalAtual
	spawnGlobalMu.RUnlock()
	return template
}
