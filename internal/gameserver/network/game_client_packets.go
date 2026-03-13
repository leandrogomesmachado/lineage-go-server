package network

import "strings"

type sendProtocolVersionPacket struct {
	version uint32
}

func lerSendProtocolVersionPacket(dados []byte) *sendProtocolVersionPacket {
	leitor := novoLeitorPacket(dados)
	return &sendProtocolVersionPacket{version: leitor.lerD()}
}

type authLoginPacket struct {
	loginName string
	playKey2  uint32
	playKey1  uint32
	loginKey1 uint32
	loginKey2 uint32
	extra     uint32
}

func lerAuthLoginPacket(dados []byte) *authLoginPacket {
	leitor := novoLeitorPacket(dados)
	return &authLoginPacket{
		loginName: strings.ToLower(leitor.lerS()),
		playKey2:  leitor.lerD(),
		playKey1:  leitor.lerD(),
		loginKey1: leitor.lerD(),
		loginKey2: leitor.lerD(),
		extra:     leitor.lerD(),
	}
}

type requestNewCharacterPacket struct {
}

func lerRequestNewCharacterPacket(dados []byte) *requestNewCharacterPacket {
	return &requestNewCharacterPacket{}
}

type requestCharacterCreatePacket struct {
	nome      string
	race      int32
	sexo      int32
	classID   int32
	hairStyle int32
	hairColor int32
	face      int32
}

func lerRequestCharacterCreatePacket(dados []byte) *requestCharacterCreatePacket {
	leitor := novoLeitorPacket(dados)
	pacote := &requestCharacterCreatePacket{
		nome:    leitor.lerS(),
		race:    int32(leitor.lerD()),
		sexo:    int32(leitor.lerD()),
		classID: int32(leitor.lerD()),
	}
	leitor.lerD()
	leitor.lerD()
	leitor.lerD()
	leitor.lerD()
	leitor.lerD()
	leitor.lerD()
	pacote.hairStyle = int32(leitor.lerD())
	pacote.hairColor = int32(leitor.lerD())
	pacote.face = int32(leitor.lerD())
	return pacote
}

type requestCharacterDeletePacket struct {
	slot int32
}

func lerRequestCharacterDeletePacket(dados []byte) *requestCharacterDeletePacket {
	leitor := novoLeitorPacket(dados)
	return &requestCharacterDeletePacket{slot: int32(leitor.lerD())}
}

type requestGameStartPacket struct {
	slot int32
}

func lerRequestGameStartPacket(dados []byte) *requestGameStartPacket {
	leitor := novoLeitorPacket(dados)
	pacote := &requestGameStartPacket{slot: int32(leitor.lerD())}
	leitor.lerH()
	leitor.lerD()
	leitor.lerD()
	leitor.lerD()
	return pacote
}

type enterWorldPacket struct {
}

func lerEnterWorldPacket(dados []byte) *enterWorldPacket {
	return &enterWorldPacket{}
}

type requestItemListPacket struct {
}

func lerRequestItemListPacket(dados []byte) *requestItemListPacket {
	_ = dados
	return &requestItemListPacket{}
}

type moveBackwardToLocationPacket struct {
	targetX       int32
	targetY       int32
	targetZ       int32
	originX       int32
	originY       int32
	originZ       int32
	tipoMovimento int32
}

func lerMoveBackwardToLocationPacket(dados []byte) *moveBackwardToLocationPacket {
	leitor := novoLeitorPacket(dados)
	return &moveBackwardToLocationPacket{
		targetX:       int32(leitor.lerD()),
		targetY:       int32(leitor.lerD()),
		targetZ:       int32(leitor.lerD()),
		originX:       int32(leitor.lerD()),
		originY:       int32(leitor.lerD()),
		originZ:       int32(leitor.lerD()),
		tipoMovimento: int32(leitor.lerD()),
	}
}

type validatePositionPacket struct {
	x       int32
	y       int32
	z       int32
	heading int32
	boatID  int32
}

func lerValidatePositionPacket(dados []byte) *validatePositionPacket {
	leitor := novoLeitorPacket(dados)
	return &validatePositionPacket{
		x:       int32(leitor.lerD()),
		y:       int32(leitor.lerD()),
		z:       int32(leitor.lerD()),
		heading: int32(leitor.lerD()),
		boatID:  int32(leitor.lerD()),
	}
}

type requestRestartPacket struct {
}

func lerRequestRestartPacket(dados []byte) *requestRestartPacket {
	return &requestRestartPacket{}
}

type requestSkillCoolTimePacket struct {
}

func lerRequestSkillCoolTimePacket(dados []byte) *requestSkillCoolTimePacket {
	return &requestSkillCoolTimePacket{}
}

type characterRestorePacket struct {
	slot int32
}

func lerCharacterRestorePacket(dados []byte) *characterRestorePacket {
	leitor := novoLeitorPacket(dados)
	return &characterRestorePacket{slot: int32(leitor.lerD())}
}

type logoutPacket struct {
}

func lerLogoutPacket(dados []byte) *logoutPacket {
	return &logoutPacket{}
}
