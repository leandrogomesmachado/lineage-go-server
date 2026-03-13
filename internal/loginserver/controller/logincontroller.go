package controller

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/crypto"
	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/protocol"
	"golang.org/x/crypto/bcrypt"
)

const (
	LoginTimeout      = 60 * 1000
	RSAKeyPairCount   = 10
	BlowfishKeyCount  = 20
	BlowfishKeyLength = 16
)

type LoginController struct {
	rsaKeyPairs    []*crypto.RSAKeyPair
	blowfishKeys   [][]byte
	clients        sync.Map
	failedAttempts sync.Map
	accountRepo    *database.AccountRepository

	autoCreateAccounts bool
	loginTryBeforeBan  int
	loginBlockAfterBan int
	showLicence        bool
}

type sessaoAutenticada struct {
	sessionKey      *protocol.SessionKey
	joinedGS        bool
	connectionStart time.Time
}

var (
	instance *LoginController
	once     sync.Once
)

func GetInstance() *LoginController {
	once.Do(func() {
		instance = &LoginController{
			rsaKeyPairs:        make([]*crypto.RSAKeyPair, RSAKeyPairCount),
			blowfishKeys:       make([][]byte, BlowfishKeyCount),
			autoCreateAccounts: true,
			loginTryBeforeBan:  3,
			loginBlockAfterBan: 600,
			showLicence:        false,
		}
		instance.initialize()
	})
	return instance
}

func (lc *LoginController) SetAccountRepository(repo *database.AccountRepository) {
	lc.accountRepo = repo
	logger.Infof("LoginController %p recebeu AccountRepository", lc)
}

func (lc *LoginController) initialize() {
	logger.Info("Inicializando LoginController...")

	for i := 0; i < RSAKeyPairCount; i++ {
		keyPair, err := crypto.GenerateRSAKeyPair()
		if err != nil {
			logger.Errorf("Erro ao gerar RSA KeyPair %d: %v", i, err)
			continue
		}
		lc.rsaKeyPairs[i] = keyPair
	}
	logger.Infof("Cached %d KeyPairs para comunicacao RSA", RSAKeyPairCount)

	for i := 0; i < BlowfishKeyCount; i++ {
		lc.blowfishKeys[i] = make([]byte, BlowfishKeyLength)
		for j := 0; j < BlowfishKeyLength; j++ {
			n, _ := rand.Int(rand.Reader, big.NewInt(255))
			lc.blowfishKeys[i][j] = byte(n.Int64() + 1)
		}
	}
	logger.Infof("Stored %d keys para comunicacao Blowfish", BlowfishKeyCount)

	go lc.purgeThread()
}

func (lc *LoginController) GetRandomBlowfishKey() []byte {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(BlowfishKeyCount)))
	return lc.blowfishKeys[n.Int64()]
}

func (lc *LoginController) GetScrambledRSAKeyPair() *crypto.RSAKeyPair {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(RSAKeyPairCount)))
	return lc.rsaKeyPairs[n.Int64()]
}

func (lc *LoginController) recordFailedAttempt(addr net.Addr) {
	ip := addr.String()

	var attempts int
	if val, ok := lc.failedAttempts.Load(ip); ok {
		attempts = val.(int) + 1
	} else {
		attempts = 1
	}

	lc.failedAttempts.Store(ip, attempts)

	if attempts >= lc.loginTryBeforeBan {
		lc.failedAttempts.Delete(ip)
		logger.Infof("IP address: %s foi banido devido a muitas tentativas de login", ip)
	}
}

func (lc *LoginController) RetrieveAccountInfo(ctx context.Context, clientAddr net.Addr, login, password string) (*database.Account, *protocol.SessionKey, error) {
	currentTime := time.Now()

	account, err := lc.accountRepo.FindByLogin(ctx, login)
	if err != nil {
		logger.Errorf("Erro ao buscar conta: %v", err)
		return nil, nil, err
	}

	if account == nil {
		if !lc.autoCreateAccounts {
			lc.recordFailedAttempt(clientAddr)
			return nil, nil, protocol.ErrUserOrPassWrong
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Errorf("Erro ao gerar hash de senha: %v", err)
			return nil, nil, protocol.ErrAccessFailed
		}

		account = &database.Account{
			Login:       login,
			Password:    string(hashedPassword),
			AccessLevel: 0,
			LastActive:  currentTime,
			LastServer:  1,
			LastIP:      clientAddr.String(),
		}

		if err := lc.accountRepo.Create(ctx, account); err != nil {
			if errors.Is(err, database.ErrContaJaExiste) {
				logger.Warnf("Tentativa de criar conta duplicada para login %s", login)
				return nil, nil, protocol.ErrAccessFailed
			}
			logger.Errorf("Erro ao criar conta: %v", err)
			return nil, nil, protocol.ErrAccessFailed
		}

		logger.Infof("Auto created account '%s'", login)
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
			lc.recordFailedAttempt(clientAddr)
			return nil, nil, protocol.ErrPassWrong
		}

		lc.failedAttempts.Delete(clientAddr.String())

		if err := lc.accountRepo.UpdateLastLogin(ctx, login, clientAddr.String(), 1); err != nil {
			logger.Errorf("Erro ao atualizar last login: %v", err)
			return nil, nil, protocol.ErrAccessFailed
		}
	}

	if account.AccessLevel < 0 {
		return nil, nil, protocol.ErrPermanentlyBanned
	}

	if valorExistente, exists := lc.clients.Load(login); exists {
		sessaoExistente := valorExistente.(*sessaoAutenticada)
		if sessaoExistente.connectionStart.Add(time.Duration(LoginTimeout) * time.Millisecond).Before(time.Now()) {
			lc.clients.Delete(login)
		}
		if _, aindaExiste := lc.clients.Load(login); aindaExiste {
			return nil, nil, protocol.ErrAccountInUse
		}
	}

	sessionKey := lc.generateSessionKey()
	logger.Infof("LoginController %p armazenando sessao autenticada para conta=%q len=%d: play=%d/%d login=%d/%d", lc, login, len(login), sessionKey.PlayOkID1, sessionKey.PlayOkID2, sessionKey.LoginOkID1, sessionKey.LoginOkID2)
	lc.clients.Store(login, &sessaoAutenticada{
		sessionKey:      sessionKey,
		joinedGS:        false,
		connectionStart: time.Now(),
	})

	return account, sessionKey, nil
}

