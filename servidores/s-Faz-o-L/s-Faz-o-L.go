package main

import (
	"gopbl-2/db"
	"gopbl-2/modelo"
	"gopbl-2/models"

	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"
)

type Posto struct {
	ID         string  `json:"id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Disponivel bool    `json:"disponivel"`
}

var (
	postos = make(map[string]*Posto)
	mutex  = sync.RWMutex{}
)

var clienteDB *mongo.Client

// recebe mensagem MQTT do posto e atualiza o mapa
var mqttHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var p Posto
	err := json.Unmarshal(msg.Payload(), &p)
	if err != nil {
		fmt.Println("Erro ao decodificar payload:", err)
		return
	}
	mutex.Lock()
	postos[p.ID] = &p
	mutex.Unlock()
	fmt.Println("Posto atualizado via MQTT:", p.ID)
}

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
	router.POST("/publicar", publicarPostoHandler)

	// Rodar servidor
	router.Run(":8080")
}

func cadastrarPostoHandler(c *gin.Context) {
	var novoPosto modelo.Posto

	if erro := c.ShouldBindJSON(&novoPosto); erro != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	collection := clienteDB.Database("reservasRedes").Collection("postos")
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao inserir o posto"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Posto publicado com sucesso"})
}
