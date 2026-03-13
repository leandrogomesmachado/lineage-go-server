package database

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrContaJaExiste = errors.New("conta com login ja existente")

type Account struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Login       string             `bson:"login"`
	Password    string             `bson:"password"`
	AccessLevel int                `bson:"access_level"`
	LastActive  time.Time          `bson:"last_active"`
	LastIP      string             `bson:"last_ip"`
	LastServer  int                `bson:"last_server"`
	BannedUntil *time.Time         `bson:"banned_until,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
}

type AccountRepository struct {
	collection *mongo.Collection
}

func NewAccountRepository(db *mongo.Database) *AccountRepository {
	return &AccountRepository{
		collection: db.Collection("accounts"),
	}
}

func (r *AccountRepository) FindByLogin(ctx context.Context, login string) (*Account, error) {
	var account Account
	err := r.collection.FindOne(ctx, bson.M{"login": login}).Decode(&account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) Create(ctx context.Context, account *Account) error {
	account.CreatedAt = time.Now()
	documentoInsercao := bson.M{
		"login":        account.Login,
		"password":     account.Password,
		"access_level": account.AccessLevel,
		"last_active":  account.LastActive,
		"last_ip":      account.LastIP,
		"last_server":  account.LastServer,
		"created_at":   account.CreatedAt,
	}
	if account.BannedUntil != nil {
		documentoInsercao["banned_until"] = account.BannedUntil
	}
	resultado, err := r.collection.UpdateOne(
		ctx,
		bson.M{"login": account.Login},
		bson.M{
			"$setOnInsert": documentoInsercao,
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	if resultado.UpsertedCount == 0 {
		return ErrContaJaExiste
	}
	id, ok := resultado.UpsertedID.(primitive.ObjectID)
	if ok {
		account.ID = id
	}
	return nil
}

func (r *AccountRepository) UpdateLastLogin(ctx context.Context, login, ip string, serverId int) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"login": login},
		bson.M{
			"$set": bson.M{
				"last_active": time.Now(),
				"last_ip":     ip,
				"last_server": serverId,
			},
		},
	)
	return err
}

func (r *AccountRepository) BanAccount(ctx context.Context, login string, duration time.Duration) error {
	bannedUntil := time.Now().Add(duration)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"login": login},
		bson.M{
			"$set": bson.M{
				"banned_until": bannedUntil,
			},
		},
	)
	return err
}

func (r *AccountRepository) IsBanned(ctx context.Context, login string) (bool, error) {
	account, err := r.FindByLogin(ctx, login)
	if err != nil {
		return false, err
	}

	if account.BannedUntil != nil && account.BannedUntil.After(time.Now()) {
		return true, nil
	}

	return false, nil
}
