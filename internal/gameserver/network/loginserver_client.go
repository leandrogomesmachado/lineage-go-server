package network

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"net"
	"sync"
	"time"

	logincrypto "github.com/leandrogomesmachado/l2raptors-go/internal/loginserver/crypto"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/config"
	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

const revisaoLoginServer = 0x0101

type clienteLoginServer struct {
	config      *config.GameServerConfig
	server      *gameServer
	conn        net.Conn
	blowfish    *logincrypto.NewCrypt
	blowfishKey []byte
	serverID    byte
	serverName  string
	mutexEnvio  sync.Mutex
}

func NovoClienteLoginServer(cfg *config.GameServerConfig, server *gameServer) *clienteLoginServer {
	return &clienteLoginServer{config: cfg, server: server}
}

func (c *clienteLoginServer) Iniciar() {
	go c.loop()
}

func (c *clienteLoginServer) loop() {
	for {
		if err := c.conectar(); err != nil {
			logger.Errorf("Erro na conexao com LoginServer: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		if err := c.processar(); err != nil {
			logger.Errorf("Erro no fluxo com LoginServer: %v", err)
		}
		if c.conn != nil {
			c.conn.Close()
		}
		time.Sleep(10 * time.Second)
	}
}

func (c *clienteLoginServer) conectar() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.config.LoginServer.Host, c.config.LoginServer.Port))
	if err != nil {
		return err
	}
	blowfish, err := logincrypto.NewNewCrypt([]byte("_;v.]05-31!|+-%xT!^["))
	if err != nil {
		conn.Close()
		return err
	}
	c.conn = conn
	c.blowfish = blowfish
	c.blowfishKey = gerarHexAleatorio(40)
	logger.Infof("Conectado ao LoginServer em %s:%d", c.config.LoginServer.Host, c.config.LoginServer.Port)
	return nil
}

func (c *clienteLoginServer) processar() error {
	for {
		dados, err := c.lerPacket()
		if err != nil {
			return err
		}
		if len(dados) == 0 {
			continue
		}
		opcode := dados[0]
		if opcode == 0x00 {
			if err := c.processarInitLS(dados); err != nil {
				return err
			}
			continue
		}
		if opcode == 0x01 {
			packet := lerLoginServerFailPacket(dados)
			return fmt.Errorf("loginserver recusou registro do gameserver, motivo=%d", packet.motivo)
		}
		if opcode == 0x02 {
			packet := lerAuthResponsePacket(dados)
			c.serverID = packet.serverID
			c.serverName = packet.serverName
			logger.Infof("GameServer registrado no LoginServer: id=%d nome=%s", c.serverID, c.serverName)
			continue
		}
		if opcode == 0x03 {
			packet := lerPlayerAuthResponsePacket(dados)
			logger.Infof("PlayerAuthResponse recebido do LoginServer para conta %s authed=%t", packet.conta, packet.authed)
			if c.server != nil {
				if err = c.server.concluirAutenticacao(packet.conta, packet.authed); err != nil {
					return err
				}
			}
			if packet.authed {
				if err = c.enviarPacket(montarPlayerInGamePacket([]string{packet.conta})); err != nil {
					return err
				}
			}
			continue
		}
	}
}

func (c *clienteLoginServer) processarInitLS(dados []byte) error {
	packet := lerInitLSPacket(dados)
	if packet.revision != revisaoLoginServer {
		return fmt.Errorf("revisao divergente do LoginServer: %d", packet.revision)
	}
	chavePublica, err := construirChavePublica(packet.chaveRSA)
	if err != nil {
		return err
	}
	chaveCriptografada, err := criptografarRSASemPadding(chavePublica, c.blowfishKey)
	if err != nil {
		return err
	}
	if err := c.enviarPacket(montarBlowFishKeyPacket(chaveCriptografada)); err != nil {
		return err
	}
	blowfish, err := logincrypto.NewNewCrypt(c.blowfishKey)
	if err != nil {
		return err
	}
	c.blowfish = blowfish
	return c.enviarPacket(montarAuthRequestPacket(byte(c.config.Server.ID), false, gerarHexAleatorio(16), c.config.Server.Host, uint16(c.config.Server.Port), false, uint32(c.config.Server.MaxPlayers)))
}

