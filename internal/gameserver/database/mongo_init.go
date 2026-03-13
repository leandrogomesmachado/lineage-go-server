package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func InicializarColecoesMongoJogo(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nomesExistentes, err := db.ListCollectionNames(ctx, struct{}{})
	if err != nil {
		return err
	}

	colecoesExistentes := make(map[string]bool, len(nomesExistentes))
	for _, nome := range nomesExistentes {
		colecoesExistentes[nome] = true
	}

	for _, nomeColecao := range ListarColecoesMongoJogo() {
		if colecoesExistentes[nomeColecao] {
			continue
		}
		errCreate := db.CreateCollection(ctx, nomeColecao)
		if errCreate != nil {
			return errCreate
		}
	}

	return nil
}