func (lc *LoginController) generateSessionKey() *protocol.SessionKey {
	n1, _ := rand.Int(rand.Reader, big.NewInt(0x7FFFFFFF))
	n2, _ := rand.Int(rand.Reader, big.NewInt(0x7FFFFFFF))
	n3, _ := rand.Int(rand.Reader, big.NewInt(0x7FFFFFFF))
	n4, _ := rand.Int(rand.Reader, big.NewInt(0x7FFFFFFF))

	return &protocol.SessionKey{
		PlayOkID1:  uint32(n1.Int64()),
		PlayOkID2:  uint32(n2.Int64()),
		LoginOkID1: uint32(n3.Int64()),
		LoginOkID2: uint32(n4.Int64()),
	}
}

func (lc *LoginController) RemoveAuthedLoginClient(login string) {
	if login == "" {
		return
	}
	logger.Infof("LoginController %p removendo sessao autenticada de conta=%q len=%d", lc, login, len(login))
	lc.clients.Delete(login)
}

func (lc *LoginController) GetKeyForAccount(login string) *protocol.SessionKey {
	if val, ok := lc.clients.Load(login); ok {
		sessao := val.(*sessaoAutenticada)
		logger.Infof("LoginController %p encontrou sessao para conta=%q len=%d: play=%d/%d login=%d/%d joinedGS=%t", lc, login, len(login), sessao.sessionKey.PlayOkID1, sessao.sessionKey.PlayOkID2, sessao.sessionKey.LoginOkID1, sessao.sessionKey.LoginOkID2, sessao.joinedGS)
		return sessao.sessionKey
	}
	logger.Warnf("LoginController %p nao encontrou sessao para conta=%q len=%d. Contas armazenadas: %s", lc, login, len(login), lc.resumirContasAutenticadas())
	return nil
}

func (lc *LoginController) MarkJoinedGameServer(login string) {
	if login == "" {
		return
	}
	if val, ok := lc.clients.Load(login); ok {
		sessao := val.(*sessaoAutenticada)
		sessao.joinedGS = true
		logger.Infof("LoginController %p marcou conta=%q len=%d como joinedGS", lc, login, len(login))
		return
	}
	logger.Warnf("LoginController %p nao conseguiu marcar conta=%q len=%d como joinedGS porque a sessao nao existe. Contas armazenadas: %s", lc, login, len(login), lc.resumirContasAutenticadas())
}

func (lc *LoginController) resumirContasAutenticadas() string {
	contas := ""
	lc.clients.Range(func(chave, valor interface{}) bool {
		login := chave.(string)
		if contas == "" {
			contas = fmt.Sprintf("%q(len=%d)", login, len(login))
			return true
		}
		contas += ", " + fmt.Sprintf("%q(len=%d)", login, len(login))
		return true
	})
	if contas != "" {
		return contas
	}
	return "nenhuma"
}

func (lc *LoginController) ShowLicence() bool {
	return lc.showLicence
}

func (lc *LoginController) purgeThread() {
	ticker := time.NewTicker(time.Duration(LoginTimeout/2) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		agora := time.Now()
		lc.clients.Range(func(chave, valor interface{}) bool {
			login := chave.(string)
			sessao := valor.(*sessaoAutenticada)
			if sessao.joinedGS && sessao.connectionStart.Add(time.Duration(LoginTimeout)*time.Millisecond).After(agora) {
				return true
			}
			if sessao.connectionStart.Add(time.Duration(LoginTimeout) * time.Millisecond).After(agora) {
				return true
			}
			lc.clients.Delete(login)
			return true
		})
	}
}
