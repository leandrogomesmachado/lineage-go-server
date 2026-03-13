package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/controller"
	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/crypto"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type gameServerListener struct {
	config          *config.LoginServerConfig
	loginController *controller.LoginController
	listener        net.Listener
	gameServers     sync.Map
}

func novoGameServerListener(cfg *config.LoginServerConfig, loginCtrl *controller.LoginController) *gameServerListener {
	return &gameServerListener{
		config:          cfg,
		loginController: loginCtrl,
	}
}

func (g *gameServerListener) iniciar() error {
	endereco := fmt.Sprintf("%s:%d", g.config.LoginServer.Host, g.config.LoginServer.Port)
	listener, err := net.Listen("tcp", endereco)
	if err != nil {
		return err
	}
	g.listener = listener
	logger.Infof("GameServerListener iniciado em %s", endereco)
	go g.aceitarConexoes()
	return nil
}

func (g *gameServerListener) aceitarConexoes() {
	for {
		conn, err := g.listener.Accept()
		if err != nil {
			return
		}
		thread, erroThread := novoGameServerThread(conn, g.config, g.loginController)
		if erroThread != nil {
			logger.Errorf("Erro ao criar GameServerThread: %v", erroThread)
			conn.Close()
			continue
		}
		g.gameServers.Store(conn.RemoteAddr().String(), thread)
	}
}

func (g *gameServerListener) parar() {
	if g.listener == nil {
		return
	}
	g.listener.Close()
}

func novaCriptografiaGameServer() (*crypto.NewCrypt, error) {
	return crypto.NewNewCrypt([]byte("_;v.]05-31!|+-%xT!^["))
}
