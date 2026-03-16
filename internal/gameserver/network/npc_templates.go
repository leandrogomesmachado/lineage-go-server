package network

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type npcTemplate struct {
	npcID         int32
	idTemplate    int32
	nome          string
	titulo        string
	alias         string
	tipo          string
	nivel         int32
	hp            int32
	mp            int32
	exp           int64
	sp            int32
	aggroRange    int32
	radius        float64
	height        float64
	pAtk          int32
	pDef          int32
	mAtk          int32
	mDef          int32
	crit          int32
	runSpd        int32
	walkSpd       int32
	pAtkSpd       int32
	mAtkSpd       int32
	rHand         int32
	lHand         int32
	canMove       bool
	canBeAttacked bool
	canSeeThrough bool
	aiParams      npcAiParametros
	drops         []npcDropCategoriaTemplate
}

type npcDropCategoriaTemplate struct {
	tipo   string
	chance float64
	drops  []npcDropTemplate
}

type npcDropTemplate struct {
	itemID int32
	min    int64
	max    int64
	chance float64
}

type xmlListaNpcs struct {
	Npcs []xmlNpc `xml:"npc"`
}

type xmlNpc struct {
	ID         string                `xml:"id,attr"`
	IDTemplate string                `xml:"idTemplate,attr"`
	Name       string                `xml:"name,attr"`
	Title      string                `xml:"title,attr"`
	Alias      string                `xml:"alias,attr"`
	Sets       []xmlNpcSet           `xml:"set"`
	AiSets     []xmlNpcSet           `xml:"ai>set"`
	Drops      []xmlNpcDropCategoria `xml:"drops>category"`
	Skills     []xmlNpcSkill         `xml:"skills>skill"`
}

type xmlNpcSet struct {
	Name string `xml:"name,attr"`
	Val  string `xml:"val,attr"`
}

type xmlNpcSkill struct {
	ID    string `xml:"id,attr"`
	Level string `xml:"level,attr"`
	Type  string `xml:"type,attr"`
}

type xmlNpcDropCategoria struct {
	Type   string       `xml:"type,attr"`
	Chance string       `xml:"chance,attr"`
	Drops  []xmlNpcDrop `xml:"drop"`
}

type xmlNpcDrop struct {
	ItemID string `xml:"itemid,attr"`
	Min    string `xml:"min,attr"`
	Max    string `xml:"max,attr"`
	Chance string `xml:"chance,attr"`
}

var npcTemplates = map[int32]npcTemplate{}
var npcTemplatesMu sync.RWMutex

func carregarTemplatesNpc(datapackPath string) error {
	npcsPath := filepath.Join(datapackPath, "data", "xml", "npcs")
	arquivos, err := filepath.Glob(filepath.Join(npcsPath, "*.xml"))
	if err != nil {
		return err
	}
	novoMapa := make(map[int32]npcTemplate)
	for _, arquivo := range arquivos {
		dados, errLeitura := os.ReadFile(arquivo)
		if errLeitura != nil {
			logger.Warnf("Falha ao ler XML de NPC %s: %v", arquivo, errLeitura)
			continue
		}
		var lista xmlListaNpcs
		errXml := xml.Unmarshal(dados, &lista)
		if errXml != nil {
			logger.Warnf("Falha ao parsear XML de NPC %s: %v", arquivo, errXml)
			continue
		}
		for _, item := range lista.Npcs {
			template, ok := converterXmlNpcParaTemplate(item)
			if !ok {
				continue
			}
			novoMapa[template.npcID] = template
		}
	}
	npcTemplatesMu.Lock()
	npcTemplates = novoMapa
	npcTemplatesMu.Unlock()
	logger.Infof("Templates de NPC carregados: %d", len(novoMapa))
	return nil
}

