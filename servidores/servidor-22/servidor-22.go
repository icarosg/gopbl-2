package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	//"net/http"
	// "os"
	//"sync"
	"time"
	//"gopbl-2/modelo"
	// "io"
	// "log"
	// "math"
	// "net"
	"encoding/json"
)

type PostoJson struct {
	ID              string  `json:"id"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	QuantidadeFila  int     `json:"quantidade de carros na fila"`
	Disponibilidade bool    `json:"bomba disponivel"`
}

type PagamentoJson struct {
	ID_veiculo string  `json:"id_veiculo"`
	Valor      float64 `json:"valor"`
	Posto      string  `json:"id_posto"`
}

type Requisicao struct {
	Comando string          `json:"comando"`
	Dados   json.RawMessage `json:"dados"`
}

var messagePubHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

// Variável global para o cliente MQTT
var mqttClient MQTT.Client

func main() {

	setupMQTT()
	defer mqttClient.Disconnect(250)

	gin.DisableBindValidation()
	gin.SetMode(gin.ReleaseMode)
	rota := gin.Default()
	// rota.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "pong",
	// 	})
	// })	
	rota.Run("localhost:8080")	
	fmt.Println("Servidor iniciado e conectado ao MQTT Broker")

}

func subscribeToTopics() {
	// Exemplo de subscription
	topic := "topic/receba"
	//token := mqttClient.Subscribe(topic, 1, nil)
	
	token := mqttClient.Subscribe(topic, 1, nil)

	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

func publishMessage(topic string, message string) {
	token := mqttClient.Publish(topic, 0, false, message)
	token.Wait()
	time.Sleep(time.Second)
}

func setupMQTT() {
	var broker = "localhost"
	var port = 1883
	opts := MQTT.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_server")
	opts.SetUsername("emqx")
	opts.SetPassword("public")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	mqttClient = MQTT.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Subscrever aos tópicos necessários
	subscribeToTopics()

	go func(){
		for {
			fmt.Println("Menu")
			fmt.Println("1 - enviar mensagem pro cliente")
			var opcao int
			fmt.Scanln(&opcao)
			switch opcao {
			case 1:
				fmt.Println("Digite a mensagem")
				var mensagem string
				fmt.Scanln(&mensagem)
				mqttClient.Publish("topic/testar", 0, false, mensagem)
			}
		}
		

	}()

}