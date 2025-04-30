package mqtt

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Configurações padrão de conexão MQTT
const (
	BrokerURL = "tcp://broker.emqx.io:1883"
	Topic     = "topico/teste"
)

// Callback para quando uma mensagem é recebida
var MensagemRecebida mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Mensagem recebida no tópico: %s\n", msg.Topic())
	fmt.Printf("Conteúdo: %s\n", msg.Payload())
}

// NewMQTTClient cria e configura um novo cliente MQTT
func NewMQTTClient(clientID string, handler mqtt.MessageHandler) mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker(BrokerURL)
	opts.SetClientID(clientID)

	if handler != nil {
		opts.SetDefaultPublishHandler(handler)
	}

	client := mqtt.NewClient(opts)
	return client
}

// ConnectMQTT conecta o cliente ao broker MQTT
func ConnectMQTT(client mqtt.Client) error {
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
