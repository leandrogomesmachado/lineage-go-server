package network

import "time"

type tipoNpcDesire string

const (
	tipoNpcDesireIdle     tipoNpcDesire = "idle"
	tipoNpcDesireAttack   tipoNpcDesire = "attack"
	tipoNpcDesireWander   tipoNpcDesire = "wander"
	tipoNpcDesireNothing  tipoNpcDesire = "nothing"
)

type npcDesire struct {
	tipo         tipoNpcDesire
	alvoObjID    int32
	peso         float64
	timerMs      int64
	moveToTarget bool
	timestampMs  int64
}

type npcDesireQueue struct {
	desires []npcDesire
}

func novoNpcDesireQueue() npcDesireQueue {
	return npcDesireQueue{desires: make([]npcDesire, 0, 8)}
}

func (f *npcDesireQueue) limpar() {
	if f == nil {
		return
	}
	f.desires = f.desires[:0]
}

func (f *npcDesireQueue) estaVazia() bool {
	if f == nil {
		return true
	}
	return len(f.desires) == 0
}

func (f *npcDesireQueue) adicionarOuAtualizar(desire npcDesire) {
	if f == nil {
		return
	}
	for indice := range f.desires {
		atual := &f.desires[indice]
		if atual.tipo != desire.tipo {
			continue
		}
		if atual.alvoObjID != desire.alvoObjID {
			continue
		}
		atual.peso += desire.peso
		if desire.timestampMs > atual.timestampMs {
			atual.timestampMs = desire.timestampMs
		}
		return
	}
	f.desires = append(f.desires, desire)
}

func (f *npcDesireQueue) removerPorTipoEAlvo(tipo tipoNpcDesire, alvoObjID int32) {
	if f == nil {
		return
	}
	if len(f.desires) == 0 {
		return
	}
	filtrados := f.desires[:0]
	for _, desire := range f.desires {
		if desire.tipo == tipo && desire.alvoObjID == alvoObjID {
			continue
		}
		filtrados = append(filtrados, desire)
	}
	f.desires = filtrados
}

func (f *npcDesireQueue) diminuirPesoPorTipo(tipo tipoNpcDesire, quantidade float64) {
	if f == nil {
		return
	}
	if len(f.desires) == 0 {
		return
	}
	filtrados := f.desires[:0]
	for _, desire := range f.desires {
		if desire.tipo != tipo {
			filtrados = append(filtrados, desire)
			continue
		}
		desire.peso -= quantidade
		if desire.peso > 0 {
			filtrados = append(filtrados, desire)
		}
	}
	f.desires = filtrados
}

func (f *npcDesireQueue) obterMaiorPeso() *npcDesire {
	if f == nil {
		return nil
	}
	if len(f.desires) == 0 {
		return nil
	}
	melhorIndice := 0
	for indice := 1; indice < len(f.desires); indice++ {
		if f.desires[indice].peso > f.desires[melhorIndice].peso {
			melhorIndice = indice
			continue
		}
		if f.desires[indice].peso == f.desires[melhorIndice].peso && f.desires[indice].timestampMs > f.desires[melhorIndice].timestampMs {
			melhorIndice = indice
		}
	}
	return &f.desires[melhorIndice]
}

func novoNpcDesireAttack(alvoObjID int32, peso float64, moveToTarget bool) npcDesire {
	agoraMs := time.Now().UnixMilli()
	return npcDesire{tipo: tipoNpcDesireAttack, alvoObjID: alvoObjID, peso: peso, moveToTarget: moveToTarget, timestampMs: agoraMs}
}
