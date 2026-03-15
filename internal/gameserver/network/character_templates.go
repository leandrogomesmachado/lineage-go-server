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

type templatePersonagemInicial struct {
	nome            string
	classID         int32
	baseLvl         int32
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
	spawns          []locSpawnInicial
}

var templatesPersonagemInicialPadrao = []templatePersonagemInicial{
	{nome: "Human Fighter", classID: 0, baseLvl: 1, race: 0, str: 40, dex: 30, con: 43, intel: 21, wit: 11, men: 25, x: -71338, y: 258271, z: -3104, maxHp: 80, maxMp: 30, maxCp: 32, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 115, walkSpd: 80, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 9, radiusFeminino: 8, heightMasculino: 23, heightFeminino: 23.5, spawns: []locSpawnInicial{{x: -71338, y: 258271, z: -3104}, {x: -71417, y: 258270, z: -3104}, {x: -71453, y: 258305, z: -3104}, {x: -71467, y: 258378, z: -3104}}},
	{nome: "Human Mystic", classID: 10, baseLvl: 1, race: 0, str: 22, dex: 21, con: 27, intel: 41, wit: 20, men: 39, x: -90875, y: 248162, z: -3570, maxHp: 101, maxMp: 40, maxCp: 51, pAtk: 3, pDef: 54, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 120, walkSpd: 80, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 6.5, heightMasculino: 22.8, heightFeminino: 22.5, spawns: []locSpawnInicial{{x: -90875, y: 248162, z: -3570}, {x: -90954, y: 248118, z: -3570}, {x: -90918, y: 248070, z: -3570}, {x: -90890, y: 248027, z: -3570}}},
	{nome: "Elven Fighter", classID: 18, baseLvl: 1, race: 1, str: 36, dex: 35, con: 36, intel: 23, wit: 14, men: 26, x: 46045, y: 41251, z: -3440, maxHp: 89, maxMp: 30, maxCp: 36, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 122, walkSpd: 85, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 7.5, heightMasculino: 24, heightFeminino: 23, spawns: []locSpawnInicial{{x: 46045, y: 41251, z: -3440}, {x: 46117, y: 41247, z: -3440}, {x: 46182, y: 41198, z: -3440}, {x: 46115, y: 41141, z: -3440}, {x: 46048, y: 41141, z: -3440}, {x: 45978, y: 41196, z: -3440}}},
	{nome: "Elven Mystic", classID: 25, baseLvl: 1, race: 1, str: 21, dex: 24, con: 25, intel: 37, wit: 23, men: 40, x: 46045, y: 41251, z: -3440, maxHp: 104, maxMp: 40, maxCp: 52, pAtk: 3, pDef: 54, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 122, walkSpd: 85, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 7.5, heightMasculino: 23.5, heightFeminino: 22.5, spawns: []locSpawnInicial{{x: 46045, y: 41251, z: -3440}, {x: 46117, y: 41247, z: -3440}, {x: 46182, y: 41198, z: -3440}, {x: 46115, y: 41141, z: -3440}, {x: 46048, y: 41141, z: -3440}, {x: 45978, y: 41196, z: -3440}}},
	{nome: "Dark Fighter", classID: 31, baseLvl: 1, race: 2, str: 41, dex: 34, con: 32, intel: 25, wit: 12, men: 26, x: 28295, y: 11063, z: -4224, maxHp: 94, maxMp: 30, maxCp: 38, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 122, walkSpd: 85, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 7, heightMasculino: 24, heightFeminino: 23.5, spawns: []locSpawnInicial{{x: 28295, y: 11063, z: -4224}, {x: 28302, y: 11008, z: -4224}, {x: 28377, y: 10916, z: -4224}, {x: 28456, y: 10997, z: -4224}, {x: 28461, y: 11044, z: -4224}, {x: 28395, y: 11127, z: -4224}}},
	{nome: "Dark Mystic", classID: 38, baseLvl: 1, race: 2, str: 23, dex: 23, con: 24, intel: 44, wit: 19, men: 37, x: 28295, y: 11063, z: -4224, maxHp: 106, maxMp: 40, maxCp: 53, pAtk: 3, pDef: 54, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 122, walkSpd: 85, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 7, heightMasculino: 24, heightFeminino: 23.5, spawns: []locSpawnInicial{{x: 28295, y: 11063, z: -4224}, {x: 28302, y: 11008, z: -4224}, {x: 28377, y: 10916, z: -4224}, {x: 28456, y: 10997, z: -4224}, {x: 28461, y: 11044, z: -4224}, {x: 28395, y: 11127, z: -4224}}},
	{nome: "Orc Fighter", classID: 44, baseLvl: 1, race: 3, str: 40, dex: 26, con: 47, intel: 18, wit: 12, men: 27, x: -56733, y: -113459, z: -690, maxHp: 80, maxMp: 30, maxCp: 32, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 117, walkSpd: 80, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 11, radiusFeminino: 7, heightMasculino: 28, heightFeminino: 25, spawns: []locSpawnInicial{{x: -56733, y: -113459, z: -690}, {x: -56686, y: -113470, z: -690}, {x: -56728, y: -113610, z: -690}, {x: -56693, y: -113610, z: -690}, {x: -56743, y: -113757, z: -690}, {x: -56682, y: -113730, z: -690}}},
	{nome: "Orc Mystic", classID: 49, baseLvl: 1, race: 3, str: 27, dex: 24, con: 31, intel: 31, wit: 15, men: 42, x: -56733, y: -113459, z: -690, maxHp: 95, maxMp: 40, maxCp: 47, pAtk: 3, pDef: 54, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 117, walkSpd: 80, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 7, radiusFeminino: 8, heightMasculino: 24, heightFeminino: 26, spawns: []locSpawnInicial{{x: -56733, y: -113459, z: -690}, {x: -56686, y: -113470, z: -690}, {x: -56728, y: -113610, z: -690}, {x: -56693, y: -113610, z: -690}, {x: -56743, y: -113757, z: -690}, {x: -56682, y: -113730, z: -690}}},
	{nome: "Dwarf Fighter", classID: 53, baseLvl: 1, race: 4, str: 39, dex: 29, con: 45, intel: 20, wit: 10, men: 27, x: 108644, y: -173947, z: -400, maxHp: 80, maxMp: 30, maxCp: 32, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, baseAtkSpd: 300, baseCrit: 4, runSpd: 126, walkSpd: 87, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 9, radiusFeminino: 5, heightMasculino: 18.5, heightFeminino: 19, spawns: []locSpawnInicial{{x: 108644, y: -173947, z: -400}, {x: 108678, y: -174002, z: -400}, {x: 108505, y: -173964, z: -400}, {x: 108512, y: -174026, z: -400}, {x: 108549, y: -174075, z: -400}, {x: 108576, y: -174122, z: -400}}},
}

