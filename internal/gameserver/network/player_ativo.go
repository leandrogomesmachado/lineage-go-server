package network

import (
	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
)

type playerAtivo struct {
	objID     int32
	conta     string
	nome      string
	classID   int32
	race      int32
	sexo      int32
	nivel     int32
	exp       int64
	sp        int32
	hpAtual   int32
	hpMaximo  int32
	mpAtual   int32
	mpMaximo  int32
	cpAtual   int32
	cpMaximo  int32
	x         int32
	y         int32
	z         int32
	heading   int32
	titulo    string
	clanID    int32
	karma     int32
	pkKills   int32
	pvpKills  int32
	hairStyle int32
	hairColor int32
	face      int32
	baseClass int32
	regiaoX   int32
	regiaoY   int32
}

func novoPlayerAtivo(conta string, slot gsdb.CharacterSlot) *playerAtivo {
	player := &playerAtivo{
		objID:     slot.ObjID,
		conta:     conta,
		nome:      slot.CharName,
		classID:   slot.ClassID,
		race:      slot.Race,
		sexo:      slot.Sex,
		nivel:     slot.Level,
		exp:       slot.Exp,
		sp:        slot.Sp,
		hpAtual:   slot.CurHp,
		hpMaximo:  slot.MaxHp,
		mpAtual:   slot.CurMp,
		mpMaximo:  slot.MaxMp,
		cpAtual:   slot.CurCp,
		cpMaximo:  slot.MaxCp,
		x:         slot.X,
		y:         slot.Y,
		z:         slot.Z,
		heading:   0,
		titulo:    slot.Title,
		clanID:    slot.ClanID,
		karma:     slot.Karma,
		pkKills:   slot.PkKills,
		pvpKills:  slot.PvpKills,
		hairStyle: slot.HairStyle,
		hairColor: slot.HairColor,
		face:      slot.Face,
		baseClass: slot.BaseClass,
	}
	player.atualizarRegiao()
	return player
}

func (p *playerAtivo) aplicarPosicao(x int32, y int32, z int32, heading int32) {
	p.x = x
	p.y = y
	p.z = z
	p.heading = heading
	p.atualizarRegiao()
}

func (p *playerAtivo) atualizarRegiao() {
	p.regiaoX = calcularRegiaoX(p.x)
	p.regiaoY = calcularRegiaoY(p.y)
}
