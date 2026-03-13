package database

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CharacterCreateInput struct {
	AccountName string
	CharName    string
	Race        int32
	Sex         int32
	ClassID     int32
	BaseClass   int32
	HairStyle   int32
	HairColor   int32
	Face        int32
	X           int32
	Y           int32
	Z           int32
	Level       int32
	MaxHp       int32
	CurHp       int32
	MaxMp       int32
	CurMp       int32
	MaxCp       int32
	CurCp       int32
	Exp         int64
	Sp          int32
	Title       string
	AccessLevel int32
}

func (r *CharacterRepository) CountByAccount(ctx context.Context, conta string) (int64, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return r.collection.CountDocuments(ctxTimeout, bson.M{"account_name": conta})
}

func (r *CharacterRepository) ExistsByName(ctx context.Context, nome string) (bool, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	quantidade, err := r.collection.CountDocuments(ctxTimeout, bson.M{"char_name_lower": strings.ToLower(nome)})
	if err != nil {
		return false, err
	}
	return quantidade > 0, nil
}

func (r *CharacterRepository) Create(ctx context.Context, entrada CharacterCreateInput) (*CharacterSlot, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	objID, err := r.proximoObjID(ctxTimeout)
	if err != nil {
		return nil, err
	}
	agora := time.Now().UnixMilli()
	documento := bson.M{
		"obj_id":                  objID,
		"account_name":            entrada.AccountName,
		"char_name":               entrada.CharName,
		"char_name_lower":         strings.ToLower(entrada.CharName),
		"level":                   entrada.Level,
		"maxHp":                   entrada.MaxHp,
		"curHp":                   entrada.CurHp,
		"maxMp":                   entrada.MaxMp,
		"curMp":                   entrada.CurMp,
		"maxCp":                   entrada.MaxCp,
		"curCp":                   entrada.CurCp,
		"face":                    entrada.Face,
		"hairStyle":               entrada.HairStyle,
		"hairColor":               entrada.HairColor,
		"sex":                     entrada.Sex,
		"x":                       entrada.X,
		"y":                       entrada.Y,
		"z":                       entrada.Z,
		"exp":                     entrada.Exp,
		"sp":                      entrada.Sp,
		"karma":                   int32(0),
		"pvpkills":                int32(0),
		"pkkills":                 int32(0),
		"clanid":                  int32(0),
		"race":                    entrada.Race,
		"classid":                 entrada.ClassID,
		"base_class":              entrada.BaseClass,
		"deletetime":              int64(0),
		"title":                   entrada.Title,
		"heading":                 int32(0),
		"rec_have":                int32(0),
		"rec_left":                int32(0),
		"online":                  int32(0),
		"onlinetime":              int32(0),
		"wantspeace":              int32(0),
		"isin7sdungeon":           int32(0),
		"punish_level":            int32(0),
		"punish_timer":            int64(0),
		"power_grade":             int32(0),
		"nobless":                 int32(0),
		"hero":                    int32(0),
		"subpledge":               int32(0),
		"lvl_joined_academy":      int32(0),
		"apprentice":              int32(0),
		"sponsor":                 int32(0),
		"varka_ketra_ally":        int32(0),
		"clan_join_expiry_time":   int64(0),
		"clan_create_expiry_time": int64(0),
		"death_penalty_level":     int32(0),
		"accesslevel":             entrada.AccessLevel,
		"lastAccess":              agora,
		"createdAt":               agora,
		"updatedAt":               agora,
	}
	_, err = r.collection.InsertOne(ctxTimeout, documento)
	if err != nil {
		return nil, err
	}
	return &CharacterSlot{
		ObjID:                objID,
		CharName:             entrada.CharName,
		Level:                entrada.Level,
		MaxHp:                entrada.MaxHp,
		CurHp:                entrada.CurHp,
		MaxMp:                entrada.MaxMp,
		CurMp:                entrada.CurMp,
		MaxCp:                entrada.MaxCp,
		CurCp:                entrada.CurCp,
		Face:                 entrada.Face,
		HairStyle:            entrada.HairStyle,
		HairColor:            entrada.HairColor,
		Sex:                  entrada.Sex,
		X:                    entrada.X,
		Y:                    entrada.Y,
		Z:                    entrada.Z,
		Exp:                  entrada.Exp,
		Sp:                   entrada.Sp,
		Karma:                0,
		PvpKills:             0,
		PkKills:              0,
		ClanID:               0,
		Race:                 entrada.Race,
		ClassID:              entrada.ClassID,
		BaseClass:            entrada.BaseClass,
		DeleteTime:           0,
		Title:                entrada.Title,
		Heading:              0,
		RecHave:              0,
		RecLeft:              0,
		Online:               0,
		OnlineTime:           0,
		WantsPeace:           0,
		IsIn7sDungeon:        0,
		PunishLevel:          0,
		PunishTimer:          0,
		PowerGrade:           0,
		Nobless:              0,
		Hero:                 0,
		SubPledge:            0,
		LvlJoinedAcademy:     0,
		Apprentice:           0,
		Sponsor:              0,
		VarkaKetraAlly:       0,
		ClanJoinExpiryTime:   0,
		ClanCreateExpiryTime: 0,
		DeathPenaltyLevel:    0,
		AccessLevel:          entrada.AccessLevel,
		LastAccess:           agora,
	}, nil
}

func (r *CharacterRepository) proximoObjID(ctx context.Context) (int32, error) {
	opcoes := options.FindOne().SetSort(bson.D{{Key: "obj_id", Value: -1}})
	resultado := r.collection.FindOne(ctx, bson.D{}, opcoes)
	if resultado.Err() != nil {
		if resultado.Err().Error() == "mongo: no documents in result" {
			return 1, nil
		}
		return 0, resultado.Err()
	}
	var slot CharacterSlot
	err := resultado.Decode(&slot)
	if err != nil {
		return 0, err
	}
	return slot.ObjID + 1, nil
}
