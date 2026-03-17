package network

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"
	"sync"

	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/protocol"
)

func resumirHexGameServer(dados []byte, limite int) string {
	if len(dados) == 0 {
		return ""
	}
	if len(dados) <= limite {
		return hex.EncodeToString(dados)
	}
	return hex.EncodeToString(dados[:limite])
}

type estadoGameClient int

const (
	estadoConectado estadoGameClient = iota
	estadoAuthed
	estadoEntering
	estadoInGame
)

type gameClient struct {
	conn               net.Conn
	server             *gameServer
	crypt              *gameCrypt
	estado             estadoGameClient
	conta              string
	sessionKey         *protocol.SessionKey
	chaveCriptografia  []byte
	mutexEnvio         sync.Mutex
	personagemAtual    *gsdb.CharacterSlot
	slotSelecionado    int32
	playerAtivo        *playerAtivo
	hennasAtivas       []gsdb.CharacterHenna
	skillsAtivas       []gsdb.CharacterSkill
	atalhosAtivos      []gsdb.CharacterShortcut
	subclassesAtivas   []gsdb.CharacterSubclass
	itensAtivos        []gsdb.CharacterItem
	augmentacoesAtivas []gsdb.CharacterAugmentation
	trainerPessoal     *npcTrainerRuntime
	canalEncerramento  chan struct{}
	cooldownsSkill     map[int32]int64
}

func novoGameClient(conn net.Conn, server *gameServer) *gameClient {
	return &gameClient{
		conn:              conn,
		server:            server,
		crypt:             novoGameCrypt(),
		estado:            estadoConectado,
		chaveCriptografia: gerarChaveGameCliente(),
		canalEncerramento: make(chan struct{}),
		cooldownsSkill:    make(map[int32]int64),
	}
}

func (g *gameClient) loop() {
	defer close(g.canalEncerramento)
	defer g.conn.Close()
	for {
		dados, err := g.lerPacket()
		if err != nil {
			if err != io.EOF {
				logger.Errorf("Erro ao ler packet do cliente de jogo %s: %v", g.conn.RemoteAddr().String(), err)
			}
			if err == io.EOF {
				logger.Infof("Cliente de jogo desconectou: %s", g.conn.RemoteAddr().String())
			}
			g.server.removerCliente(g)
			return
		}
		logger.Infof("Packet recebido do cliente de jogo %s: %d bytes opcode=0x%02X estado=%d", g.conn.RemoteAddr().String(), len(dados), dados[0], g.estado)
		if err = g.processarPacket(dados); err != nil {
			logger.Errorf("Erro ao processar packet do cliente de jogo: %v", err)
			g.server.removerCliente(g)
			return
		}
	}
}

func (g *gameClient) lerPacket() ([]byte, error) {
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
	g.crypt.decrypt(dados, 0, len(dados))
	return dados, nil
}

func (g *gameClient) enviarPacket(payload []byte) error {
	if len(payload) > 0 {
		opcode := payload[0]
		if opcode == 0x03 || opcode == 0x04 || opcode == 0xc1 || opcode == 0xe4 || opcode == 0xf3 {
			logger.Infof("Enviando packet opcode=0x%02X tamanho=%d conta=%s hex=%s", opcode, len(payload), g.conta, resumirHexGameServer(payload, 256))
		}
		if opcode == 0xfe && len(payload) >= 3 {
			subOpcode := binary.LittleEndian.Uint16(payload[1:3])
			if subOpcode == 0x2e {
				logger.Infof("Enviando packet opcode=0xFE:0x%04X tamanho=%d conta=%s hex=%s", subOpcode, len(payload), g.conta, resumirHexGameServer(payload, 256))
			}
		}
	}
	buffer := make([]byte, len(payload))
	copy(buffer, payload)
	g.crypt.encrypt(buffer, 0, len(buffer))
	pacote := make([]byte, len(buffer)+2)
	binary.LittleEndian.PutUint16(pacote[:2], uint16(len(buffer)+2))
	copy(pacote[2:], buffer)
	g.mutexEnvio.Lock()
	defer g.mutexEnvio.Unlock()
	_, err := io.Copy(g.conn, bytes.NewReader(pacote))
	return err
}

