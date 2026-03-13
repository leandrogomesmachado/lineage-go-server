package network

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/controller"
	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/crypto"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/protocol"
)

type LoginClient struct {
	conn            net.Conn
	loginController *controller.LoginController
	rsaKeyPair      *crypto.RSAKeyPair
	loginCrypt      *crypto.LoginCrypt
	sessionID       uint32
	sessionKey      *protocol.SessionKey
	state           protocol.LoginClientState
	username        string
	accountID       uint32
	joinedGS        bool
	connectionStart time.Time
}

func NewLoginClient(conn net.Conn, loginCtrl *controller.LoginController) *LoginClient {
	rsaKeyPair := loginCtrl.GetScrambledRSAKeyPair()
	sessionID := generateRandomSessionID()
	logger.Infof("Criando LoginClient para %s com LoginController %p", conn.RemoteAddr().String(), loginCtrl)

	return &LoginClient{
		conn:            conn,
		loginController: loginCtrl,
		rsaKeyPair:      rsaKeyPair,
		sessionID:       sessionID,
		state:           protocol.StateConnected,
		connectionStart: time.Now(),
	}
}

func generateRandomSessionID() uint32 {
	n, _ := rand.Int(rand.Reader, big.NewInt(0x7FFFFFFF))
	return uint32(n.Int64())
}

func hashStringToUint32(s string) uint32 {
	hash := uint32(0)
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return hash
}

func resumirHexadecimal(dados []byte, limite int) string {
	if len(dados) == 0 {
		return ""
	}
	if len(dados) <= limite {
		return hex.EncodeToString(dados)
	}
	return hex.EncodeToString(dados[:limite])
}

func normalizarCredencialTexto(dados []byte) string {
	texto := strings.TrimSpace(string(dados))
	texto = strings.Trim(texto, "\x00")
	if indiceNulo := strings.IndexByte(texto, 0); indiceNulo >= 0 {
		return strings.TrimSpace(texto[:indiceNulo])
	}
	return texto
}

func (lc *LoginClient) Handle(ctx context.Context) {
	logger.Infof("Iniciando handler para cliente %s", lc.conn.RemoteAddr().String())
	defer lc.onDisconnect()

	if err := lc.sendInit(); err != nil {
		logger.Errorf("Erro ao enviar Init: %v", err)
		return
	}
	logger.Info("Init packet enviado com sucesso")

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Contexto cancelado, encerrando handler")
			return
		default:
			packet, err := lc.readPacket()
			if err != nil {
				if err != io.EOF {
					logger.Errorf("Erro ao ler packet: %v", err)
				} else {
					logger.Info("Cliente desconectou")
				}
				return
			}
			logger.Infof("Packet recebido: %d bytes, opcode: 0x%02X", len(packet), packet[0])

			if err := lc.processarPacket(ctx, packet); err != nil {
				logger.Errorf("Erro ao processar packet: %v", err)
				return
			}
		}
	}
}

func (lc *LoginClient) sendInit() error {
	modulus := lc.rsaKeyPair.GetModulusBytes()
	blowfishKey := lc.loginController.GetRandomBlowfishKey()

	var err error
	lc.loginCrypt, err = crypto.NewLoginCrypt(blowfishKey)
	if err != nil {
		return err
	}

	if len(modulus) > 128 {
		modulus = modulus[:128]
	}
	if len(modulus) < 128 {
		moduloAjustado := make([]byte, 128)
		copy(moduloAjustado[128-len(modulus):], modulus)
		modulus = moduloAjustado
	}

	buf := lc.montarInitPacket(modulus, blowfishKey)

	logger.Infof("Init packet montado: %d bytes (antes criptografia)", len(buf))
	logger.Infof("Init payload hex: %s", resumirHexadecimal(buf, 64))

	return lc.writePacket(buf)
}

func (lc *LoginClient) readPacket() ([]byte, error) {
	lc.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	sizeBuf := make([]byte, 2)
	if _, err := io.ReadFull(lc.conn, sizeBuf); err != nil {
		return nil, err
	}

	size := binary.LittleEndian.Uint16(sizeBuf)
	if size < 2 {
		return nil, io.ErrUnexpectedEOF
	}

	packet := make([]byte, size)
	copy(packet[:2], sizeBuf)
	if _, err := io.ReadFull(lc.conn, packet[2:]); err != nil {
		return nil, err
	}

	if lc.loginCrypt != nil {
		if !lc.loginCrypt.Decrypt(packet, 2, int(size-2)) {
			logger.Warnf("Falha ao descriptografar packet (estado: %s, tamanho: %d)", lc.state, size-2)
			return nil, io.ErrUnexpectedEOF
		}
		logger.Debugf("Packet descriptografado com sucesso (estado: %s)", lc.state)
	}

	return packet[2:], nil
}

