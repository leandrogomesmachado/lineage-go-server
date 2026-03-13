package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/controller"
	"github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/crypto"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type gameServerThread struct {
	conn            net.Conn
	config          *config.LoginServerConfig
	loginController *controller.LoginController
	rsaKeyPair      *crypto.RSAKeyPair
	blowfish        *crypto.NewCrypt
	info            *gameServerInfo
	mutexEnvio      sync.Mutex
	enderecoConexao string
}

func novoGameServerThread(conn net.Conn, cfg *config.LoginServerConfig, loginCtrl *controller.LoginController) (*gameServerThread, error) {
	rsaKeyPair := loginCtrl.GetScrambledRSAKeyPair()
	blowfish, err := novaCriptografiaGameServer()
	if err != nil {
		return nil, err
	}
	thread := &gameServerThread{
		conn:            conn,
		config:          cfg,
		loginController: loginCtrl,
		rsaKeyPair:      rsaKeyPair,
		blowfish:        blowfish,
		enderecoConexao: conn.RemoteAddr().String(),
	}
	go thread.loop()
	return thread, nil
}

func (g *gameServerThread) loop() {
	defer g.conn.Close()
	if err := g.enviarPacketGameServer(montarInitLSPacket(g.rsaKeyPair.GetPublicModulusBytes())); err != nil {
		logger.Errorf("Erro ao enviar InitLS para GameServer %s: %v", g.enderecoConexao, err)
		return
	}

	for {
		dados, err := g.lerPacketGameServer()
		if err != nil {
			if err != io.EOF {
				logger.Errorf("Erro ao ler packet do GameServer %s: %v", g.enderecoConexao, err)
			}
			return
		}
		if err := g.processarPacketGameServer(dados); err != nil {
			logger.Errorf("Erro ao processar packet do GameServer %s: %v", g.enderecoConexao, err)
			return
		}
	}
}

func (g *gameServerThread) lerPacketGameServer() ([]byte, error) {
	tamanhoBuf := make([]byte, 2)
	if _, err := io.ReadFull(g.conn, tamanhoBuf); err != nil {
		return nil, err
	}
	tamanho := binary.LittleEndian.Uint16(tamanhoBuf)
	if tamanho < 2 {
		return nil, io.ErrUnexpectedEOF
	}
	dados := make([]byte, tamanho-2)
	if _, err := io.ReadFull(g.conn, dados); err != nil {
		return nil, err
	}
	if err := g.blowfish.Decrypt(dados, 0, len(dados)); err != nil {
		return nil, err
	}
	if !crypto.VerifyChecksum(dados, 0, len(dados)) {
		return nil, fmt.Errorf("checksum invalido no packet do gameserver")
	}
	return dados, nil
}

func (g *gameServerThread) enviarPacketGameServer(payload []byte) error {
	buffer := make([]byte, len(payload)+12)
	copy(buffer, payload)
	tamanho := len(payload) + 4
	if resto := tamanho % 8; resto != 0 {
		tamanho += 8 - resto
	}
	crypto.AppendChecksum(buffer, 0, tamanho)
	if err := g.blowfish.Encrypt(buffer, 0, tamanho); err != nil {
		return err
	}
	g.mutexEnvio.Lock()
	defer g.mutexEnvio.Unlock()
	pacote := make([]byte, tamanho+2)
	binary.LittleEndian.PutUint16(pacote[:2], uint16(tamanho+2))
	copy(pacote[2:], buffer[:tamanho])
	_, err := io.Copy(g.conn, bytes.NewReader(pacote))
	return err
}

func (g *gameServerThread) processarPacketGameServer(dados []byte) error {
	if len(dados) == 0 {
		return nil
	}
	opcode := dados[0]
	if opcode == 0x00 {
		return g.processarBlowfishKey(append([]byte{0x00}, dados[1:]...))
	}
	if opcode == 0x01 {
		return g.processarGameServerAuth(append([]byte{0x01}, dados[1:]...))
	}
	if opcode == 0x02 {
		return g.processarPlayerInGame(append([]byte{0x02}, dados[1:]...))
	}
	if opcode == 0x03 {
		return g.processarPlayerLogout(append([]byte{0x03}, dados[1:]...))
	}
	if opcode == 0x05 {
		return g.processarPlayerAuthRequest(append([]byte{0x05}, dados[1:]...))
	}
	return fmt.Errorf("opcode de gameserver nao suportado: 0x%02X", opcode)
}

