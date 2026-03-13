package network

import "sync"

type mundoGameServer struct {
	mutex   sync.RWMutex
	players map[int32]*gameClient
}

func novoMundoGameServer() *mundoGameServer {
	return &mundoGameServer{players: make(map[int32]*gameClient)}
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
