package db

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var PostosCollection *mongo.Collection

func ConectarMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clienteOpts := options.Client().ApplyURI("mongodb://localhost:27017")
	client, erro := mongo.Connect(ctx, clienteOpts)
	if erro != nil {
		return erro
	}

	PostosCollection = client.Database("reservasRedes").Collection("postos")
	return nil
}
