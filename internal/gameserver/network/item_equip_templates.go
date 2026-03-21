package network

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type itemEquipTemplate struct {
	itemID   int32
	bodypart string
}

type xmlItemEquipSet struct {
	Name string `xml:"name,attr"`
	Val  string `xml:"val,attr"`
}

type xmlItemEquipEntry struct {
	ID   string            `xml:"id,attr"`
	Sets []xmlItemEquipSet `xml:"set"`
}

type xmlListaItemsEquip struct {
	Items []xmlItemEquipEntry `xml:"item"`
}

var templatesItemEquip = map[int32]itemEquipTemplate{}
var templatesItemEquipMu sync.RWMutex

func carregarTemplatesItemEquip(datapackPath string) error {
	itemsPath := filepath.Join(datapackPath, "data", "xml", "items")
	arquivos, err := filepath.Glob(filepath.Join(itemsPath, "*.xml"))
	if err != nil {
		return err
	}
	novoMapa := make(map[int32]itemEquipTemplate)
	for _, arquivo := range arquivos {
		dados, errLeitura := os.ReadFile(arquivo)
		if errLeitura != nil {
			logger.Warnf("Falha ao ler XML de item equip %s: %v", arquivo, errLeitura)
			continue
		}
		var lista xmlListaItemsEquip
		errXml := xml.Unmarshal(dados, &lista)
		if errXml != nil {
			logger.Warnf("Falha ao parsear XML de item equip %s: %v", arquivo, errXml)
			continue
		}
		for _, entry := range lista.Items {
			itemID := parseInt32Seguro(entry.ID)
			if itemID <= 0 {
				continue
			}
			for _, set := range entry.Sets {
				if set.Name != "bodypart" {
					continue
				}
				novoMapa[itemID] = itemEquipTemplate{itemID: itemID, bodypart: set.Val}
				break
			}
		}
	}
	templatesItemEquipMu.Lock()
	templatesItemEquip = novoMapa
	templatesItemEquipMu.Unlock()
	logger.Infof("Templates de equip de itens carregados: %d", len(novoMapa))
	return nil
}

func resolverSlotsEquipamento(itemID int32) []int32 {
	templatesItemEquipMu.RLock()
	template, ok := templatesItemEquip[itemID]
	templatesItemEquipMu.RUnlock()
	if !ok {
		return nil
	}
	return bodypartParaSlots(template.bodypart)
}

func bodypartParaSlots(bodypart string) []int32 {
	switch bodypart {
	case "rhand":
		return []int32{7}
	case "lhand":
		return []int32{8}
	case "lrhand":
		return []int32{7}
	case "head":
		return []int32{6}
	case "chest":
		return []int32{10}
	case "legs":
		return []int32{11}
	case "gloves":
		return []int32{9}
	case "feet":
		return []int32{12}
	case "neck":
		return []int32{3}
	case "rear;lear":
		return []int32{2, 1}
	case "rfinger;lfinger":
		return []int32{5, 4}
	case "underwear":
		return []int32{15}
	case "fullarmor":
		return []int32{10}
	case "hair":
		return []int32{15}
	case "hairall":
		return []int32{15}
	case "face":
		return []int32{16}
	case "back":
		return []int32{13}
	}
	return nil
}
