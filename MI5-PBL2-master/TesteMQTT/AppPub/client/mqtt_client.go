package client

import (
	"encoding/json"
	"fmt"
	"log"
	"main/global"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTT struct {
	mqttClient mqtt.Client
	mqtt.MessageHandler
	mqtt.ConnectionLostHandler
	mqtt.OnConnectHandler
	mqtt.ConnectionAttemptHandler
	mqtt.ReconnectHandler
	logger *log.Logger
}

// NewMQTTClient cria um novo cliente MQTT
func NewMQTTClient(port int, broker string) (*MQTT, error) {
	// Configurações do cliente MQTT
	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	client := mqtt.NewClient(options)

	// Autenticação do token
	token := client.Connect()
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	return &MQTT{
		mqttClient: client,
	}, nil
}

// Publica uma mensagem em um tópico
func (mq *MQTT) Publish(message global.MQTT_Message) error {

	buffer, err := json.Marshal(message)
	if err != nil {
		err = fmt.Errorf("error marshaling message: %v", err)
		return err
	}
	token := mq.mqttClient.Publish(message.Topic.String(), 0, false, buffer)
	token.Wait()
	err = token.Error()
	if err != nil {
		return err
	}
	return nil
}

// Se inscreve em um tópico
func (mq *MQTT) Subscribe(topic global.Topics, callback mqtt.MessageHandler) {
	mq.mqttClient.Subscribe(topic.String(), 0, callback)
}
