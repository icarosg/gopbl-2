package main

import (
	"gopbl-2/db"
	"gopbl-2/modelo"

	//"gopbl-2/models"

	"context"
	//"encoding/json"
	"fmt"
	"log"
	"net/http"

	//"sync"
	"time"

	//mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// conecta ao mongo
	if erro := db.ConectarMongo(); erro != nil {
		log.Fatal("erroo ao conectar ao MongoDB:", erro)
	}
	fmt.Println("Conectado ao MongoDB com sucesso!")

	router := gin.Default() // cria o servidor Gin

	// registra rota
	router.GET("/postosDisponiveis", postosDisponiveisHandler)
	router.POST("/cadastrar", cadastrarPostoHandler)
	router.PUT("/reservar", editarPostoHandler)

	router.Run(":8082")
}

func postosDisponiveisHandler(c *gin.Context) {
	collection := db.PostosCollection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"disponivel": true} // filtro para pegar apenas os postos disponiveis

	cursor, erro := collection.Find(ctx, filter)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erroo ao buscar postos disponíveis"})
		return
	}
	defer cursor.Close(ctx)

	var postosDisponiveis []modelo.Posto
	if erro = cursor.All(ctx, &postosDisponiveis); erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao decodificar os dados"})
		return
	}

	c.JSON(http.StatusOK, postosDisponiveis)
}

func cadastrarPostoHandler(c *gin.Context) {
	var novoPosto modelo.Posto

	if erro := c.ShouldBindJSON(&novoPosto); erro != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	collection := db.PostosCollection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// verifica se já existe um posto com mesmo nome
	filter := bson.M{"ID": novoPosto.ID}
	var existente modelo.Posto
	erro := collection.FindOne(ctx, filter).Decode(&existente)
	if erro == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Posto já existente"})
		return
	}

	// se não encontrar, insere
	_, erro = collection.InsertOne(ctx, novoPosto)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao inserir o posto"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Posto cadastrado com sucesso"})
}

func editarPostoHandler(c *gin.Context) {
	var idsParaAtualizar []string

	if erro := c.ShouldBindJSON(&idsParaAtualizar); erro != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido."})
		return
	}

	fmt.Println("test", idsParaAtualizar)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// verifica se todos os postos estão disponíveis
	filter := bson.M{"id": bson.M{"$in": idsParaAtualizar}, "disponivel": true} //$in retorna os documentos cujo ID esteja dentro da lista fornecida.
	cont, err := db.PostosCollection.CountDocuments(ctx, filter)
	if err != nil || cont != int64(len(idsParaAtualizar)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nem todos os postos estão disponíveis"})
		return
	}

	// atualiza todos
	update := bson.M{"$set": bson.M{"disponivel": false}}
	
	_, err = db.PostosCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar os postos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Postos atualizados com sucesso"})
}
