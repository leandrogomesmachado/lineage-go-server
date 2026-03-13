package auth

import (
	"context"
	"errors"
	"time"

	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/crypto"
	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

var (
	ErrInvalidCredentials = errors.New("credenciais invalidas")
	ErrAccountBanned      = errors.New("conta banida")
	ErrAccountExists      = errors.New("conta ja existe")
)

type AuthManager struct {
	repo     *database.AccountRepository
	config   *config.SecurityConfig
	attempts map[string]int
}

func NewAuthManager(repo *database.AccountRepository, cfg *config.SecurityConfig) *AuthManager {
	return &AuthManager{
		repo:     repo,
		config:   cfg,
		attempts: make(map[string]int),
	}
}

func (am *AuthManager) Authenticate(ctx context.Context, login, password, ip string) (*database.Account, error) {
	banned, err := am.repo.IsBanned(ctx, login)
	if err == nil && banned {
		return nil, ErrAccountBanned
	}

	account, err := am.repo.FindByLogin(ctx, login)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if account == nil {
		if !am.config.AutoCreateAccounts {
			return nil, ErrInvalidCredentials
		}
		logger.Infof("Criando nova conta: %s", login)
		return am.createAccount(ctx, login, password)
	}

	if !crypto.VerifyPassword(password, account.Password) {
		am.recordFailedAttempt(login)
		return nil, ErrInvalidCredentials
	}

	am.clearFailedAttempts(login)

	if err := am.repo.UpdateLastLogin(ctx, login, ip, 0); err != nil {
		logger.Warnf("Erro ao atualizar ultimo login: %v", err)
	}

	return account, nil
}

func (am *AuthManager) createAccount(ctx context.Context, login, password string) (*database.Account, error) {
	account := &database.Account{
		Login:       login,
		Password:    crypto.HashPassword(password),
		AccessLevel: 0,
		LastActive:  time.Now(),
	}

	if err := am.repo.Create(ctx, account); err != nil {
		if errors.Is(err, database.ErrContaJaExiste) {
			return nil, ErrAccountExists
		}
		return nil, err
	}

	logger.Infof("Conta criada com sucesso: %s", login)
	return account, nil
}

func (am *AuthManager) recordFailedAttempt(login string) {
	am.attempts[login]++

	if am.attempts[login] >= am.config.MaxLoginAttempts {
		ctx := context.Background()
		duration := time.Duration(am.config.BanDurationMinutes) * time.Minute

		if err := am.repo.BanAccount(ctx, login, duration); err != nil {
			logger.Errorf("Erro ao banir conta %s: %v", login, err)
		} else {
			logger.Warnf("Conta %s banida por %d minutos devido a multiplas tentativas falhas",
				login, am.config.BanDurationMinutes)
		}

		delete(am.attempts, login)
	}
}

func (am *AuthManager) clearFailedAttempts(login string) {
	delete(am.attempts, login)
}