func converterXmlNpcParaTemplate(item xmlNpc) (npcTemplate, bool) {
	npcID := parseInt32Seguro(item.ID)
	if npcID <= 0 {
		return npcTemplate{}, false
	}
	template := npcTemplate{
		npcID:         npcID,
		idTemplate:    npcID,
		nome:          strings.TrimSpace(item.Name),
		titulo:        strings.TrimSpace(item.Title),
		alias:         strings.TrimSpace(item.Alias),
		nivel:         1,
		radius:        8,
		height:        23,
		hp:            1,
		mp:            1,
		pAtk:          1,
		pDef:          1,
		mAtk:          1,
		mDef:          1,
		crit:          4,
		runSpd:        120,
		walkSpd:       40,
		pAtkSpd:       253,
		mAtkSpd:       333,
		rHand:         0,
		lHand:         0,
		canMove:       true,
		canBeAttacked: true,
		aiParams:      novoNpcAiParametros(),
	}
	idTemplate := int32(0)
	if strings.TrimSpace(item.IDTemplate) != "" {
		idTemplate = parseInt32Seguro(item.IDTemplate)
		if idTemplate > 0 {
			template.idTemplate = idTemplate
		}
	}
	for _, set := range item.Sets {
		nome := strings.TrimSpace(set.Name)
		valor := strings.TrimSpace(set.Val)
		if strings.EqualFold(nome, "type") {
			template.tipo = valor
			continue
		}
		if strings.EqualFold(nome, "level") {
			template.nivel = parseInt32Seguro(valor)
			continue
		}
		if strings.EqualFold(nome, "hp") {
			template.hp = int32(parseFloat64Seguro(valor))
			continue
		}
		if strings.EqualFold(nome, "mp") {
			template.mp = int32(parseFloat64Seguro(valor))
			continue
		}
		if strings.EqualFold(nome, "exp") {
			template.exp = int64(parseFloat64Seguro(valor))
			continue
		}
		if strings.EqualFold(nome, "sp") {
			template.sp = parseInt32Seguro(valor)
			continue
		}
		if strings.EqualFold(nome, "aggroRange") {
			template.aggroRange = parseInt32Seguro(valor)
			continue
		}
		if strings.EqualFold(nome, "radius") {
			radius := parseFloat64Seguro(valor)
			if radius > 0 {
				template.radius = radius
			}
			continue
		}
		if strings.EqualFold(nome, "height") {
			height := parseFloat64Seguro(valor)
			if height > 0 {
				template.height = height
			}
			continue
		}
		if strings.EqualFold(nome, "runSpd") {
			template.runSpd = parseInt32Seguro(valor)
			continue
		}
		if strings.EqualFold(nome, "pAtk") {
			template.pAtk = int32(parseFloat64Seguro(valor))
			continue
		}
		if strings.EqualFold(nome, "pDef") {
			template.pDef = int32(parseFloat64Seguro(valor))
			continue
		}
		if strings.EqualFold(nome, "mAtk") {
			template.mAtk = int32(parseFloat64Seguro(valor))
			continue
		}
		if strings.EqualFold(nome, "mDef") {
			template.mDef = int32(parseFloat64Seguro(valor))
			continue
		}
		if strings.EqualFold(nome, "crit") {
			template.crit = int32(parseFloat64Seguro(valor))
			continue
		}
		if strings.EqualFold(nome, "walkSpd") {
			template.walkSpd = parseInt32Seguro(valor)
			continue
		}
		if strings.EqualFold(nome, "rHand") {
			template.rHand = parseInt32Seguro(valor)
			continue
		}
		if strings.EqualFold(nome, "lHand") {
			template.lHand = parseInt32Seguro(valor)
			continue
		}
		if strings.EqualFold(nome, "atkSpd") {
			atkSpd := parseFloat64Seguro(valor)
			if atkSpd > 0 {
				template.pAtkSpd = int32(atkSpd)
			}
			continue
		}
		if strings.EqualFold(nome, "canMove") {
			template.canMove = strings.EqualFold(valor, "true")
			continue
		}
		if strings.EqualFold(nome, "canSeeThrough") {
			template.canSeeThrough = strings.EqualFold(valor, "true")
			template.aiParams.aplicar(nome, valor)
			continue
		}
		if strings.EqualFold(nome, "canBeAttacked") {
			template.canBeAttacked = strings.EqualFold(valor, "true")
		}
	}
	for _, aiSet := range item.AiSets {
		template.aiParams.aplicar(aiSet.Name, aiSet.Val)
	}
	for _, categoriaXml := range item.Drops {
		categoria := npcDropCategoriaTemplate{tipo: strings.TrimSpace(categoriaXml.Type), chance: parseFloat64Seguro(strings.TrimSpace(categoriaXml.Chance)), drops: make([]npcDropTemplate, 0, len(categoriaXml.Drops))}
		for _, dropXml := range categoriaXml.Drops {
			drop := npcDropTemplate{itemID: parseInt32Seguro(strings.TrimSpace(dropXml.ItemID)), min: int64(parseInt32Seguro(strings.TrimSpace(dropXml.Min))), max: int64(parseInt32Seguro(strings.TrimSpace(dropXml.Max))), chance: parseFloat64Seguro(strings.TrimSpace(dropXml.Chance))}
			if drop.itemID <= 0 {
				continue
			}
			if drop.min <= 0 {
				drop.min = 1
			}
			if drop.max < drop.min {
				drop.max = drop.min
			}
			categoria.drops = append(categoria.drops, drop)
		}
		if len(categoria.drops) == 0 {
			continue
		}
		template.drops = append(template.drops, categoria)
	}
	if template.nome == "" {
		template.nome = "NPC"
	}
	return template, true
}

func obterTemplateNpc(npcID int32) (npcTemplate, bool) {
	npcTemplatesMu.RLock()
	template, ok := npcTemplates[npcID]
	npcTemplatesMu.RUnlock()
	return template, ok
}

func (t npcTemplate) ehMonster() bool {
	if strings.EqualFold(t.tipo, "Monster") {
		return true
	}
	if strings.EqualFold(t.tipo, "RaidBoss") {
		return true
	}
	if strings.EqualFold(t.tipo, "GrandBoss") {
		return true
	}
	return false
}
