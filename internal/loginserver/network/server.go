package network

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/controller"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type LoginServer struct {
	config             *config.LoginServerConfig
	loginController    *controller.LoginController
	listener           net.Listener
	gameServerListener *gameServerListener
	clients            sync.Map
	shutdown           chan struct{}
}

func NewLoginServer(cfg *config.LoginServerConfig, loginCtrl *controller.LoginController) *LoginServer {
	return &LoginServer{
		config:             cfg,
		loginController:    loginCtrl,
		gameServerListener: novoGameServerListener(cfg, loginCtrl),
		shutdown:           make(chan struct{}),
	}
}

func (ls *LoginServer) Start() error {
	addr := fmt.Sprintf("%s:%d", ls.config.Server.Host, ls.config.Server.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("erro ao iniciar listener: %w", err)
	}

	ls.listener = listener
	logger.Infof("LoginServer iniciado em %s", addr)
	logger.Infof("Aguardando conexoes...")
	if err := ls.gameServerListener.iniciar(); err != nil {
		return fmt.Errorf("erro ao iniciar listener de gameserver: %w", err)
	}

	go ls.acceptConnections()

	<-ls.shutdown
	return nil
}

func (ls *LoginServer) acceptConnections() {
	logger.Info("Thread de aceitacao de conexoes iniciada")
	for {
		logger.Debug("Aguardando nova conexao...")
		conn, err := ls.listener.Accept()
		if err != nil {
			select {
			case <-ls.shutdown:
				logger.Info("Shutdown detectado, encerrando acceptConnections")
				return
			default:
				logger.Errorf("Erro ao aceitar conexao: %v", err)
				continue
			}
		}

		logger.Infof("Nova conexao aceita de: %s", conn.RemoteAddr().String())
		go ls.handleClient(conn)
	}
}

func (ls *LoginServer) handleClient(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	logger.Debugf("Nova conexao de: %s", clientAddr)

	client := NewLoginClient(conn, ls.loginController)
	ls.clients.Store(clientAddr, client)
	defer ls.clients.Delete(clientAddr)

	client.Handle(context.Background())

	logger.Debugf("Conexao encerrada: %s", clientAddr)
}

func (ls *LoginServer) Stop() {
	logger.Info("Encerrando LoginServer...")
	close(ls.shutdown)

	if ls.listener != nil {
		ls.listener.Close()
	}
	if ls.gameServerListener != nil {
		ls.gameServerListener.parar()
	}

	ls.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*LoginClient); ok {
			client.Close()
		}
		return true
	})

	logger.Info("LoginServer encerrado")
}
