package network

type initLSPacket struct {
	revision uint32
	chaveRSA []byte
}

func lerInitLSPacket(dados []byte) *initLSPacket {
	leitor := novoLeitorPacket(dados)
	revision := leitor.lerD()
	tamanhoChave := int(leitor.lerD())
	return &initLSPacket{
		revision: revision,
		chaveRSA: leitor.lerB(tamanhoChave),
	}
}

type authResponsePacket struct {
	serverID   byte
	serverName string
}

func lerAuthResponsePacket(dados []byte) *authResponsePacket {
	leitor := novoLeitorPacket(dados)
	return &authResponsePacket{
		serverID:   leitor.lerC(),
		serverName: leitor.lerS(),
	}
}

type loginServerFailPacket struct {
	motivo byte
}

func lerLoginServerFailPacket(dados []byte) *loginServerFailPacket {
	leitor := novoLeitorPacket(dados)
	return &loginServerFailPacket{motivo: leitor.lerC()}
}

type playerAuthResponsePacket struct {
	conta  string
	authed bool
}

func lerPlayerAuthResponsePacket(dados []byte) *playerAuthResponsePacket {
	leitor := novoLeitorPacket(dados)
	return &playerAuthResponsePacket{
		conta:  leitor.lerS(),
		authed: leitor.lerC() != 0,
	}
}
