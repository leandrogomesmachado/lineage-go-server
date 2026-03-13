package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/network"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	pkgdb "github.com/leandrogomesmachado/l2raptors-go/pkg/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func main() {
	configPath := flag.String("config", "configs/gameserver.yaml", "Caminho do arquivo de configuracao")
	flag.Parse()

	cfg, err := config.LoadGameServerConfig(*configPath)
	if err != nil {
		panic("Erro ao carregar configuracao: " + err.Error())
	}

	if err := logger.Init(cfg.Logging.Level, cfg.Logging.File, cfg.Logging.Console); err != nil {
		panic("Erro ao inicializar logger: " + err.Error())
	}

	logger.Info("========================================")
	logger.Info("  L2Raptors GameServer - Go Edition")
	logger.Info("========================================")
	logger.Info("")

	logger.Infof("Server ID: %d", cfg.Server.ID)
	logger.Infof("Server Name: %s", cfg.Server.Name)
	logger.Infof("Max Players: %d", cfg.Server.MaxPlayers)
	logger.Info("")

	logger.Info("Conectando ao MongoDB...")
	db, err := pkgdb.NewMongoDB(cfg.Database.URI, cfg.Database.Database, cfg.Database.Timeout)
	if err != nil {
		logger.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer db.Close()
	logger.Info("MongoDB conectado com sucesso")

	logger.Infof("Datapack path: %s", cfg.Datapack.Path)
	logger.Infof("Geodata enabled: %v", cfg.Geodata.Enabled)
	if cfg.Geodata.Enabled {
		logger.Infof("Geodata path: %s", cfg.Geodata.Path)
	}

	logger.Info("")
	logger.Info("Rates configuradas:")
	logger.Infof("  XP: %.1fx", cfg.Rates.XP)
	logger.Infof("  SP: %.1fx", cfg.Rates.SP)
	logger.Infof("  Adena: %.1fx", cfg.Rates.Adena)
	logger.Infof("  Drop: %.1fx", cfg.Rates.Drop)
	logger.Infof("  Spoil: %.1fx", cfg.Rates.Spoil)

	logger.Info("")
	logger.Infof("Iniciando GameServer na porta %d...", cfg.Server.Port)
	characterRepo := gsdb.NewCharacterRepository(db.Database)
	serverJogo := network.NovoGameServer(cfg, characterRepo)
	if err = serverJogo.Iniciar(); err != nil {
		logger.Fatalf("Erro ao iniciar GameServer de clientes: %v", err)
	}
	logger.Info("GameServer pronto para conexoes!")
	logger.Info("Pressione Ctrl+C para encerrar")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Encerrando GameServer...")
	logger.Info("GameServer encerrado")
}
