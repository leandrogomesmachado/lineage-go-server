package network

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type metadataSkill struct {
	skillID     int32
	niveis      int32
	operateType string
	passiva     bool
	toggle      bool
}

type xmlListaSkillsMetadata struct {
	Skills []xmlSkillMetadata `xml:"skill"`
}

type xmlSkillMetadata struct {
	ID     string             `xml:"id,attr"`
	Levels string             `xml:"levels,attr"`
	Sets   []xmlSetSkillValor `xml:"set"`
}

type xmlSetSkillValor struct {
	Name string `xml:"name,attr"`
	Val  string `xml:"val,attr"`
}

var metadadosSkills = map[int32]metadataSkill{}
var metadadosSkillsMu sync.RWMutex

func carregarMetadadosSkills(datapackPath string) error {
	skillsPath := filepath.Join(datapackPath, "data", "xml", "skills")
	arquivos, err := filepath.Glob(filepath.Join(skillsPath, "*.xml"))
	if err != nil {
		return err
	}
	novoMapa := make(map[int32]metadataSkill)
	for _, arquivo := range arquivos {
		dados, errLeitura := os.ReadFile(arquivo)
		if errLeitura != nil {
			logger.Warnf("Falha ao ler XML de skill %s: %v", arquivo, errLeitura)
			continue
		}
		var lista xmlListaSkillsMetadata
		errXml := xml.Unmarshal(dados, &lista)
		if errXml != nil {
			logger.Warnf("Falha ao parsear XML de skill %s: %v", arquivo, errXml)
			continue
		}
		for _, skill := range lista.Skills {
			metadata, ok := converterXmlSkillParaMetadata(skill)
			if !ok {
				continue
			}
			novoMapa[metadata.skillID] = metadata
		}
	}
	metadadosSkillsMu.Lock()
	metadadosSkills = novoMapa
	metadadosSkillsMu.Unlock()
	logger.Infof("Metadados de skills carregados: %d", len(novoMapa))
	return nil
}

func converterXmlSkillParaMetadata(skill xmlSkillMetadata) (metadataSkill, bool) {
	skillID := parseInt32SeguroSkill(skill.ID)
	if skillID <= 0 {
		return metadataSkill{}, false
	}
	operateType := "ACTIVE"
	for _, item := range skill.Sets {
		if !strings.EqualFold(item.Name, "operateType") {
			continue
		}
		operateType = normalizarOperateTypeSkill(item.Val)
		break
	}
	metadata := metadataSkill{
		skillID:     skillID,
		niveis:      parseInt32SeguroSkill(skill.Levels),
		operateType: operateType,
	}
	if operateType == "PASSIVE" {
		metadata.passiva = true
	}
	if operateType == "TOGGLE" {
		metadata.toggle = true
	}
	return metadata, true
}

func normalizarOperateTypeSkill(valor string) string {
	texto := strings.ToUpper(strings.TrimSpace(valor))
	if strings.HasPrefix(texto, "OP_") {
		texto = strings.TrimPrefix(texto, "OP_")
	}
	if texto == "PASSIVE" {
		return "PASSIVE"
	}
	if texto == "TOGGLE" {
		return "TOGGLE"
	}
	return "ACTIVE"
}

func obterMetadataSkill(skillID int32) (metadataSkill, bool) {
	metadadosSkillsMu.RLock()
	metadata, ok := metadadosSkills[skillID]
	metadadosSkillsMu.RUnlock()
	return metadata, ok
}

func skillEhPassiva(skillID int32, skillLevel int32) bool {
	_ = skillLevel
	metadata, ok := obterMetadataSkill(skillID)
	if !ok {
		return false
	}
	return metadata.passiva
}

func skillEhToggle(skillID int32, skillLevel int32) bool {
	_ = skillLevel
	metadata, ok := obterMetadataSkill(skillID)
	if !ok {
		return false
	}
	return metadata.toggle
}
