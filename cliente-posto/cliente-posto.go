package main

import (
	"gopbl-2/modelo"

	"fmt"
	"time"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var posto_criado modelo.Posto
var client mqtt.Client

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")

	cadastrarPosto()
	opts.SetClientID(posto_criado.ID)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao conectar ao broker:", token.Error())
		return
	}

	fmt.Printf("Conectado ao broker MQTT em %s\n", "tcp://localhost:1883")

	// publica atualização
	for {
		payload, _ := json.Marshal(posto_criado)
		token := client.Publish(modelo.TopicCadastrarPosto, 1, false, payload)
		token.Wait()
		time.Sleep(5 * time.Second)
	}
}

func cadastrarPosto() {
	var id string
	var lat float64
	var long float64
	var servidor string

	fmt.Println("Cadastro do Posto")
	fmt.Print("Digite o ID do posto: ")
	fmt.Scanln(&id)
	fmt.Print("Digite a latitude do posto: ")
	fmt.Scanln(&lat)
	fmt.Print("Digite a longitude do posto: ")
	fmt.Scanln(&long)
	fmt.Print("Digite o nome do servidor de origem: ")
	fmt.Scanln(&servidor)

	posto_criado = modelo.NovoPosto(id, lat, long, servidor)
}
