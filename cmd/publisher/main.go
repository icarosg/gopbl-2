package main

import (
	"fmt"
	"os"
	"time"

	"gopbl-2/pkg/mqtt"
)

func main() {
	fmt.Println("Iniciando publisher MQTT...")

	// Criar e conectar cliente MQTT
	client := mqtt.NewMQTTClient("go-mqtt-publisher", nil)
	if err := mqtt.ConnectMQTT(client); err != nil {
		fmt.Println("Erro ao conectar:", err)
		os.Exit(1)
	}

	// Publica mensagens a cada 2 segundos
	for i := 0; i < 5; i++ {
		mensagem := fmt.Sprintf("Mensagem %d do servidor 1", i)
		token := client.Publish(mqtt.Topic, 0, false, mensagem)
		token.Wait()
		fmt.Printf("Publicado: %s\n", mensagem)
		time.Sleep(2 * time.Second)
	}

	client.Disconnect(250)
	fmt.Println("Publisher desconectado.")
}