func (g *gameServerThread) processarBlowfishKey(dados []byte) error {
	leitor := novoLeitorPacketCliente(dados)
	tamanhoChave := int(leitor.lerD())
	chaveCriptografada := leitor.lerB(tamanhoChave)
	chaveDecriptada, err := g.rsaKeyPair.Decrypt(chaveCriptografada)
	if err != nil {
		return err
	}
	indice := 0
	for indice < len(chaveDecriptada) && chaveDecriptada[indice] == 0 {
		indice++
	}
	chaveBlowfish := chaveDecriptada[indice:]
	blowfish, err := crypto.NewNewCrypt(chaveBlowfish)
	if err != nil {
		return err
	}
	g.blowfish = blowfish
	logger.Infof("Chave Blowfish do GameServer %s atualizada", g.enderecoConexao)
	return nil
}

func (g *gameServerThread) processarGameServerAuth(dados []byte) error {
	packet := lerGameServerAuthPacket(dados)
	for _, serverCfg := range g.config.GameServers {
		if serverCfg.ID != int(packet.desiredID) {
			continue
		}
		info := &gameServerInfo{
			id:         serverCfg.ID,
			nome:       serverCfg.Name,
			host:       serverCfg.Host,
			porta:      int(packet.porta),
			maxPlayers: int(packet.maxPlayers),
			thread:     g,
		}
		if strings.TrimSpace(packet.hostName) != "" && packet.hostName != "*" {
			info.host = packet.hostName
		}
		info.setAuthed(true)
		g.info = info
		logger.Infof("GameServer autenticado: id=%d nome=%s host=%s porta=%d", info.id, info.nome, info.host, info.porta)
		return g.enviarPacketGameServer(montarAuthResponsePacket(byte(info.id), info.nome))
	}
	return g.enviarPacketGameServer(montarLoginServerFailPacket(3))
}

func (g *gameServerThread) processarPlayerInGame(dados []byte) error {
	if g.info == nil || !g.info.isAuthed() {
		return g.enviarPacketGameServer(montarLoginServerFailPacket(6))
	}
	packet := lerPlayerInGamePacket(dados)
	for _, conta := range packet.contas {
		g.info.adicionarConta(conta)
	}
	return nil
}

func (g *gameServerThread) processarPlayerLogout(dados []byte) error {
	if g.info == nil || !g.info.isAuthed() {
		return g.enviarPacketGameServer(montarLoginServerFailPacket(6))
	}
	packet := lerPlayerLogoutPacket(dados)
	g.info.removerConta(packet.conta)
	return nil
}

func (g *gameServerThread) processarPlayerAuthRequest(dados []byte) error {
	if g.info == nil || !g.info.isAuthed() {
		return g.enviarPacketGameServer(montarLoginServerFailPacket(6))
	}
	packet := lerPlayerAuthRequestPacket(dados)
	logger.Infof("GameServerThread usando LoginController %p recebeu PlayerAuthRequest para conta=%q len=%d com chave play=%d/%d login=%d/%d", g.loginController, packet.conta, len(packet.conta), packet.playOkID1, packet.playOkID2, packet.loginOkID1, packet.loginOkID2)
	key := g.loginController.GetKeyForAccount(packet.conta)
	if key == nil {
		logger.Warnf("Nenhuma session key encontrada no LoginServer para conta %s", packet.conta)
		return g.enviarPacketGameServer(montarPlayerAuthResponsePacket(packet.conta, false))
	}
	logger.Infof("Session key esperada no LoginServer para conta %s: play=%d/%d login=%d/%d", packet.conta, key.PlayOkID1, key.PlayOkID2, key.LoginOkID1, key.LoginOkID2)
	if key.PlayOkID1 != packet.playOkID1 || key.PlayOkID2 != packet.playOkID2 {
		logger.Warnf("Session key divergente para conta %s", packet.conta)
		return g.enviarPacketGameServer(montarPlayerAuthResponsePacket(packet.conta, false))
	}
	if g.loginController.ShowLicence() && (key.LoginOkID1 != packet.loginOkID1 || key.LoginOkID2 != packet.loginOkID2) {
		logger.Warnf("Session key de login divergente para conta %s com ShowLicence ativo", packet.conta)
		return g.enviarPacketGameServer(montarPlayerAuthResponsePacket(packet.conta, false))
	}
	if !g.loginController.ShowLicence() && (packet.loginOkID1 != 0 || packet.loginOkID2 != 0) {
		logger.Infof("Ignorando login keys informadas no PlayerAuthRequest para conta %s porque ShowLicence esta desativado: login=%d/%d", packet.conta, packet.loginOkID1, packet.loginOkID2)
	}
	logger.Infof("PlayerAuthRequest validado com sucesso para conta %s", packet.conta)
	g.loginController.RemoveAuthedLoginClient(packet.conta)
	return g.enviarPacketGameServer(montarPlayerAuthResponsePacket(packet.conta, true))
}
