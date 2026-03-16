package network

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type npcScriptAiInfo struct {
	descritor string
	base      string
	variante  string
	arquivo   string
}

var npcScriptsAiPorNpcID = map[int32]npcScriptAiInfo{}
var npcScriptsAiMu sync.RWMutex
var npcScriptsAiCarregados bool

var regexNpcIdsJava = regexp.MustCompile(`(?s)protected\s+final\s+int\[\]\s+_npcIds\s*=\s*\{(.*?)\};`)
var regexNumeroJava = regexp.MustCompile(`\d+`)
var regexSuperDescritorJava = regexp.MustCompile(`super\("([^"]+)"\)`)

func carregarMapaScriptsAiMonstrosJava() {
	npcScriptsAiMu.Lock()
	defer npcScriptsAiMu.Unlock()
	if npcScriptsAiCarregados {
		return
	}
	npcScriptsAiCarregados = true
	caminhoBase := resolverCaminhoScriptsAiMonstrosJava()
	if caminhoBase == "" {
		return
	}
	infos := make(map[int32]npcScriptAiInfo)
	_ = filepath.Walk(caminhoBase, func(caminho string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".java") {
			return nil
		}
		conteudoBytes, errLeitura := os.ReadFile(caminho)
		if errLeitura != nil {
			return nil
		}
		conteudo := string(conteudoBytes)
		if !strings.Contains(conteudo, "_npcIds") {
			return nil
		}
		descritor := extrairDescritorScriptAiJava(conteudo)
		if descritor == "" {
			return nil
		}
		npcIDs := extrairNpcIDsScriptAiJava(conteudo)
		if len(npcIDs) == 0 {
			return nil
		}
		base, variante := extrairBaseEVarianteScriptAi(descritor)
		for _, npcID := range npcIDs {
			infos[npcID] = npcScriptAiInfo{descritor: descritor, base: base, variante: variante, arquivo: caminho}
		}
		return nil
	})
	npcScriptsAiPorNpcID = infos
	logger.Infof("Scripts AI de monstros carregados do Java: npcs=%d base=%s", len(infos), caminhoBase)
}

func resolverCaminhoScriptsAiMonstrosJava() string {
	candidatos := []string{
		filepath.Join("c:\\dev\\l2raptors\\l2raptors-java", "raptors_gameserver", "java", "net", "sf", "l2j", "gameserver", "scripting", "script", "ai", "individual", "Monster"),
		filepath.Join("..", "l2raptors-java", "raptors_gameserver", "java", "net", "sf", "l2j", "gameserver", "scripting", "script", "ai", "individual", "Monster"),
	}
	for _, candidato := range candidatos {
		info, err := os.Stat(candidato)
		if err != nil {
			continue
		}
		if info == nil {
			continue
		}
		if !info.IsDir() {
			continue
		}
		return candidato
	}
	return ""
}

func extrairBaseEVarianteScriptAi(descritor string) (string, string) {
	descritorLimpo := strings.TrimSpace(descritor)
	if descritorLimpo == "" {
		return "", ""
	}
	partes := strings.Split(descritorLimpo, "/")
	if len(partes) == 0 {
		return "", ""
	}
	if len(partes) == 1 {
		return partes[0], ""
	}
	ultimaParte := strings.TrimSpace(partes[len(partes)-1])
	penultimaParte := strings.TrimSpace(partes[len(partes)-2])
	if len(partes) >= 2 {
		return penultimaParte, ultimaParte
	}
	return ultimaParte, ""
}

func extrairDescritorScriptAiJava(conteudo string) string {
	matches := regexSuperDescritorJava.FindAllStringSubmatch(conteudo, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		descritor := strings.TrimSpace(match[1])
		if descritor == "" {
			continue
		}
		if !strings.Contains(strings.ToLower(descritor), "monster") {
			continue
		}
		return descritor
	}
	return ""
}

func extrairNpcIDsScriptAiJava(conteudo string) []int32 {
	match := regexNpcIdsJava.FindStringSubmatch(conteudo)
	if len(match) < 2 {
		return nil
	}
	numeros := regexNumeroJava.FindAllString(match[1], -1)
	resultado := make([]int32, 0, len(numeros))
	for _, numero := range numeros {
		npcID := parseInt32Seguro(numero)
		if npcID <= 0 {
			continue
		}
		resultado = append(resultado, npcID)
	}
	return resultado
}

func obterScriptAiMonsterPorNpcID(npcID int32) npcScriptAiInfo {
	carregarMapaScriptsAiMonstrosJava()
	npcScriptsAiMu.RLock()
	info := npcScriptsAiPorNpcID[npcID]
	npcScriptsAiMu.RUnlock()
	return info
}

func possuiScriptAiMonsterCarregado(npcID int32) bool {
	if npcID <= 0 {
		return false
	}
	info := obterScriptAiMonsterPorNpcID(npcID)
	if strings.TrimSpace(info.descritor) == "" {
		return false
	}
	return true
}
