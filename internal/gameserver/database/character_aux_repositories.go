package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type CharacterAugmentation struct {
	ItemOID    int32 `bson:"item_oid"`
	Attributes int32 `bson:"attributes"`
	SkillID    int32 `bson:"skill_id"`
	SkillLevel int32 `bson:"skill_level"`
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

type CharacterAugmentationRepository struct {
	collection *mongo.Collection
}

type CharacterDataRepositories struct {
	Characters          *CharacterRepository
	CharacterHennas     *CharacterHennaRepository
	CharacterSkills     *CharacterSkillRepository
	CharacterShortcuts  *CharacterShortcutRepository
	CharacterSubclasses *CharacterSubclassRepository
	CharacterItems      *CharacterItemRepository
	CharacterAugments   *CharacterAugmentationRepository
}

func NewCharacterDataRepositories(db *mongo.Database) *CharacterDataRepositories {
	return &CharacterDataRepositories{
		Characters:          NewCharacterRepository(db),
		CharacterHennas:     NewCharacterHennaRepository(db),
		CharacterSkills:     NewCharacterSkillRepository(db),
		CharacterShortcuts:  NewCharacterShortcutRepository(db),
		CharacterSubclasses: NewCharacterSubclassRepository(db),
		CharacterItems:      NewCharacterItemRepository(db),
		CharacterAugments:   NewCharacterAugmentationRepository(db),
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

func NewCharacterAugmentationRepository(db *mongo.Database) *CharacterAugmentationRepository {
	return &CharacterAugmentationRepository{collection: db.Collection("augmentations")}
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

func (r *CharacterItemRepository) InserirOuSomarItem(ctx context.Context, ownerID int32, itemID int32, quantidade int64) (*CharacterItem, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if quantidade <= 0 {
		return nil, nil
	}
	filtroExistente := bson.M{"owner_id": ownerID, "item_id": itemID, "loc": "INVENTORY"}
	var existente CharacterItem
	err := r.collection.FindOne(ctxTimeout, filtroExistente).Decode(&existente)
	if err == nil {
		novoTotal := existente.Count + quantidade
		_, errUpdate := r.collection.UpdateOne(ctxTimeout, bson.M{"object_id": existente.ObjectID}, bson.M{"$set": bson.M{"count": novoTotal}})
		if errUpdate != nil {
			return nil, errUpdate
		}
		existente.Count = novoTotal
		return &existente, nil
	}
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	objID, errObjID := r.proximoObjectID(ctxTimeout)
	if errObjID != nil {
		return nil, errObjID
	}
	novo := CharacterItem{OwnerID: ownerID, ObjectID: objID, ItemID: itemID, Count: quantidade, EnchantLevel: 0, Loc: "INVENTORY", LocData: 0, CustomType1: 0, CustomType2: 0, ManaLeft: -1, Time: 0}
	_, errInsert := r.collection.InsertOne(ctxTimeout, novo)
	if errInsert != nil {
		return nil, errInsert
	}
	return &novo, nil
}

func (r *CharacterItemRepository) proximoObjectID(ctx context.Context) (int32, error) {
	opcoes := options.FindOne().SetSort(bson.D{{Key: "object_id", Value: -1}})
	resultado := r.collection.FindOne(ctx, bson.D{}, opcoes)
	if resultado.Err() == mongo.ErrNoDocuments {
		return 1, nil
	}
	if resultado.Err() != nil {
		return 0, resultado.Err()
	}
	var item CharacterItem
	err := resultado.Decode(&item)
	if err != nil {
		return 0, err
	}
	return item.ObjectID + 1, nil
}

func (r *CharacterSkillRepository) InserirLote(ctx context.Context, skills []CharacterSkill) error {
	if len(skills) == 0 {
		return nil
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for _, skill := range skills {
		filtro := bson.M{
			"char_obj_id": skill.CharObjID,
			"class_index": skill.ClassIndex,
			"skill_id":    skill.SkillID,
		}
		_, err := r.collection.DeleteMany(ctxTimeout, filtro)
		if err != nil {
			return err
		}
		_, err = r.collection.InsertOne(ctxTimeout, skill)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CharacterAugmentationRepository) ListarPorItens(ctx context.Context, objectIDs []int32) ([]CharacterAugmentation, error) {
	if len(objectIDs) == 0 {
		return []CharacterAugmentation{}, nil
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cursor, err := r.collection.Find(ctxTimeout, bson.M{"item_oid": bson.M{"$in": objectIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxTimeout)
	resultado := make([]CharacterAugmentation, 0)
	for cursor.Next(ctxTimeout) {
		var item CharacterAugmentation
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
