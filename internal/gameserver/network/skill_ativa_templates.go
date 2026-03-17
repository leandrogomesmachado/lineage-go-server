package network

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type templateSkillAtiva struct {
	skillID    int32
	nome       string
	niveis     int32
	hitTime    int32
	reuseDelay int32
	castRange  int32
	skillType  string
	targetType string
	isMagic    bool
	powerBase  int32
	powerNivel []int32
	mpConsume  int32
}

type xmlListaSkillsAtivas struct {
	Skills []xmlSkillAtiva `xml:"skill"`
}

type xmlSkillAtiva struct {
	ID     string             `xml:"id,attr"`
	Levels string             `xml:"levels,attr"`
	Name   string             `xml:"name,attr"`
	Sets   []xmlSetSkillValor `xml:"set"`
	Tables []xmlTabelaSkill   `xml:"table"`
}

type xmlTabelaSkill struct {
	Name string `xml:"name,attr"`
	Val  string `xml:",chardata"`
}

var templatesSkillsAtivas = map[int32]templateSkillAtiva{}
var templatesSkillsAtivasMu sync.RWMutex

func carregarTemplatesSkillsAtivas(datapackPath string) error {
	skillsPath := filepath.Join(datapackPath, "data", "xml", "skills")
	arquivos, err := filepath.Glob(filepath.Join(skillsPath, "*.xml"))
	if err != nil {
		return err
	}
	novoMapa := make(map[int32]templateSkillAtiva)
	for _, arquivo := range arquivos {
		dados, errLeitura := os.ReadFile(arquivo)
		if errLeitura != nil {
			logger.Warnf("Falha ao ler skill ativa %s: %v", arquivo, errLeitura)
			continue
		}
		var lista xmlListaSkillsAtivas
		if errXml := xml.Unmarshal(dados, &lista); errXml != nil {
			continue
		}
		for _, skillXml := range lista.Skills {
			tmpl, ok := converterXmlSkillAtiva(skillXml)
			if !ok {
				continue
			}
			novoMapa[tmpl.skillID] = tmpl
		}
	}
	templatesSkillsAtivasMu.Lock()
	templatesSkillsAtivas = novoMapa
	templatesSkillsAtivasMu.Unlock()
	logger.Infof("Templates de skills ativas carregados: %d", len(novoMapa))
	return nil
}

func converterXmlSkillAtiva(skillXml xmlSkillAtiva) (templateSkillAtiva, bool) {
	skillID := parseInt32SeguroSkill(skillXml.ID)
	if skillID <= 0 {
		return templateSkillAtiva{}, false
	}
	tabelas := map[string][]int32{}
	for _, t := range skillXml.Tables {
		nome := strings.TrimSpace(t.Name)
		vals := parseListaInteiros(t.Val)
		tabelas[nome] = vals
	}
	niveis := parseInt32SeguroSkill(skillXml.Levels)
	if niveis <= 0 {
		niveis = 1
	}
	tmpl := templateSkillAtiva{
		skillID: skillID,
		nome:    strings.TrimSpace(skillXml.Name),
		niveis:  niveis,
	}
	for _, s := range skillXml.Sets {
		nome := strings.ToLower(strings.TrimSpace(s.Name))
		val := strings.TrimSpace(s.Val)
		switch nome {
		case "hittime":
			tmpl.hitTime = parseInt32SeguroSkill(val)
		case "reusedelay":
			tmpl.reuseDelay = parseInt32SeguroSkill(val)
		case "castrange":
			tmpl.castRange = parseInt32SeguroSkill(val)
		case "skilltype":
			tmpl.skillType = strings.ToUpper(val)
		case "target":
			tmpl.targetType = strings.ToUpper(val)
		case "ismagic":
			tmpl.isMagic = strings.EqualFold(val, "true") || val == "1"
		case "power":
			if strings.HasPrefix(val, "#") {
				tabelaVals, ok := tabelas[val]
				if ok {
					tmpl.powerNivel = tabelaVals
					if len(tabelaVals) > 0 {
						tmpl.powerBase = tabelaVals[0]
					}
				}
			} else {
				tmpl.powerBase = parseInt32SeguroSkill(val)
			}
		case "mpconsume":
			if strings.HasPrefix(val, "#") {
				tabelaVals, ok := tabelas[val]
				if ok && len(tabelaVals) > 0 {
					tmpl.mpConsume = tabelaVals[0]
				}
			} else {
				tmpl.mpConsume = parseInt32SeguroSkill(val)
			}
		}
	}
	return tmpl, true
}

func obterTemplateSkillAtiva(skillID int32) (templateSkillAtiva, bool) {
	templatesSkillsAtivasMu.RLock()
	tmpl, ok := templatesSkillsAtivas[skillID]
	templatesSkillsAtivasMu.RUnlock()
	return tmpl, ok
}

func (t templateSkillAtiva) obterPoderPorNivel(nivel int32) int32 {
	idx := int(nivel) - 1
	if idx >= 0 && idx < len(t.powerNivel) {
		return t.powerNivel[idx]
	}
	if t.powerBase > 0 {
		return t.powerBase
	}
	return 1
}

func parseListaInteiros(texto string) []int32 {
	partes := strings.Fields(strings.TrimSpace(texto))
	resultado := make([]int32, 0, len(partes))
	for _, p := range partes {
		n, err := strconv.ParseInt(p, 10, 64)
		if err == nil {
			resultado = append(resultado, int32(n))
		}
	}
	return resultado
}
