package database

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var Database *mongo.Database

// ConnectDB abre a conexão uma vez
func ConnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("MONGO_URI") // ex: "mongodb://mongodb:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	MongoClient = client

	// Obtém o nome do banco de dados da variável de ambiente
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		panic("DB_NAME não definido")
	}

	// Cria uma referência ao banco de dados
	Database = MongoClient.Database(dbName)
}

// DisconnectDB fecha a conexão
func DisconnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	MongoClient.Disconnect(ctx)
}
