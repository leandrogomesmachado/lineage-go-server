package network

type npcHateList struct {
	valores map[int32]float64
}

func novoNpcHateList() npcHateList {
	return npcHateList{valores: make(map[int32]float64)}
}

func (h *npcHateList) limpar() {
	if h == nil {
		return
	}
	h.valores = make(map[int32]float64)
}

func (h *npcHateList) adicionar(alvoObjID int32, valor float64) {
	if h == nil {
		return
	}
	if alvoObjID <= 0 {
		return
	}
	if valor <= 0 {
		return
	}
	if h.valores == nil {
		h.valores = make(map[int32]float64)
	}
	h.valores[alvoObjID] = h.valores[alvoObjID] + valor
}

func (h *npcHateList) obter(alvoObjID int32) float64 {
	if h == nil {
		return 0
	}
	if h.valores == nil {
		return 0
	}
	return h.valores[alvoObjID]
}

func (h *npcHateList) remover(alvoObjID int32) {
	if h == nil {
		return
	}
	if h.valores == nil {
		return
	}
	delete(h.valores, alvoObjID)
}

func (h *npcHateList) reduzirTodos(quantidade float64) {
	if h == nil {
		return
	}
	if h.valores == nil {
		return
	}
	for alvoObjID, valor := range h.valores {
		novoValor := valor - quantidade
		if novoValor > 0 {
			h.valores[alvoObjID] = novoValor
			continue
		}
		delete(h.valores, alvoObjID)
	}
}

func (h *npcHateList) obterMaisOdiado() int32 {
	if h == nil {
		return 0
	}
	if len(h.valores) == 0 {
		return 0
	}
	melhorObjID := int32(0)
	maiorValor := 0.0
	for alvoObjID, valor := range h.valores {
		if melhorObjID == 0 {
			melhorObjID = alvoObjID
			maiorValor = valor
			continue
		}
		if valor > maiorValor {
			melhorObjID = alvoObjID
			maiorValor = valor
		}
	}
	return melhorObjID
}
