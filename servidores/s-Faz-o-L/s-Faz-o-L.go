package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

// Dados do posto
type Posto struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Disponivel bool   `json:"disponivel"`
}

var (
	postos = make(map[string]*Posto)
	mutex  = sync.RWMutex{}
)

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
	// conectar ao broker MQTT
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("servidor-api-1")
	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("Conectado ao broker MQTT")
		if token := c.Subscribe("postos/+", 0, mqttHandler); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// API REST
	router := gin.Default()

	// GET - listar postos disponíveis
	router.GET("/postos", func(c *gin.Context) {
		mutex.RLock()
		defer mutex.RUnlock()
		lista := []*Posto{}
		for _, p := range postos {
			if p.Disponivel {
				lista = append(lista, p)
			}
		}
		c.JSON(http.StatusOK, lista)
	})

	// POST - reservar posto (disponível -> falso)
	router.POST("/reservar/:id", func(c *gin.Context) {
		id := c.Param("id")
		mutex.Lock()
		defer mutex.Unlock()
		if posto, ok := postos[id]; ok && posto.Disponivel {
			posto.Disponivel = false
			c.JSON(http.StatusOK, gin.H{"message": "Posto reservado com sucesso"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Posto não encontrado ou já reservado"})
		}
	})

	router.Run(":8082")
}
