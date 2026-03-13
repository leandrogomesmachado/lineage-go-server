package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/controller"
	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/network"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	pkgdb "github.com/leandrogomesmachado/l2raptors-go/pkg/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

func main() {
	configPath := flag.String("config", "configs/loginserver.yaml", "Caminho do arquivo de configuracao")
	flag.Parse()

	cfg, err := config.LoadLoginServerConfig(*configPath)
	if err != nil {
		panic("Erro ao carregar configuracao: " + err.Error())
	}

	if err := logger.Init(cfg.Logging.Level, cfg.Logging.File, cfg.Logging.Console); err != nil {
		panic("Erro ao inicializar logger: " + err.Error())
	}

	logger.Info("========================================")
	logger.Info("  L2Raptors LoginServer - Go Edition")
	logger.Info("========================================")
	logger.Info("")

	logger.Info("Conectando ao MongoDB...")
	db, err := pkgdb.NewMongoDB(cfg.Database.URI, cfg.Database.Database, cfg.Database.Timeout)
	if err != nil {
		logger.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer db.Close()
	logger.Info("MongoDB conectado com sucesso")

	accountRepo := database.NewAccountRepository(db.Database)

	loginController := controller.GetInstance()
	loginController.SetAccountRepository(accountRepo)

	server := network.NewLoginServer(cfg, loginController)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Sinal de interrupcao recebido")
		server.Stop()
	}()

	logger.Infof("Iniciando LoginServer na porta %d...", cfg.Server.Port)
	if err := server.Start(); err != nil {
		logger.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
