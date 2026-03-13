package protocol

import "fmt"

// LoginClientState representa os estados do cliente durante autenticação
type LoginClientState int

const (
	StateConnected LoginClientState = iota
	StateAuthedGG
	StateAuthedLogin
)

func (s LoginClientState) String() string {
	switch s {
	case StateConnected:
		return "CONNECTED"
	case StateAuthedGG:
		return "AUTHED_GG"
	case StateAuthedLogin:
		return "AUTHED_LOGIN"
	default:
		return "UNKNOWN"
	}
}

// SessionKey armazena as chaves de sessão para autenticação no GameServer
type SessionKey struct {
	PlayOkID1  uint32
	PlayOkID2  uint32
	LoginOkID1 uint32
	LoginOkID2 uint32
}

func NewSessionKey(sessionID uint32) *SessionKey {
	return &SessionKey{
		PlayOkID1:  sessionID,
		PlayOkID2:  sessionID ^ 0xFFFFFFFF,
		LoginOkID1: sessionID,
		LoginOkID2: sessionID ^ 0xFFFFFFFF,
	}
}

func (sk *SessionKey) CheckLoginPair(loginOk1, loginOk2 uint32) bool {
	return sk.LoginOkID1 == loginOk1 && sk.LoginOkID2 == loginOk2
}

func (sk *SessionKey) String() string {
	return fmt.Sprintf("PlayOk: %d %d LoginOk: %d %d",
		sk.PlayOkID1, sk.PlayOkID2, sk.LoginOkID1, sk.LoginOkID2)
}

var (
	ErrUserOrPassWrong   = fmt.Errorf("usuario ou senha incorretos")
	ErrPassWrong         = fmt.Errorf("senha incorreta")
	ErrAccessFailed      = fmt.Errorf("falha de acesso")
	ErrAccountInUse      = fmt.Errorf("conta em uso")
	ErrPermanentlyBanned = fmt.Errorf("conta banida permanentemente")
)
