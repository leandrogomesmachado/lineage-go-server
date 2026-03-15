package network

import (
	gsdb "github.com/leandrogomesmachado/l2raptors-go/internal/gameserver/database"
)

type playerAtivo struct {
	objID            int32
	conta            string
	nome             string
	classID          int32
	race             int32
	sexo             int32
	nivel            int32
	exp              int64
	sp               int32
	hpAtual          int32
	hpMaximo         int32
	mpAtual          int32
	mpMaximo         int32
	cpAtual          int32
	cpMaximo         int32
	x                int32
	y                int32
	z                int32
	heading          int32
	movendo          bool
	destinoX         int32
	destinoY         int32
	destinoZ         int32
	origemMovX       int32
	origemMovY       int32
	origemMovZ       int32
	titulo           string
	clanID           int32
	clanCrestID      int32
	clanCrestLargeID int32
	allyID           int32
	allyCrestID      int32
	karma            int32
	pkKills          int32
	pvpKills         int32
	nameColor        int32
	titleColor       int32
	clanPrivileges   int32
	pledgeClass      int32
	pledgeType       int32
	mountType        int32
	mountNpcID       int32
	operateType      int32
	team             int32
	abnormalEffect   int32
	recHave          int32
	recLeft          int32
	fishing          int32
	fishingX         int32
	fishingY         int32
	fishingZ         int32
	cubicIDs         []int32
	relation         int32
	hero             int32
	nobless          int32
	hairStyle        int32
	hairColor        int32
	face             int32
	baseClass        int32
	regiaoX          int32
	regiaoY          int32
	alvoObjID        int32
	sentado          bool
	correndo         bool
	ultimoMoveX      int32
	ultimoMoveY      int32
	ultimoMoveZ      int32
	ultimoPersistMs  int64
}

func novoPlayerAtivo(conta string, slot gsdb.CharacterSlot) *playerAtivo {
	player := &playerAtivo{
		objID:            slot.ObjID,
		conta:            conta,
		nome:             slot.CharName,
		classID:          slot.ClassID,
		race:             slot.Race,
		sexo:             slot.Sex,
		nivel:            slot.Level,
		exp:              slot.Exp,
		sp:               slot.Sp,
		hpAtual:          slot.CurHp,
		hpMaximo:         slot.MaxHp,
		mpAtual:          slot.CurMp,
		mpMaximo:         slot.MaxMp,
		cpAtual:          slot.CurCp,
		cpMaximo:         slot.MaxCp,
		x:                slot.X,
		y:                slot.Y,
		z:                slot.Z,
		heading:          slot.Heading,
		titulo:           slot.Title,
		clanID:           slot.ClanID,
		clanCrestID:      slot.ClanCrestID,
		clanCrestLargeID: slot.ClanCrestLargeID,
		allyID:           slot.AllyID,
		allyCrestID:      slot.AllyCrestID,
		karma:            slot.Karma,
		pkKills:          slot.PkKills,
		pvpKills:         slot.PvpKills,
		nameColor:        slot.NameColor,
		titleColor:       slot.TitleColor,
		clanPrivileges:   slot.ClanPrivileges,
		pledgeClass:      slot.PledgeClass,
		pledgeType:       slot.PledgeType,
		mountType:        slot.MountType,
		mountNpcID:       slot.MountNpcID,
		operateType:      slot.OperateType,
		team:             slot.Team,
		abnormalEffect:   slot.AbnormalEffect,
		recHave:          slot.RecHave,
		recLeft:          slot.RecLeft,
		fishing:          slot.Fishing,
		fishingX:         slot.FishingX,
		fishingY:         slot.FishingY,
		fishingZ:         slot.FishingZ,
		relation:         slot.Relation,
		hero:             slot.Hero,
		nobless:          slot.Nobless,
		hairStyle:        slot.HairStyle,
		hairColor:        slot.HairColor,
		face:             slot.Face,
		baseClass:        slot.BaseClass,
		correndo:         true,
	}
	player.ultimoMoveX = player.x
	player.ultimoMoveY = player.y
	player.ultimoMoveZ = player.z
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

func (p *playerAtivo) definirAlvo(objID int32) {
	p.alvoObjID = objID
}

func (p *playerAtivo) limparAlvo() {
	p.alvoObjID = 0
}

func (p *playerAtivo) estaSentado() bool {
	return p.sentado
}

func (p *playerAtivo) sentar() {
	p.sentado = true
}

func (p *playerAtivo) levantar() {
	p.sentado = false
}
