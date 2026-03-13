package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CharacterHenna struct {
	CharObjID  int32 `bson:"char_obj_id"`
	SymbolID   int32 `bson:"symbol_id"`
	Slot       int32 `bson:"slot"`
	ClassIndex int32 `bson:"class_index"`
}

type CharacterSkill struct {
	CharObjID  int32 `bson:"char_obj_id"`
	SkillID    int32 `bson:"skill_id"`
	SkillLevel int32 `bson:"skill_level"`
	ClassIndex int32 `bson:"class_index"`
}

type CharacterShortcut struct {
	CharObjID  int32  `bson:"char_obj_id"`
	Slot       int32  `bson:"slot"`
	Page       int32  `bson:"page"`
	Type       string `bson:"type"`
	ID         int32  `bson:"id"`
	Level      int32  `bson:"level"`
	ClassIndex int32  `bson:"class_index"`
}

type CharacterSubclass struct {
	CharObjID  int32 `bson:"char_obj_id"`
	ClassID    int32 `bson:"class_id"`
	Exp        int64 `bson:"exp"`
	Sp         int32 `bson:"sp"`
	Level      int32 `bson:"level"`
	ClassIndex int32 `bson:"class_index"`
}

type CharacterItem struct {
	OwnerID      int32  `bson:"owner_id"`
	ObjectID     int32  `bson:"object_id"`
	ItemID       int32  `bson:"item_id"`
	Count        int64  `bson:"count"`
	EnchantLevel int32  `bson:"enchant_level"`
	Loc          string `bson:"loc"`
	LocData      int32  `bson:"loc_data"`
	CustomType1  int32  `bson:"custom_type1"`
	CustomType2  int32  `bson:"custom_type2"`
	ManaLeft     int32  `bson:"mana_left"`
	Time         int64  `bson:"time"`
}

type CharacterHennaRepository struct {
	collection *mongo.Collection
}

type CharacterSkillRepository struct {
	collection *mongo.Collection
}

type CharacterShortcutRepository struct {
	collection *mongo.Collection
}

type CharacterSubclassRepository struct {
	collection *mongo.Collection
}

type CharacterItemRepository struct {
	collection *mongo.Collection
}

type CharacterDataRepositories struct {
	Characters          *CharacterRepository
	CharacterHennas     *CharacterHennaRepository
	CharacterSkills     *CharacterSkillRepository
	CharacterShortcuts  *CharacterShortcutRepository
	CharacterSubclasses *CharacterSubclassRepository
	CharacterItems      *CharacterItemRepository
}

func NewCharacterDataRepositories(db *mongo.Database) *CharacterDataRepositories {
	return &CharacterDataRepositories{
		Characters:          NewCharacterRepository(db),
		CharacterHennas:     NewCharacterHennaRepository(db),
		CharacterSkills:     NewCharacterSkillRepository(db),
		CharacterShortcuts:  NewCharacterShortcutRepository(db),
		CharacterSubclasses: NewCharacterSubclassRepository(db),
		CharacterItems:      NewCharacterItemRepository(db),
	}
}

func NewCharacterHennaRepository(db *mongo.Database) *CharacterHennaRepository {
	return &CharacterHennaRepository{collection: db.Collection("characterHennas")}
}

func NewCharacterSkillRepository(db *mongo.Database) *CharacterSkillRepository {
	return &CharacterSkillRepository{collection: db.Collection("characterSkills")}
}

func NewCharacterShortcutRepository(db *mongo.Database) *CharacterShortcutRepository {
	return &CharacterShortcutRepository{collection: db.Collection("characterShortcuts")}
}

func NewCharacterSubclassRepository(db *mongo.Database) *CharacterSubclassRepository {
	return &CharacterSubclassRepository{collection: db.Collection("characterSubclasses")}
}

func NewCharacterItemRepository(db *mongo.Database) *CharacterItemRepository {
	return &CharacterItemRepository{collection: db.Collection("items")}
}

func (r *CharacterHennaRepository) ListarPorPersonagem(ctx context.Context, objID int32, classIndex int32) ([]CharacterHenna, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cursor, err := r.collection.Find(ctxTimeout, bson.M{"char_obj_id": objID, "class_index": classIndex})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxTimeout)
	resultado := make([]CharacterHenna, 0)
	for cursor.Next(ctxTimeout) {
		var item CharacterHenna
		errDecode := cursor.Decode(&item)
		if errDecode != nil {
			return nil, errDecode
		}
		resultado = append(resultado, item)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return resultado, nil
}

func (r *CharacterItemRepository) ListarPorPersonagem(ctx context.Context, objID int32) ([]CharacterItem, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cursor, err := r.collection.Find(ctxTimeout, bson.M{"owner_id": objID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxTimeout)
	resultado := make([]CharacterItem, 0)
	for cursor.Next(ctxTimeout) {
		var item CharacterItem
		errDecode := cursor.Decode(&item)
		if errDecode != nil {
			return nil, errDecode
		}
		resultado = append(resultado, item)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return resultado, nil
}

func (r *CharacterSkillRepository) ListarPorPersonagem(ctx context.Context, objID int32, classIndex int32) ([]CharacterSkill, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cursor, err := r.collection.Find(ctxTimeout, bson.M{"char_obj_id": objID, "class_index": classIndex})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxTimeout)
	resultado := make([]CharacterSkill, 0)
	for cursor.Next(ctxTimeout) {
		var item CharacterSkill
		errDecode := cursor.Decode(&item)
		if errDecode != nil {
			return nil, errDecode
		}
		resultado = append(resultado, item)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return resultado, nil
}

func (r *CharacterShortcutRepository) ListarPorPersonagem(ctx context.Context, objID int32, classIndex int32) ([]CharacterShortcut, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cursor, err := r.collection.Find(ctxTimeout, bson.M{"char_obj_id": objID, "class_index": classIndex})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxTimeout)
	resultado := make([]CharacterShortcut, 0)
	for cursor.Next(ctxTimeout) {
		var item CharacterShortcut
		errDecode := cursor.Decode(&item)
		if errDecode != nil {
			return nil, errDecode
		}
		resultado = append(resultado, item)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return resultado, nil
}

func (r *CharacterSubclassRepository) ListarPorPersonagem(ctx context.Context, objID int32) ([]CharacterSubclass, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cursor, err := r.collection.Find(ctxTimeout, bson.M{"char_obj_id": objID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxTimeout)
	resultado := make([]CharacterSubclass, 0)
	for cursor.Next(ctxTimeout) {
		var item CharacterSubclass
		errDecode := cursor.Decode(&item)
		if errDecode != nil {
			return nil, errDecode
		}
		resultado = append(resultado, item)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return resultado, nil
}
