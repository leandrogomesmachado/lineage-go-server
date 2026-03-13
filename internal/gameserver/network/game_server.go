package network

import (
	"context"
	"fmt"
	"net"
	"sync"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type gameServer struct {
	config            *config.GameServerConfig
	characterRepo     *gsdb.CharacterRepository
	loginServerClient *clienteLoginServer
	listener          net.Listener
	clientesPendentes sync.Map
	mundo             *mundoGameServer
}

func NovoGameServer(cfg *config.GameServerConfig, repo *gsdb.CharacterRepository) *gameServer {
	server := &gameServer{
		config:        cfg,
		characterRepo: repo,
		mundo:         novoMundoGameServer(),
	}
	server.loginServerClient = NovoClienteLoginServer(cfg, server)
	return server
}

func (g *gameServer) Iniciar() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", g.config.Server.Host, g.config.Server.Port))
	if err != nil {
		return err
	}
	g.listener = listener
	g.loginServerClient.Iniciar()
	go g.aceitarConexoes()
	logger.Infof("GameServer de clientes iniciado em %s:%d", g.config.Server.Host, g.config.Server.Port)
	return nil
}

func (g *gameServer) aceitarConexoes() {
	for {
		conn, err := g.listener.Accept()
		if err != nil {
			return
		}
		logger.Infof("Nova conexao do cliente de jogo: %s", conn.RemoteAddr().String())
		cliente := novoGameClient(conn, g)
		go cliente.loop()
	}
}

func (g *gameServer) Parar() {
	if g.listener == nil {
		return
	}
	g.listener.Close()
}

func (g *gameServer) registrarAutenticacaoPendente(cliente *gameClient) error {
	g.clientesPendentes.Store(cliente.conta, cliente)
	return g.loginServerClient.registrarClienteJogo(cliente)
}

func (g *gameServer) concluirAutenticacao(conta string, sucesso bool) error {
	valor, ok := g.clientesPendentes.Load(conta)
	if !ok {
		return nil
	}
	g.clientesPendentes.Delete(conta)
	cliente := valor.(*gameClient)
	if !sucesso {
		errFalha := cliente.autenticarComFalha()
		cliente.conn.Close()
		return errFalha
	}
	slots, err := g.characterRepo.FindByAccount(context.Background(), conta)
	if err != nil {
		return err
	}
	return cliente.autenticarComSucesso(slots)
}

func (g *gameServer) removerCliente(cliente *gameClient) {
	if cliente == nil {
		return
	}
	if cliente.playerAtivo != nil {
		g.mundo.remover(cliente.playerAtivo.objID)
	}
	if cliente.conta == "" {
		return
	}
	g.clientesPendentes.Delete(cliente.conta)
}