var templatesPersonagemInicial = clonarTemplatesPadrao(templatesPersonagemInicialPadrao)
var templatesPersonagemInicialMu sync.RWMutex

type xmlListaClasses struct {
	Classes []xmlClasse `xml:"class"`
}

type xmlClasse struct {
	Sets       []xmlSetClasse   `xml:"set"`
	Spawns     []xmlSpawnClasse `xml:"spawns>spawn"`
	Comentario string           `xml:",comment"`
}

type xmlSetClasse struct {
	ID           string `xml:"id,attr"`
	BaseLvl      string `xml:"baseLvl,attr"`
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
}

type xmlSpawnClasse struct {
	X string `xml:"x,attr"`
	Y string `xml:"y,attr"`
	Z string `xml:"z,attr"`
}

func clonarTemplatesPadrao(origem []templatePersonagemInicial) map[int32]templatePersonagemInicial {
	resultado := make(map[int32]templatePersonagemInicial, len(origem))
	for _, template := range origem {
		resultado[template.classID] = template
	}
	return resultado
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
	novoMapa := clonarTemplatesPadrao(templatesPersonagemInicialPadrao)
	classesPath := filepath.Join(datapackPath, "data", "xml", "classes")
	arquivos, err := filepath.Glob(filepath.Join(classesPath, "*.xml"))
	if err != nil {
		return err
	}
	if len(arquivos) == 0 {
		logger.Warnf("Nenhum XML de classe encontrado em %s, usando templates padrao em memoria", classesPath)
		templatesPersonagemInicialMu.Lock()
		templatesPersonagemInicial = novoMapa
		templatesPersonagemInicialMu.Unlock()
		return nil
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
	}
	if template.classID == 0 && !strings.Contains(strings.ToLower(template.nome), "human fighter") {
		return templatePersonagemInicial{}, false
	}
	template.race = obterRacePorClassID(template.classID)
	template.baseAtkSpd = 300
	template.baseCrit = 4
	template.pAtkSpd = template.baseAtkSpd
	template.mAtkSpd = 333
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
