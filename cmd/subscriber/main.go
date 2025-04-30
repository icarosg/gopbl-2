package main

import (
	"fmt"
	"os"
	"time"

	"gopbl-2/pkg/mqtt"
)

func main() {
	fmt.Println("Iniciando subscriber MQTT...")

	// Criar e conectar cliente MQTT
	client := mqtt.NewMQTTClient("go-mqtt-subscriber", mqtt.MensagemRecebida)
	if err := mqtt.ConnectMQTT(client); err != nil {
		fmt.Println("Erro ao conectar:", err)
		os.Exit(1)
	}

	// Assina o tópico
	if token := client.Subscribe(mqtt.Topic, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao assinar:", token.Error())
		os.Exit(1)
	}

	fmt.Println("Aguardando mensagens...")
	fmt.Printf("Inscrito no tópico: %s\n", mqtt.Topic)

	// Mantém o programa em execução por 30 segundos
	time.Sleep(30 * time.Second)

	// Cancela a assinatura e desconecta
	if token := client.Unsubscribe(mqtt.Topic); token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao cancelar assinatura:", token.Error())
	}

	client.Disconnect(250)
	fmt.Println("Subscriber desconectado.")
}
