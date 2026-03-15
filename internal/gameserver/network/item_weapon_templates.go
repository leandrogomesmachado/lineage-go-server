package network

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type itemWeaponTemplate struct {
	itemID  int32
	pAtkSpd int32
}

type xmlListaItems struct {
	Items []xmlItemWeapon `xml:"item"`
}

type xmlItemWeapon struct {
	ID   string             `xml:"id,attr"`
	Type string             `xml:"type,attr"`
	For  []xmlForWeaponStat `xml:"for"`
}

type xmlForWeaponStat struct {
	Sets []xmlSetWeaponStat `xml:"set"`
}

type xmlSetWeaponStat struct {
	Stat string `xml:"stat,attr"`
	Val  string `xml:"val,attr"`
}

var templatesItemWeapon = map[int32]itemWeaponTemplate{}
var templatesItemWeaponMu sync.RWMutex

func carregarTemplatesItemWeapon(datapackPath string) error {
	itemsPath := filepath.Join(datapackPath, "data", "xml", "items")
	arquivos, err := filepath.Glob(filepath.Join(itemsPath, "*.xml"))
	if err != nil {
		return err
	}
	novoMapa := make(map[int32]itemWeaponTemplate)
	for _, arquivo := range arquivos {
		dados, errLeitura := os.ReadFile(arquivo)
		if errLeitura != nil {
			logger.Warnf("Falha ao ler XML de item %s: %v", arquivo, errLeitura)
			continue
		}
		var lista xmlListaItems
		errXml := xml.Unmarshal(dados, &lista)
		if errXml != nil {
			logger.Warnf("Falha ao parsear XML de item %s: %v", arquivo, errXml)
			continue
		}
		for _, item := range lista.Items {
			if item.Type != "Weapon" {
				continue
			}
			itemID := parseInt32Seguro(item.ID)
			if itemID <= 0 {
				continue
			}
			for _, bloco := range item.For {
				for _, set := range bloco.Sets {
					if set.Stat != "pAtkSpd" {
						continue
					}
					novoMapa[itemID] = itemWeaponTemplate{itemID: itemID, pAtkSpd: parseInt32Seguro(set.Val)}
				}
			}
		}
	}
	templatesItemWeaponMu.Lock()
	templatesItemWeapon = novoMapa
	templatesItemWeaponMu.Unlock()
	logger.Infof("Templates de armas carregados: %d", len(novoMapa))
	return nil
}

func obterPAtkSpdArma(itemID int32) (int32, bool) {
	templatesItemWeaponMu.RLock()
	template, ok := templatesItemWeapon[itemID]
	templatesItemWeaponMu.RUnlock()
	if !ok {
		return 0, false
	}
	return template.pAtkSpd, true
}
