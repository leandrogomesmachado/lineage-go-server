package network

import "sync"

type mundoGameServer struct {
	mutex   sync.RWMutex
	players map[int32]*gameClient
	npcs    map[int32]*npcGlobalRuntime
}

func novoMundoGameServer() *mundoGameServer {
	return &mundoGameServer{players: make(map[int32]*gameClient), npcs: make(map[int32]*npcGlobalRuntime)}
}

func (m *mundoGameServer) registrar(cliente *gameClient) {
	if cliente == nil {
		return
	}
	if cliente.playerAtivo == nil {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.players[cliente.playerAtivo.objID] = cliente
}

func (m *mundoGameServer) remover(objID int32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.players, objID)
}

func (m *mundoGameServer) registrarNpc(npc *npcGlobalRuntime) {
	if npc == nil {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.npcs[npc.objID] = npc
}

func (m *mundoGameServer) limparNpcs() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.npcs = make(map[int32]*npcGlobalRuntime)
}

func (m *mundoGameServer) removerNpc(objID int32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.npcs, objID)
}

func (m *mundoGameServer) listarOutros(objID int32) []*gameClient {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	resultado := make([]*gameClient, 0, len(m.players))
	for id, cliente := range m.players {
		if id == objID {
			continue
		}
		resultado = append(resultado, cliente)
	}
	return resultado
}

func (m *mundoGameServer) listarPlayersVisiveisParaNpc(npc *npcGlobalRuntime) []*gameClient {
	if npc == nil {
		return nil
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	resultado := make([]*gameClient, 0, len(m.players))
	for _, cliente := range m.players {
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		if !posicaoNpcNoRaioVisivel(cliente.playerAtivo, npc) {
			continue
		}
		resultado = append(resultado, cliente)
	}
	return resultado
}

func (m *mundoGameServer) listarNpcsGlobais() []*npcGlobalRuntime {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	resultado := make([]*npcGlobalRuntime, 0, len(m.npcs))
	for _, npc := range m.npcs {
		if npc == nil {
			continue
		}
		resultado = append(resultado, npc)
	}
	return resultado
}

func (m *mundoGameServer) listarPlayersAtivos() []*gameClient {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	resultado := make([]*gameClient, 0, len(m.players))
	for _, cliente := range m.players {
		if cliente == nil {
			continue
		}
		if cliente.playerAtivo == nil {
			continue
		}
		resultado = append(resultado, cliente)
	}
	return resultado
}

func (m *mundoGameServer) listarNpcsVisiveisPara(cliente *gameClient) []*npcGlobalRuntime {
	if cliente == nil {
		return nil
	}
	if cliente.playerAtivo == nil {
		return nil
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	resultado := make([]*npcGlobalRuntime, 0, len(m.npcs))
	for _, npc := range m.npcs {
		if npc == nil {
			continue
		}
		if !posicaoNpcNoRaioVisivel(cliente.playerAtivo, npc) {
			continue
		}
		resultado = append(resultado, npc)
	}
	return resultado
}

func (m *mundoGameServer) obterPorObjID(objID int32) *gameClient {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.players[objID]
}

func (m *mundoGameServer) obterNpcPorObjID(objID int32) *npcGlobalRuntime {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.npcs[objID]
}

func (m *mundoGameServer) listarVisiveisPara(cliente *gameClient) []*gameClient {
	if cliente == nil {
		return nil
	}
	if cliente.playerAtivo == nil {
		return nil
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	resultado := make([]*gameClient, 0, len(m.players))
	for id, outroCliente := range m.players {
		if id == cliente.playerAtivo.objID {
			continue
		}
		if outroCliente == nil {
			continue
		}
		if outroCliente.playerAtivo == nil {
			continue
		}
		if !posicaoNoRaioVisivel(cliente.playerAtivo, outroCliente.playerAtivo) {
			continue
		}
		resultado = append(resultado, outroCliente)
	}
	return resultado
}
