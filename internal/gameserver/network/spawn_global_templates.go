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
	nos  []pontoTerritorio
}

type pontoTerritorio struct {
	x int32
	y int32
}

type npcSpawnGlobalTemplate struct {
	npcID       int32
	total       int32
	respawn     string
	respawnRand string
	pos         string
}

type makerSpawnGlobalTemplate struct {
	nome           string
	territorioNome string
	tipoAI         string
	npcs           []npcSpawnGlobalTemplate
}

type spawnGlobalTemplate struct {
	territorios map[string]territorioSpawn
	makers      []makerSpawnGlobalTemplate
}

type xmlListaSpawnGlobal struct {
	Territorios []xmlTerritorioSpawn  `xml:"territory"`
	NpcMakers   []xmlNpcMakerSpawn    `xml:"npcmaker"`
}

type xmlTerritorioSpawn struct {
	Name string                `xml:"name,attr"`
	MinZ string                `xml:"minZ,attr"`
	MaxZ string                `xml:"maxZ,attr"`
	Nos  []xmlNoTerritorioSpawn `xml:"node"`
}

type xmlNoTerritorioSpawn struct {
	X string `xml:"x,attr"`
	Y string `xml:"y,attr"`
}

type xmlNpcMakerSpawn struct {
	Name      string               `xml:"name,attr"`
	Territory string               `xml:"territory,attr"`
	AIs       []xmlAiMakerSpawn    `xml:"ai"`
	Npcs      []xmlNpcSpawnGlobal  `xml:"npc"`
}

type xmlAiMakerSpawn struct {
	Type string `xml:"type,attr"`
}

type xmlNpcSpawnGlobal struct {
	ID          string `xml:"id,attr"`
	Total       string `xml:"total,attr"`
	Respawn     string `xml:"respawn,attr"`
	RespawnRand string `xml:"respawnRand,attr"`
	Pos         string `xml:"pos,attr"`
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
		for _, territorioXml := range lista.Territorios {
			territorio := territorioSpawn{nome: strings.TrimSpace(territorioXml.Name), minZ: parseInt32Seguro(territorioXml.MinZ), maxZ: parseInt32Seguro(territorioXml.MaxZ), nos: make([]pontoTerritorio, 0, len(territorioXml.Nos))}
			for _, no := range territorioXml.Nos {
				territorio.nos = append(territorio.nos, pontoTerritorio{x: parseInt32Seguro(no.X), y: parseInt32Seguro(no.Y)})
			}
			if territorio.nome == "" {
				continue
			}
			if len(territorio.nos) == 0 {
				continue
			}
			template.territorios[territorio.nome] = territorio
		}
		for _, makerXml := range lista.NpcMakers {
			maker := makerSpawnGlobalTemplate{nome: strings.TrimSpace(makerXml.Name), territorioNome: strings.TrimSpace(makerXml.Territory), npcs: make([]npcSpawnGlobalTemplate, 0, len(makerXml.Npcs))}
			for _, ai := range makerXml.AIs {
				maker.tipoAI = strings.TrimSpace(ai.Type)
				break
			}
			for _, npcXml := range makerXml.Npcs {
				maker.npcs = append(maker.npcs, npcSpawnGlobalTemplate{npcID: parseInt32Seguro(npcXml.ID), total: parseInt32Seguro(npcXml.Total), respawn: strings.TrimSpace(npcXml.Respawn), respawnRand: strings.TrimSpace(npcXml.RespawnRand), pos: strings.TrimSpace(npcXml.Pos)})
			}
			if maker.nome == "" {
				continue
			}
			if maker.territorioNome == "" {
				continue
			}
			if len(maker.npcs) == 0 {
				continue
			}
			template.makers = append(template.makers, maker)
		}
	}
	spawnGlobalMu.Lock()
	spawnGlobalAtual = template
	spawnGlobalMu.Unlock()
	logger.Infof("Templates globais de spawn carregados: territorios=%d makers=%d", len(template.territorios), len(template.makers))
	return nil
}

func obterTemplatesSpawnGlobal() spawnGlobalTemplate {
	spawnGlobalMu.RLock()
	template := spawnGlobalAtual
	spawnGlobalMu.RUnlock()
	return template
}
