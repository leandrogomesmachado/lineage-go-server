package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CharacterRepository struct {
	collection *mongo.Collection
}

type CharacterSlot struct {
	ObjID                int32  `bson:"obj_id"`
	CharName             string `bson:"char_name"`
	Level                int32  `bson:"level"`
	MaxHp                int32  `bson:"maxHp"`
	CurHp                int32  `bson:"curHp"`
	MaxMp                int32  `bson:"maxMp"`
	CurMp                int32  `bson:"curMp"`
	MaxCp                int32  `bson:"maxCp"`
	CurCp                int32  `bson:"curCp"`
	Face                 int32  `bson:"face"`
	HairStyle            int32  `bson:"hairStyle"`
	HairColor            int32  `bson:"hairColor"`
	Sex                  int32  `bson:"sex"`
	X                    int32  `bson:"x"`
	Y                    int32  `bson:"y"`
	Z                    int32  `bson:"z"`
	Exp                  int64  `bson:"exp"`
	Sp                   int32  `bson:"sp"`
	Karma                int32  `bson:"karma"`
	PvpKills             int32  `bson:"pvpkills"`
	PkKills              int32  `bson:"pkkills"`
	ClanID               int32  `bson:"clanid"`
	Race                 int32  `bson:"race"`
	ClassID              int32  `bson:"classid"`
	BaseClass            int32  `bson:"base_class"`
	DeleteTime           int64  `bson:"deletetime"`
	Title                string `bson:"title"`
	Heading              int32  `bson:"heading"`
	RecHave              int32  `bson:"rec_have"`
	RecLeft              int32  `bson:"rec_left"`
	Online               int32  `bson:"online"`
	OnlineTime           int32  `bson:"onlinetime"`
	WantsPeace           int32  `bson:"wantspeace"`
	IsIn7sDungeon        int32  `bson:"isin7sdungeon"`
	PunishLevel          int32  `bson:"punish_level"`
	PunishTimer          int64  `bson:"punish_timer"`
	PowerGrade           int32  `bson:"power_grade"`
	Nobless              int32  `bson:"nobless"`
	Hero                 int32  `bson:"hero"`
	SubPledge            int32  `bson:"subpledge"`
	LvlJoinedAcademy     int32  `bson:"lvl_joined_academy"`
	Apprentice           int32  `bson:"apprentice"`
	Sponsor              int32  `bson:"sponsor"`
	VarkaKetraAlly       int32  `bson:"varka_ketra_ally"`
	ClanJoinExpiryTime   int64  `bson:"clan_join_expiry_time"`
	ClanCreateExpiryTime int64  `bson:"clan_create_expiry_time"`
	DeathPenaltyLevel    int32  `bson:"death_penalty_level"`
	AccessLevel          int32  `bson:"accesslevel"`
	LastAccess           int64  `bson:"lastAccess"`
}

func NewCharacterRepository(db *mongo.Database) *CharacterRepository {
	return &CharacterRepository{collection: db.Collection("characters")}
}

func (r *CharacterRepository) FindByAccount(ctx context.Context, conta string) ([]CharacterSlot, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := r.RemoverExpiradosDaConta(ctxTimeout, conta)
	if err != nil {
		return nil, err
	}
	filtro := bson.M{"account_name": conta}
	opcoes := options.Find().SetSort(bson.D{{Key: "lastAccess", Value: -1}, {Key: "obj_id", Value: 1}})
	cursor, err := r.collection.Find(ctxTimeout, filtro, opcoes)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxTimeout)
	resultado := make([]CharacterSlot, 0)
	for cursor.Next(ctxTimeout) {
		var slot CharacterSlot
		errDecode := cursor.Decode(&slot)
		if errDecode != nil {
			return nil, errDecode
		}
		resultado = append(resultado, slot)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return resultado, nil
}

func (r *CharacterRepository) MarcarParaExcluir(ctx context.Context, objID int32, deleteTime int64) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	filtro := bson.M{"obj_id": objID}
	atualizacao := bson.M{
		"$set": bson.M{
			"deletetime": deleteTime,
			"updatedAt":  time.Now().UnixMilli(),
		},
	}
	resultado, err := r.collection.UpdateOne(ctxTimeout, filtro, atualizacao)
	if err != nil {
		return err
	}
	if resultado.MatchedCount > 0 {
		return nil
	}
	return mongo.ErrNoDocuments
}

func (r *CharacterRepository) RestaurarExclusao(ctx context.Context, objID int32) error {
	return r.MarcarParaExcluir(ctx, objID, 0)
}

func (r *CharacterRepository) DeletarPorObjID(ctx context.Context, objID int32) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	resultado, err := r.collection.DeleteOne(ctxTimeout, bson.M{"obj_id": objID})
	if err != nil {
		return err
	}
	if resultado.DeletedCount > 0 {
		return nil
	}
	return mongo.ErrNoDocuments
}

func (r *CharacterRepository) RemoverExpiradosDaConta(ctx context.Context, conta string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	agora := time.Now().UnixMilli()
	filtro := bson.M{
		"account_name": conta,
		"deletetime":   bson.M{"$gt": 0, "$lte": agora},
	}
	_, err := r.collection.DeleteMany(ctxTimeout, filtro)
	return err
}

func (r *CharacterRepository) AtualizarPosicao(ctx context.Context, objID int32, x int32, y int32, z int32) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	filtro := bson.M{"obj_id": objID}
	atualizacao := bson.M{
		"$set": bson.M{
			"x":         x,
			"y":         y,
			"z":         z,
			"updatedAt": time.Now().UnixMilli(),
		},
	}
	resultado, err := r.collection.UpdateOne(ctxTimeout, filtro, atualizacao)
	if err != nil {
		return err
	}
	if resultado.MatchedCount > 0 {
		return nil
	}
	return mongo.ErrNoDocuments
}

func (r *CharacterRepository) AtualizarLastAccess(ctx context.Context, objID int32) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	filtro := bson.M{"obj_id": objID}
	atualizacao := bson.M{
		"$set": bson.M{
			"lastAccess": time.Now().UnixMilli(),
			"updatedAt":  time.Now().UnixMilli(),
		},
	}
	resultado, err := r.collection.UpdateOne(ctxTimeout, filtro, atualizacao)
	if err != nil {
		return err
	}
	if resultado.MatchedCount > 0 {
		return nil
	}
	return mongo.ErrNoDocuments
}
