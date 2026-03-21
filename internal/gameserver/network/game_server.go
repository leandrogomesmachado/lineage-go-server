package network

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type gameServer struct {
	config            *config.GameServerConfig
	repositorios      *gsdb.CharacterDataRepositories
	characterRepo     *gsdb.CharacterRepository
	loginServerClient *clienteLoginServer
	listener          net.Listener
	clientesPendentes sync.Map
	mundo             *mundoGameServer
	canalParada       chan struct{}
}

func NovoGameServer(cfg *config.GameServerConfig, repositorios *gsdb.CharacterDataRepositories) *gameServer {
	server := &gameServer{
		config:       cfg,
		repositorios: repositorios,
		mundo:        novoMundoGameServer(),
		canalParada:  make(chan struct{}),
	}
	if repositorios != nil {
		server.characterRepo = repositorios.Characters
	}
	server.loginServerClient = NovoClienteLoginServer(cfg, server)
	return server
}

func (g *gameServer) Iniciar() error {
	datapackResolvido, err := resolverDatapackPath(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	g.config.Datapack.Path = datapackResolvido
	carregarGeodata(g.config.Datapack.Path)
	err = carregarTemplatesItemWeapon(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTemplatesItemEquip(g.config.Datapack.Path)
	if err != nil {
		logger.Warnf("Falha ao carregar templates de equip de itens: %v", err)
	}
	err = carregarTemplatesPersonagemInicial(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTemplatesNpc(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTemplatesSpawnGlobal(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarMetadadosSkills(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTemplatesSkillCubic(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTemplatesSkillAbnormal(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTemplatesSkillFishing(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTemplatesClasseSkills(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTabelaExpNivel(g.config.Datapack.Path)
	if err != nil {
		return err
	}
	err = carregarTemplatesSkillsAtivas(g.config.Datapack.Path)
	if err != nil {
		logger.Warnf("Falha ao carregar templates de skills ativas: %v", err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", g.config.Server.Host, g.config.Server.Port))
	if err != nil {
		return err
	}
	g.inicializarNpcsGlobais()
	g.listener = listener
	g.loginServerClient.Iniciar()
	go g.loopMovimentoRuntime()
	go g.aceitarConexoes()
	logger.Infof("GameServer de clientes iniciado em %s:%d", g.config.Server.Host, g.config.Server.Port)
	return nil
}

func resolverDatapackPath(datapackPath string) (string, error) {
	candidatos := []string{}
	if datapackPath != "" {
		candidatos = append(candidatos, datapackPath)
	}
	candidatos = append(candidatos,
		"raptors-go_datapack",
		filepath.Join(".", "raptors-go_datapack"),
		filepath.Join("..", "raptors-go_datapack"),
	)
	for _, candidato := range candidatos {
		resolvido, err := filepath.Abs(candidato)
		if err != nil {
			continue
		}
		if datapackPathValido(resolvido) {
			logger.Infof("Datapack resolvido para %s", resolvido)
			return resolvido, nil
		}
	}
	return "", os.ErrNotExist
}

func datapackPathValido(datapackPath string) bool {
	if datapackPath == "" {
		return false
	}
	classesPath := filepath.Join(datapackPath, "data", "xml", "classes")
	if !diretorioExiste(classesPath) {
		return false
	}
	skillsPath := filepath.Join(datapackPath, "data", "xml", "skills")
	if !diretorioExiste(skillsPath) {
		return false
	}
	npcsPath := filepath.Join(datapackPath, "data", "xml", "npcs")
	if !diretorioExiste(npcsPath) {
		return false
	}
	spawnPath := filepath.Join(datapackPath, "data", "xml", "spawnlist")
	if !diretorioExiste(spawnPath) {
		return false
	}
	return true
}

func diretorioExiste(caminho string) bool {
	info, err := os.Stat(caminho)
	if err != nil {
		return false
	}
	return info.IsDir()
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
	if g.canalParada != nil {
		close(g.canalParada)
		g.canalParada = nil
	}
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
