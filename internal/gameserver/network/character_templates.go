package network

import (
	"encoding/xml"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type locSpawnInicial struct {
	x int32
	y int32
	z int32
}

type itemInicialClasse struct {
	itemID       int32
	count        int64
	estaEquipado bool
}

type templatePersonagemInicial struct {
	nome            string
	classID         int32
	baseLvl         int32
	fistsItemID     int32
	race            int32
	str             int32
	dex             int32
	con             int32
	intel           int32
	wit             int32
	men             int32
	x               int32
	y               int32
	z               int32
	maxHp           int32
	maxMp           int32
	maxCp           int32
	pAtk            int32
	pDef            int32
	mAtk            int32
	mDef            int32
	baseAtkSpd      int32
	baseCrit        int32
	runSpd          int32
	walkSpd         int32
	swimSpd         int32
	pAtkSpd         int32
	mAtkSpd         int32
	radiusMasculino float64
	radiusFeminino  float64
	heightMasculino float64
	heightFeminino  float64
	hpTable         []int32
	mpTable         []int32
	cpTable         []int32
	hpRegenTable    []int32
	mpRegenTable    []int32
	cpRegenTable    []int32
	spawns          []locSpawnInicial
	itensIniciais   []itemInicialClasse
}

var templatesPersonagemInicial = map[int32]templatePersonagemInicial{}
var templatesPersonagemInicialMu sync.RWMutex

type xmlListaClasses struct {
	Classes []xmlClasse `xml:"class"`
}

type xmlClasse struct {
	Sets       []xmlSetClasse   `xml:"set"`
	Spawns     []xmlSpawnClasse `xml:"spawns>spawn"`
	Itens      []xmlItemInicial `xml:"items>item"`
	Comentario string           `xml:",comment"`
}

type xmlItemInicial struct {
	ID         string `xml:"id,attr"`
	Count      string `xml:"count,attr"`
	IsEquipped string `xml:"isEquipped,attr"`
}

type xmlSetClasse struct {
	ID           string `xml:"id,attr"`
	BaseLvl      string `xml:"baseLvl,attr"`
	Fists        string `xml:"fists,attr"`
	Str          string `xml:"str,attr"`
	Con          string `xml:"con,attr"`
	Dex          string `xml:"dex,attr"`
	Intel        string `xml:"int,attr"`
	Wit          string `xml:"wit,attr"`
	Men          string `xml:"men,attr"`
	PAtk         string `xml:"pAtk,attr"`
	PDef         string `xml:"pDef,attr"`
	MAtk         string `xml:"mAtk,attr"`
	MDef         string `xml:"mDef,attr"`
	RunSpd       string `xml:"runSpd,attr"`
	WalkSpd      string `xml:"walkSpd,attr"`
	SwimSpd      string `xml:"swimSpd,attr"`
	Radius       string `xml:"radius,attr"`
	RadiusFemale string `xml:"radiusFemale,attr"`
	Height       string `xml:"height,attr"`
	HeightFemale string `xml:"heightFemale,attr"`
	HpTable      string `xml:"hpTable,attr"`
	MpTable      string `xml:"mpTable,attr"`
	CpTable      string `xml:"cpTable,attr"`
	HpRegenTable string `xml:"hpRegenTable,attr"`
	MpRegenTable string `xml:"mpRegenTable,attr"`
	CpRegenTable string `xml:"cpRegenTable,attr"`
}

type xmlSpawnClasse struct {
	X string `xml:"x,attr"`
	Y string `xml:"y,attr"`
	Z string `xml:"z,attr"`
}

func listarTemplatesPersonagemInicial() []templatePersonagemInicial {
	templatesPersonagemInicialMu.RLock()
	defer templatesPersonagemInicialMu.RUnlock()
	lista := make([]templatePersonagemInicial, 0, len(templatesPersonagemInicial))
	for _, template := range templatesPersonagemInicial {
		lista = append(lista, template)
	}
	sort.Slice(lista, func(i int, j int) bool {
		return lista[i].classID < lista[j].classID
	})
	return lista
}

func carregarTemplatesPersonagemInicial(datapackPath string) error {
	novoMapa := make(map[int32]templatePersonagemInicial)
	classesPath := filepath.Join(datapackPath, "data", "xml", "classes")
	arquivos, err := filepath.Glob(filepath.Join(classesPath, "*.xml"))
	if err != nil {
		return err
	}
	if len(arquivos) == 0 {
		return os.ErrNotExist
	}
	for _, arquivo := range arquivos {
		templatesArquivo, errArquivo := carregarTemplatesArquivoClasse(arquivo)
		if errArquivo != nil {
			logger.Warnf("Falha ao carregar XML de classe %s: %v", arquivo, errArquivo)
			continue
		}
		for _, template := range templatesArquivo {
			novoMapa[template.classID] = template
		}
	}
	if len(novoMapa) == 0 {
		return os.ErrNotExist
	}
	templatesPersonagemInicialMu.Lock()
	templatesPersonagemInicial = novoMapa
	templatesPersonagemInicialMu.Unlock()
	logger.Infof("Templates de personagem carregados: %d", len(novoMapa))
	return nil
}

func carregarTemplatesArquivoClasse(caminhoArquivo string) ([]templatePersonagemInicial, error) {
	dados, err := os.ReadFile(caminhoArquivo)
	if err != nil {
		return nil, err
	}
	var lista xmlListaClasses
	err = xml.Unmarshal(dados, &lista)
	if err != nil {
		return nil, err
	}
	templates := make([]templatePersonagemInicial, 0, len(lista.Classes))
	for _, classe := range lista.Classes {
		template, ok := converterXmlClasseParaTemplate(classe)
		if !ok {
			continue
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func converterXmlClasseParaTemplate(classe xmlClasse) (templatePersonagemInicial, bool) {
	template := templatePersonagemInicial{}
	template.nome = strings.TrimSpace(classe.Comentario)
	for _, item := range classe.Sets {
		if item.ID != "" {
			template.classID = parseInt32Seguro(item.ID)
			template.baseLvl = parseInt32Seguro(item.BaseLvl)
			template.fistsItemID = parseInt32Seguro(item.Fists)
		}
		if item.Str != "" {
			template.str = parseInt32Seguro(item.Str)
			template.con = parseInt32Seguro(item.Con)
			template.dex = parseInt32Seguro(item.Dex)
			template.intel = parseInt32Seguro(item.Intel)
			template.wit = parseInt32Seguro(item.Wit)
			template.men = parseInt32Seguro(item.Men)
		}
		if item.PAtk != "" {
			template.pAtk = parseInt32Seguro(item.PAtk)
			template.pDef = parseInt32Seguro(item.PDef)
			template.mAtk = parseInt32Seguro(item.MAtk)
			template.mDef = parseInt32Seguro(item.MDef)
			template.runSpd = parseInt32Seguro(item.RunSpd)
			template.walkSpd = parseInt32Seguro(item.WalkSpd)
			template.swimSpd = parseInt32Seguro(item.SwimSpd)
		}
		if item.Radius != "" {
			template.radiusMasculino = parseFloat64Seguro(item.Radius)
			template.radiusFeminino = parseFloat64Seguro(item.RadiusFemale)
		}
		if item.Height != "" {
			template.heightMasculino = parseFloat64Seguro(item.Height)
			template.heightFeminino = parseFloat64Seguro(item.HeightFemale)
		}
		if item.HpTable != "" {
			template.hpTable = parseTabelaNivel(item.HpTable)
			template.maxHp = obterValorTabelaPorNivel(template.hpTable, template.baseLvl)
		}
		if item.MpTable != "" {
			template.mpTable = parseTabelaNivel(item.MpTable)
			template.maxMp = obterValorTabelaPorNivel(template.mpTable, template.baseLvl)
		}
		if item.CpTable != "" {
			template.cpTable = parseTabelaNivel(item.CpTable)
			template.maxCp = obterValorTabelaPorNivel(template.cpTable, template.baseLvl)
		}
		if item.HpRegenTable != "" {
			template.hpRegenTable = parseTabelaNivel(item.HpRegenTable)
		}
		if item.MpRegenTable != "" {
			template.mpRegenTable = parseTabelaNivel(item.MpRegenTable)
		}
		if item.CpRegenTable != "" {
			template.cpRegenTable = parseTabelaNivel(item.CpRegenTable)
		}
	}
	if template.classID == 0 && !strings.Contains(strings.ToLower(template.nome), "human fighter") {
		return templatePersonagemInicial{}, false
	}
	if template.baseLvl <= 0 {
		template.baseLvl = 1
	}
	if len(template.hpTable) == 0 {
		template.hpTable = []int32{1}
	}
	if len(template.mpTable) == 0 {
		template.mpTable = []int32{1}
	}
	if len(template.cpTable) == 0 {
		template.cpTable = []int32{0}
	}
	if len(template.hpRegenTable) == 0 {
		template.hpRegenTable = []int32{1}
	}
	if len(template.mpRegenTable) == 0 {
		template.mpRegenTable = []int32{1}
	}
	if len(template.cpRegenTable) == 0 {
		template.cpRegenTable = []int32{1}
	}
	template.maxHp = obterValorTabelaPorNivel(template.hpTable, template.baseLvl)
	template.maxMp = obterValorTabelaPorNivel(template.mpTable, template.baseLvl)
	template.maxCp = obterValorTabelaPorNivel(template.cpTable, template.baseLvl)
	template.race = obterRacePorClassID(template.classID)
	template.baseAtkSpd = obterBaseAtkSpdClasse(template.fistsItemID)
	if template.baseCrit <= 0 {
		template.baseCrit = 4
	}
	template.pAtkSpd = template.baseAtkSpd
	if template.mAtkSpd <= 0 {
		template.mAtkSpd = 333
	}
	if template.runSpd <= 0 {
		template.runSpd = 120
	}
	if template.walkSpd <= 0 {
		template.walkSpd = 80
	}
	if template.swimSpd <= 0 {
		template.swimSpd = 50
	}
	for _, itemXml := range classe.Itens {
		itemID := parseInt32Seguro(itemXml.ID)
		if itemID <= 0 {
			continue
		}
		count := int64(parseInt32Seguro(itemXml.Count))
		if count <= 0 {
			count = 1
		}
		estaEquipado := itemXml.IsEquipped != "false"
		template.itensIniciais = append(template.itensIniciais, itemInicialClasse{
			itemID:       itemID,
			count:        count,
			estaEquipado: estaEquipado,
		})
	}
	for _, spawn := range classe.Spawns {
		loc := locSpawnInicial{
			x: parseInt32Seguro(spawn.X),
			y: parseInt32Seguro(spawn.Y),
			z: parseInt32Seguro(spawn.Z),
		}
		template.spawns = append(template.spawns, loc)
	}
	if len(template.spawns) > 0 {
		template.x = template.spawns[0].x
		template.y = template.spawns[0].y
		template.z = template.spawns[0].z
	}
	return template, true
}

func obterBaseAtkSpdClasse(itemID int32) int32 {
	if itemID <= 0 {
		return 300
	}
	pAtkSpd, ok := obterPAtkSpdArma(itemID)
	if ok && pAtkSpd > 0 {
		return pAtkSpd
	}
	return 300
}

func (t templatePersonagemInicial) obterHpRegenPorNivel(nivel int32) int32 {
	return obterValorTabelaPorNivel(t.hpRegenTable, nivel)
}

func (t templatePersonagemInicial) obterMpRegenPorNivel(nivel int32) int32 {
	return obterValorTabelaPorNivel(t.mpRegenTable, nivel)
}

func (t templatePersonagemInicial) obterCpRegenPorNivel(nivel int32) int32 {
	return obterValorTabelaPorNivel(t.cpRegenTable, nivel)
}

func obterRacePorClassID(classID int32) int32 {
	if classID >= 0 && classID <= 17 {
		return 0
	}
	if classID >= 18 && classID <= 30 {
		return 1
	}
	if classID >= 31 && classID <= 43 {
		return 2
	}
	if classID >= 44 && classID <= 52 {
		return 3
	}
	if classID >= 53 && classID <= 57 {
		return 4
	}
	return 0
}

func parseInt32Seguro(valor string) int32 {
	texto := strings.TrimSpace(valor)
	if texto == "" {
		return 0
	}
	inteiro, err := strconv.ParseInt(texto, 10, 32)
	if err != nil {
		return 0
	}
	return int32(inteiro)
}

func parseFloat64Seguro(valor string) float64 {
	texto := strings.TrimSpace(valor)
	if texto == "" {
		return 0
	}
	numero, err := strconv.ParseFloat(texto, 64)
	if err != nil {
		return 0
	}
	return numero
}

func parseTabelaNivel(valor string) []int32 {
	partes := strings.Split(valor, ";")
	resultado := make([]int32, 0, len(partes))
	for _, parte := range partes {
		texto := strings.TrimSpace(parte)
		if texto == "" {
			continue
		}
		numero, err := strconv.ParseFloat(texto, 64)
		if err != nil {
			continue
		}
		resultado = append(resultado, int32(math.Round(numero)))
	}
	return resultado
}

func obterValorTabelaPorNivel(tabela []int32, nivel int32) int32 {
	if len(tabela) == 0 {
		return 0
	}
	if nivel <= 1 {
		return tabela[0]
	}
	indice := int(nivel - 1)
	if indice < len(tabela) {
		return tabela[indice]
	}
	return tabela[len(tabela)-1]
}

func obterTemplatePersonagemInicial(classID int32) (templatePersonagemInicial, bool) {
	templatesPersonagemInicialMu.RLock()
	template, ok := templatesPersonagemInicial[classID]
	templatesPersonagemInicialMu.RUnlock()
	if ok {
		return template, true
	}
	return templatePersonagemInicial{}, false
}

func (t templatePersonagemInicial) obterColisao(sexo int32) (float64, float64) {
	if sexo == 0 {
		return t.radiusMasculino, t.heightMasculino
	}
	return t.radiusFeminino, t.heightFeminino
}

func (t templatePersonagemInicial) obterSpawnInicial(seletor int32) locSpawnInicial {
	if len(t.spawns) == 0 {
		return locSpawnInicial{x: t.x, y: t.y, z: t.z}
	}
	indice := int(seletor)
	if indice < 0 {
		indice = -indice
	}
	indice = indice % len(t.spawns)
	return t.spawns[indice]
}

func (t templatePersonagemInicial) obterCpMaximoPorNivel(nivel int32) int32 {
	valorTabela := obterValorTabelaPorNivel(t.cpTable, nivel)
	if valorTabela > 0 {
		return valorTabela
	}
	if nivel <= 1 {
		return t.maxCp
	}
	valorCalculado := float64(t.maxCp) + (float64(nivel-1) * (float64(t.maxCp) * 0.12))
	return int32(math.Round(valorCalculado))
}

func (t templatePersonagemInicial) obterHpMaximoPorNivel(nivel int32) int32 {
	valorTabela := obterValorTabelaPorNivel(t.hpTable, nivel)
	if valorTabela > 0 {
		return valorTabela
	}
	return t.maxHp
}

func (t templatePersonagemInicial) obterMpMaximoPorNivel(nivel int32) int32 {
	valorTabela := obterValorTabelaPorNivel(t.mpTable, nivel)
	if valorTabela > 0 {
		return valorTabela
	}
	return t.maxMp
}