func (g *gameClient) processarPacket(dados []byte) error {
	if len(dados) == 0 {
		return nil
	}
	opcode := dados[0]
	if g.estado == estadoConectado && opcode == 0x00 {
		packet := lerSendProtocolVersionPacket(dados)
		return g.processarProtocolVersion(packet)
	}
	if g.estado == estadoConectado && opcode == 0x08 {
		logger.Infof("AuthLogin bruto do cliente de jogo %s: %s", g.conn.RemoteAddr().String(), resumirHexGameServer(dados, 96))
		packet := lerAuthLoginPacket(dados)
		return g.processarAuthLogin(packet)
	}
	if g.estado == estadoAuthed && opcode == 0x0e {
		packet := lerRequestNewCharacterPacket(dados)
		return g.processarRequestNewCharacter(packet)
	}
	if g.estado == estadoAuthed && opcode == 0x0b {
		packet := lerRequestCharacterCreatePacket(dados)
		return g.processarRequestCharacterCreate(packet)
	}
	if g.estado == estadoAuthed && opcode == 0x0c {
		packet := lerRequestCharacterDeletePacket(dados)
		return g.processarRequestCharacterDelete(packet)
	}
	if g.estado == estadoAuthed && opcode == 0x0d {
		packet := lerRequestGameStartPacket(dados)
		return g.processarRequestGameStart(packet)
	}
	if g.estado == estadoAuthed && opcode == 0x62 {
		packet := lerCharacterRestorePacket(dados)
		return g.processarCharacterRestore(packet)
	}
	if g.estado == estadoEntering && opcode == 0x03 {
		packet := lerEnterWorldPacket(dados)
		return g.processarEnterWorld(packet)
	}
	if g.estado == estadoInGame && opcode == 0x01 {
		packet := lerMoveBackwardToLocationPacket(dados)
		return g.processarMoveBackwardToLocation(packet)
	}
	if g.estado == estadoInGame && opcode == 0x04 {
		packet := lerActionPacket(dados)
		return g.processarAction(packet)
	}
	if g.estado == estadoInGame && opcode == 0x0a {
		packet := lerAttackRequestPacket(dados)
		return g.processarAttackRequest(packet)
	}
	if g.estado == estadoInGame && opcode == 0x2a {
		packet := lerRequestTargetCancelPacket(dados)
		return g.processarRequestTargetCancel(packet)
	}
	if g.estado == estadoInGame && opcode == 0x0f {
		packet := lerRequestItemListPacket(dados)
		return g.processarRequestItemList(packet)
	}
	if g.estado == estadoInGame && opcode == 0x21 {
		packet := lerRequestBypassToServerPacket(dados)
		return g.processarRequestBypassToServer(packet)
	}
	if g.estado == estadoInGame && opcode == 0x2f {
		packet := lerRequestMagicSkillUsePacket(dados)
		return g.processarRequestMagicSkillUse(packet)
	}
	if g.estado == estadoInGame && opcode == 0x45 {
		packet := lerRequestActionUsePacket(dados)
		return g.processarRequestActionUse(packet)
	}
	if g.estado == estadoInGame && opcode == 0x48 {
		packet := lerValidatePositionPacket(dados)
		return g.processarValidatePosition(packet)
	}
	if g.estado == estadoInGame && opcode == 0x6b {
		packet := lerRequestAcquireSkillInfoPacket(dados)
		return g.processarRequestAcquireSkillInfo(packet)
	}
	if g.estado == estadoInGame && opcode == 0x6c {
		packet := lerRequestAcquireSkillPacket(dados)
		return g.processarRequestAcquireSkill(packet)
	}
	if g.estado == estadoInGame && opcode == 0x46 {
		packet := lerRequestRestartPacket(dados)
		return g.processarRequestRestart(packet)
	}
	if g.estado == estadoInGame && opcode == 0x9d {
		packet := lerRequestSkillCoolTimePacket(dados)
		_ = packet
		return g.enviarPacket(montarSkillCoolTimePacket())
	}
	if (g.estado == estadoAuthed || g.estado == estadoInGame) && opcode == 0x09 {
		packet := lerLogoutPacket(dados)
		return g.processarLogout(packet)
	}
	return nil
}

func (g *gameClient) processarProtocolVersion(packet *sendProtocolVersionPacket) error {
	if packet.version != versaoProtocoloInterlude1 && packet.version != versaoProtocoloInterlude2 && packet.version != versaoProtocoloInterlude3 && packet.version != versaoProtocoloInterlude4 {
		logger.Warnf("Versao de protocolo invalida do cliente de jogo %s: %d", g.conn.RemoteAddr().String(), packet.version)
		g.conn.Close()
		return nil
	}
	g.crypt.setKey(g.chaveCriptografia)
	logger.Infof("Enviando VersionCheck para %s com chave %x", g.conn.RemoteAddr().String(), g.chaveCriptografia[:8])
	return g.enviarPacket(montarVersionCheckPacket(g.chaveCriptografia))
}

func (g *gameClient) processarAuthLogin(packet *authLoginPacket) error {
	logger.Infof("AuthLogin recebido do cliente de jogo %s para conta %s", g.conn.RemoteAddr().String(), packet.loginName)
	logger.Infof("Chaves lidas do AuthLogin para conta %s: play=%d/%d login=%d/%d extra=%d", packet.loginName, packet.playKey1, packet.playKey2, packet.loginKey1, packet.loginKey2, packet.extra)
	g.conta = packet.loginName
	g.sessionKey = &protocol.SessionKey{
		PlayOkID1:  packet.playKey1,
		PlayOkID2:  packet.playKey2,
		LoginOkID1: packet.loginKey1,
		LoginOkID2: packet.loginKey2,
	}
	return g.server.registrarAutenticacaoPendente(g)
}

func (g *gameClient) autenticarComSucesso(slots []gsdb.CharacterSlot) error {
	g.estado = estadoAuthed
	logger.Infof("Autenticacao do cliente de jogo bem-sucedida para conta %s com %d slots", g.conta, len(slots))
	return g.enviarPacket(montarCharSelectInfoPacket(g.conta, g.sessionKey.PlayOkID1, slots))
}

func (g *gameClient) autenticarComFalha() error {
	logger.Warnf("Autenticacao do cliente de jogo falhou para conta %s", g.conta)
	return g.enviarPacket(montarAuthLoginFailPacket(failReasonSystemErrorLoginLater))
}

func gerarChaveGameCliente() []byte {
	chave := gerarHexAleatorio(16)
	chave[8] = 0xc8
	chave[9] = 0x27
	chave[10] = 0x93
	chave[11] = 0x01
	chave[12] = 0xa1
	chave[13] = 0x6c
	chave[14] = 0x31
	chave[15] = 0x97
	return chave
}
