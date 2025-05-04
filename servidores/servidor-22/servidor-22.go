package main

import (
	"context"
	"fmt"
	"log"
	"time"

	//"gopbl-2/modelo" // Substitua pelo caminho real do seu package

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Variável global do client
var client *mongo.Client

func main() {
	// Conectar ao MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Erro ao conectar ao MongoDB:", err)
	}

	// Testar conexão
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB indisponível:", err)
	}
	fmt.Println("✅ Conectado ao MongoDB com sucesso!")

	// Criar o servidor Gin
	router := gin.Default()

	// Registrar rota
	//router.POST("/publicar", publicarPostoHandler)

	// Rodar servidor
	router.Run(":8080")
}
