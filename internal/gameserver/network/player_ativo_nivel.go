package network

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

var (
	tabelaExpNivel   []int64
	tabelaExpNivelMu sync.RWMutex
)

type xmlPlayerLevel struct {
	Level                int32 `xml:"level,attr"`
	RequiredExpToLevelUp int64 `xml:"requiredExpToLevelUp,attr"`
}

type xmlPlayerLevels struct {
	Levels []xmlPlayerLevel `xml:"playerLevel"`
}

func carregarTabelaExpNivel(datapackPath string) error {
	path := filepath.Join(datapackPath, "data", "xml", "playerLevels.xml")
	dados, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var lista xmlPlayerLevels
	if err := xml.Unmarshal(dados, &lista); err != nil {
		return err
	}
	maxLevel := int32(0)
	for _, l := range lista.Levels {
		if l.Level > maxLevel {
			maxLevel = l.Level
		}
	}
	tabela := make([]int64, maxLevel+2)
	for _, l := range lista.Levels {
		if int(l.Level) < len(tabela) {
			tabela[l.Level] = l.RequiredExpToLevelUp
		}
	}
	tabelaExpNivelMu.Lock()
	tabelaExpNivel = tabela
	tabelaExpNivelMu.Unlock()
	logger.Infof("Tabela de EXP por nivel carregada: %d niveis", maxLevel)
	return nil
}

func calcularNivelPorExp(exp int64) int32 {
	tabelaExpNivelMu.RLock()
	tabela := tabelaExpNivel
	tabelaExpNivelMu.RUnlock()
	if len(tabela) == 0 {
		return 1
	}
	nivelAtual := int32(1)
	for nivel := int32(1); nivel < int32(len(tabela)); nivel++ {
		if tabela[nivel] > exp {
			break
		}
		nivelAtual = nivel
	}
	return nivelAtual
}

func expParaProximoNivel(nivel int32) int64 {
	tabelaExpNivelMu.RLock()
	tabela := tabelaExpNivel
	tabelaExpNivelMu.RUnlock()
	proximo := nivel + 1
	if int(proximo) >= len(tabela) {
		return int64(^uint64(0) >> 1)
	}
	return tabela[proximo]
}

func nivelMaximoTabela() int32 {
	tabelaExpNivelMu.RLock()
	tabela := tabelaExpNivel
	tabelaExpNivelMu.RUnlock()
	return int32(len(tabela) - 2)
}
