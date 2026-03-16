package network

import "strings"

func (n *npcGlobalRuntime) ehScriptAiBase(nomeBase string) bool {
	if n == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(n.scriptAiBase), strings.TrimSpace(nomeBase)) {
		return true
	}
	descritorScript := strings.ToLower(strings.TrimSpace(n.scriptAiDescritor))
	nomeBaseNormalizado := strings.ToLower(strings.TrimSpace(nomeBase))
	if nomeBaseNormalizado == "" {
		return false
	}
	if strings.Contains(descritorScript, "/"+nomeBaseNormalizado) {
		return true
	}
	return false
}

func (n *npcGlobalRuntime) ehScriptAiVariante(nomeVariante string) bool {
	if n == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(n.scriptAiVariante), strings.TrimSpace(nomeVariante)) {
		return true
	}
	descritorScript := strings.ToLower(strings.TrimSpace(n.scriptAiDescritor))
	nomeVarianteNormalizado := strings.ToLower(strings.TrimSpace(nomeVariante))
	if nomeVarianteNormalizado == "" {
		return false
	}
	if strings.HasSuffix(descritorScript, "/"+nomeVarianteNormalizado) {
		return true
	}
	return false
}
