package network

import "sync"

type gameServerInfo struct {
	id            int
	nome          string
	host          string
	porta         int
	maxPlayers    int
	authed        bool
	thread        *gameServerThread
	contasEmJogo  sync.Map
}

func (g *gameServerInfo) setAuthed(authed bool) {
	g.authed = authed
}

func (g *gameServerInfo) isAuthed() bool {
	return g.authed
}

func (g *gameServerInfo) adicionarConta(conta string) {
	g.contasEmJogo.Store(conta, true)
}

func (g *gameServerInfo) removerConta(conta string) {
	g.contasEmJogo.Delete(conta)
}

func (g *gameServerInfo) possuiConta(conta string) bool {
	_, ok := g.contasEmJogo.Load(conta)
	return ok
}