func (lc *LoginClient) writePacketRaw(data []byte) error {
	// Enviar packet SEM criptografia (usado apenas para Init)
	totalSize := uint16(len(data) + 2)
	sizeBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(sizeBuf, totalSize)

	if _, err := io.Copy(lc.conn, bytes.NewReader(sizeBuf)); err != nil {
		return err
	}

	_, err := io.Copy(lc.conn, bytes.NewReader(data))
	return err
}

func (lc *LoginClient) writePacket(data []byte) error {
	bufSize := len(data) + 22
	buf := make([]byte, bufSize)
	copy(buf[2:], data)

	var size int
	if lc.loginCrypt != nil {
		logger.Debugf("Criptografando packet: %d bytes (antes)", len(data))
		encSize, err := lc.loginCrypt.Encrypt(buf, 2, len(data))
		if err != nil {
			return err
		}
		size = encSize
		logger.Infof("Packet criptografado: %d bytes (depois)", size)
	} else {
		size = len(data)
	}

	totalSize := uint16(size + 2)
	binary.LittleEndian.PutUint16(buf[:2], totalSize)
	logger.Infof("Packet final hex: %s", resumirHexadecimal(buf[:totalSize], 96))

	logger.Debugf("Enviando header: %d bytes total (incluindo header)", totalSize)
	_, err := io.Copy(lc.conn, bytes.NewReader(buf[:totalSize]))
	return err
}

func (lc *LoginClient) handleAuthGameGuard(data []byte) error {
	if len(data) < 20 {
		logger.Warn("AuthGameGuard muito pequeno")
		return lc.sendLoginFail(protocol.ReasonAccessFailed)
	}

	packet := lerAuthGameGuardPacket(append([]byte{protocol.AuthGameGuard}, data...))
	if packet.sessionID != lc.sessionID {
		logger.Warnf("SessionID invalido no AuthGameGuard: esperado %d, recebido %d", lc.sessionID, packet.sessionID)
		return lc.sendLoginFail(protocol.ReasonAccessFailed)
	}

	logger.Debug("AuthGameGuard valido - mudando estado para AUTHED_GG")
	lc.state = protocol.StateAuthedGG
	return lc.sendGGAuth()
}

func (lc *LoginClient) handleAuthLogin(ctx context.Context, data []byte) error {
	logger.Debug("Processando RequestAuthLogin")

	if len(data) < 128 {
		logger.Errorf("RequestAuthLogin muito pequeno: %d bytes", len(data))
		return lc.sendLoginFail(protocol.ReasonSystemError)
	}

	packet := lerRequestAuthLoginPacket(append([]byte{protocol.RequestAuthLogin}, data...))
	decrypted, err := lc.rsaKeyPair.Decrypt(packet.dadosCriptografados)
	if err != nil {
		logger.Errorf("Erro ao descriptografar RSA: %v", err)
		return lc.sendLoginFail(protocol.ReasonAccessFailed)
	}

	if len(decrypted) < 124 {
		logger.Errorf("Dados descriptografados muito pequenos: %d bytes", len(decrypted))
		return lc.sendLoginFail(protocol.ReasonSystemError)
	}

	username := normalizarCredencialTexto(decrypted[0x5E : 0x5E+14])
	password := normalizarCredencialTexto(decrypted[0x6C : 0x6C+16])

	username = strings.ToLower(username)

	logger.Infof("Tentativa de login: %s", username)

	account, sessionKey, err := lc.loginController.RetrieveAccountInfo(ctx, lc.conn.RemoteAddr(), username, password)
	if err != nil {
		logger.Warnf("Falha na autenticacao para %s: %v", username, err)
		if err == protocol.ErrUserOrPassWrong || err == protocol.ErrPassWrong {
			return lc.sendLoginFail(protocol.ReasonUserOrPassWrong)
		} else if err == protocol.ErrAccountInUse {
			return lc.sendLoginFail(protocol.ReasonAccountInUse)
		} else if err == protocol.ErrPermanentlyBanned {
			return lc.sendLoginFail(protocol.ReasonSystemError)
		}
		return lc.sendLoginFail(protocol.ReasonAccessFailed)
	}

	lc.username = username
	lc.accountID = hashStringToUint32(username)
	lc.sessionKey = sessionKey
	lc.state = protocol.StateAuthedLogin

	logger.Infof("Login bem-sucedido: %s (access level: %d)", username, account.AccessLevel)

	if lc.loginController.ShowLicence() {
		return lc.sendLoginOk()
	}
	return lc.sendServerList()
}