func (c *clienteLoginServer) lerPacket() ([]byte, error) {
	tamanhoBuf := make([]byte, 2)
	if _, err := io.ReadFull(c.conn, tamanhoBuf); err != nil {
		return nil, err
	}
	tamanho := binary.LittleEndian.Uint16(tamanhoBuf)
	if tamanho < 2 {
		return nil, io.ErrUnexpectedEOF
	}
	dados := make([]byte, tamanho-2)
	if _, err := io.ReadFull(c.conn, dados); err != nil {
		return nil, err
	}
	if err := c.blowfish.Decrypt(dados, 0, len(dados)); err != nil {
		return nil, err
	}
	if !logincrypto.VerifyChecksum(dados, 0, len(dados)) {
		return nil, fmt.Errorf("checksum invalido do loginserver")
	}
	return dados, nil
}

func (c *clienteLoginServer) enviarPacket(payload []byte) error {
	buffer := make([]byte, len(payload)+12)
	copy(buffer, payload)
	tamanho := len(payload) + 4
	if resto := tamanho % 8; resto != 0 {
		tamanho += 8 - resto
	}
	logincrypto.AppendChecksum(buffer, 0, tamanho)
	if err := c.blowfish.Encrypt(buffer, 0, tamanho); err != nil {
		return err
	}
	pacote := make([]byte, tamanho+2)
	binary.LittleEndian.PutUint16(pacote[:2], uint16(tamanho+2))
	copy(pacote[2:], buffer[:tamanho])
	c.mutexEnvio.Lock()
	defer c.mutexEnvio.Unlock()
	_, err := io.Copy(c.conn, bytes.NewReader(pacote))
	return err
}

func construirChavePublica(modulus []byte) (*rsaPublicKey, error) {
	if len(modulus) == 0 {
		return nil, fmt.Errorf("modulus RSA vazio")
	}
	return &rsaPublicKey{n: new(big.Int).SetBytes(modulus), e: 65537}, nil
}

type rsaPublicKey struct {
	n *big.Int
	e int
}

func criptografarRSASemPadding(chave *rsaPublicKey, dados []byte) ([]byte, error) {
	if chave == nil {
		return nil, fmt.Errorf("chave RSA ausente")
	}
	tamanho := (chave.n.BitLen() + 7) / 8
	if len(dados) > tamanho {
		return nil, fmt.Errorf("dados maiores do que o modulo RSA")
	}
	bloco := make([]byte, tamanho)
	copy(bloco[tamanho-len(dados):], dados)
	m := new(big.Int).SetBytes(bloco)
	if m.Cmp(chave.n) >= 0 {
		return nil, fmt.Errorf("mensagem RSA invalida para o modulo")
	}
	cifra := new(big.Int).Exp(m, big.NewInt(int64(chave.e)), chave.n)
	resultado := cifra.Bytes()
	if len(resultado) == tamanho {
		return resultado, nil
	}
	ajustado := make([]byte, tamanho)
	copy(ajustado[tamanho-len(resultado):], resultado)
	return ajustado, nil
}

func gerarHexAleatorio(tamanho int) []byte {
	resultado := make([]byte, tamanho)
	for i := 0; i < tamanho; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(256))
		resultado[i] = byte(n.Int64())
	}
	return resultado
}

func (c *clienteLoginServer) registrarClienteJogo(cliente *gameClient) error {
	if cliente == nil {
		return nil
	}
	if cliente.sessionKey == nil {
		return nil
	}
	logger.Infof("Enviando PlayerAuthRequest para conta %s com chave play=%d/%d login=%d/%d", cliente.conta, cliente.sessionKey.PlayOkID1, cliente.sessionKey.PlayOkID2, cliente.sessionKey.LoginOkID1, cliente.sessionKey.LoginOkID2)
	return c.enviarPacket(montarPlayerAuthRequestPacket(cliente.conta, cliente.sessionKey))
}
