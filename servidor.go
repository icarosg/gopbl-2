package main

import (
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"

	//"net/http"
	// "os"
	"sync"
	"gopbl-2/modelo"
	"time"

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

var mutex sync.Mutex
var postos_servidor_A = make(map[string]modelo.Posto)
var confirmacoes []bool
var postosChanA = make(chan modelo.PostoConsulta, 10) // buffer de 10


var messagePubHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	//fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	switch msg.Topic() {
	case "topic/pedido-consulta-cliente-1-A":
		token := client.Publish("topic/listar-postos", 0, false, "listar postos")		
		token.Wait()		
		
		// payload, err := json.Marshal(postos_servidor_A)
		// if err != nil {
		// 	fmt.Println("erro ao codificar o dic com os postos")
		// 	return
		// }
		// token = mqttClient.Publish("topic/consulta-cliente-1-A", 0, false, payload)
		// token.Wait()
		// fmt.Println("postos enviados para o cliente 1-A")
	case "topic/receber-posto":		
		var posto modelo.PostoConsulta
		err := json.Unmarshal(msg.Payload(), &posto)
		if err != nil{
			fmt.Println("erro ao decodificar o posto recebido")
			return
		}
		postos_servidor_A[posto.ID] = posto.Posto
		payload, err := json.Marshal(postos_servidor_A)
		if err != nil {
			fmt.Println("erro ao codificar o dic com os postos")
			return
		}
		token := mqttClient.Publish("topic/consulta-cliente-1-A", 0, false, payload)
		token.Wait()
		fmt.Println("postos enviados para o cliente 1-A")
		client.Publish("topic/confirmacao-veiculo", 0, false, "")
		//postosChanA <- posto
		fmt.Println("recebi o posto")
	case "topic/reqAtomica":
		var req modelo.ReqAtomica
		err := json.Unmarshal(msg.Payload(), &req)
		if err != nil {
			fmt.Println("erro ao decodificar a req atomica")
			return
		}
		token := mqttClient.Publish("topic/possivel-reserva", 0, false, req.Veiculo)
		token.Wait()
		var all bool = true
		for i := range confirmacoes{			
			if !confirmacoes[i]{
				all = false
				break
			}
		}
		if !all {
			fmt.Println("algum postou nao pode ser reservado")
			return
		}
		// Se todos os postos puderem ser reservados, enviar a confirmação
		token = mqttClient.Publish("topic/reservar-vaga", 0, false, req.Veiculo) 
		token.Wait()
	case "topic/possivel-reserva-server":
		var confirm bool
		err := json.Unmarshal(msg.Payload(), &confirm)
		if err != nil{
			fmt.Println("erro ao receber a possivel confirmacao")
			return
		}
		confirmacoes = append(confirmacoes, confirm)
		
	}
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
	postos_servidor_A["posto-1"] = modelo.Posto{
		ID:        "posto-1",
		Latitude:  50,
		Longitude: 50,
	}
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

	topic = "topic/pedido-consulta-cliente-1-A"
	token = mqttClient.Subscribe(topic, 1, nil)
	token.Wait()

	topic = "topic/receber-posto"
	token = mqttClient.Subscribe(topic, 1, nil)
	token.Wait()

	topic = "topic/reqAtomica"
	token = mqttClient.Subscribe(topic, 1, nil)
	token.Wait()

}

func publishMessage(topic string, message string) {
	token := mqttClient.Publish(topic, 0, false, message)
	token.Wait()
	time.Sleep(time.Second)
}

func setupMQTT() {
	var broker = "192.168.0.110"
	var port = 1884
	opts := MQTT.NewClientOptions()
	//opts.AddBroker("tcp://192.168.0.110:1883")
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

	go func() {
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