func (lc *LoginClient) handleServerList(ctx context.Context, data []byte) error {
	if lc.state != protocol.StateAuthedLogin {
		logger.Warn("RequestServerList sem autenticacao")
		return lc.sendLoginFail(protocol.ReasonAccessFailed)
	}

	if len(data) < 8 {
		logger.Warn("RequestServerList muito pequeno")
		return lc.sendLoginFail(protocol.ReasonAccessFailed)
	}

	packet := lerRequestServerListPacket(append([]byte{protocol.RequestServerList}, data...))

	if !lc.sessionKey.CheckLoginPair(packet.loginOkID1, packet.loginOkID2) {
		logger.Warnf("Session key invalida: esperado %s, recebido %d %d", lc.sessionKey, packet.loginOkID1, packet.loginOkID2)
		return lc.sendLoginFail(protocol.ReasonAccessFailed)
	}

	logger.Debugf("Processando RequestServerList para %s", lc.username)

	return lc.sendServerList()
}

func (lc *LoginClient) handleServerLogin(ctx context.Context, data []byte) error {
	if lc.state != protocol.StateAuthedLogin {
		logger.Warn("RequestServerLogin sem autenticacao")
		return lc.sendPlayFail(protocol.ReasonAccessFailed)
	}

	if len(data) < 9 {
		logger.Warn("RequestServerLogin muito pequeno")
		return lc.sendPlayFail(protocol.ReasonAccessFailed)
	}

	packet := lerRequestServerLoginPacket(append([]byte{protocol.RequestServerLogin}, data...))

	if lc.loginController.ShowLicence() && !lc.sessionKey.CheckLoginPair(packet.loginOkID1, packet.loginOkID2) {
		logger.Warnf("Session key invalida no ServerLogin")
		return lc.sendLoginFail(protocol.ReasonAccessFailed)
	}

	lc.joinedGS = true
	lc.loginController.MarkJoinedGameServer(lc.username)
	logger.Infof("%s conectando ao GameServer ID: %d", lc.username, packet.serverID)

	return lc.sendPlayOk()
}

func (lc *LoginClient) onDisconnect() {
	if lc.username == "" {
		return
	}
	if lc.joinedGS {
		logger.Infof("Cliente de login %s desconectou apos joinedGS, preservando sessao autenticada", lc.username)
		return
	}
	if lc.connectionStart.Add(time.Duration(controller.LoginTimeout) * time.Millisecond).Before(time.Now()) {
		logger.Infof("Cliente de login %s desconectou por timeout, removendo sessao autenticada", lc.username)
		lc.loginController.RemoveAuthedLoginClient(lc.username)
		return
	}
	logger.Infof("Cliente de login %s desconectou antes de entrar no GameServer, removendo sessao autenticada", lc.username)
	lc.loginController.RemoveAuthedLoginClient(lc.username)
}

func (lc *LoginClient) sendGGAuth() error {
	logger.Debug("Enviando GGAuth response")
	return lc.writePacket(lc.montarGGAuthPacket())
}

func (lc *LoginClient) sendLoginOk() error {
	logger.Infof("Enviando LoginOk para %s (SessionKey: %s)", lc.username, lc.sessionKey)
	return lc.writePacket(lc.montarLoginOkPacket())
}

func (lc *LoginClient) sendServerList() error {
	logger.Infof("Enviando ServerList para %s (1 servidor disponivel)", lc.username)
	return lc.writePacket(montarServerListPacket())
}

func (lc *LoginClient) sendPlayOk() error {
	logger.Infof("%s autorizado para GameServer (PlayOk SessionKey: %d %d)", lc.username, lc.sessionKey.PlayOkID1, lc.sessionKey.PlayOkID2)
	return lc.writePacket(lc.montarPlayOkPacket())
}

func (lc *LoginClient) sendLoginFail(reason byte) error {
	return lc.writePacket(montarLoginFailPacket(reason))
}

func (lc *LoginClient) sendPlayFail(reason byte) error {
	return lc.writePacket(montarPlayFailPacket(reason))
}

func (lc *LoginClient) Close() {
	lc.conn.Close()
}
