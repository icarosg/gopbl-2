package main

import (
	"gopbl-2/modelo"

	"fmt"
	"time"

	"bytes"
	"encoding/json"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// var (
// 	id        string
// 	latitude  float64
// 	longitude float64
// )

var posto_criado modelo.Posto

func main() {
	endereco := selecionarServidorManual()

	opts := mqtt.NewClientOptions().AddBroker(endereco)

	cadastrarPosto()
	opts.SetClientID(posto_criado.ID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao conectar ao broker:", token.Error())
		return
	}

	fmt.Printf("Conectado ao broker MQTT em %s\n", endereco)

	// publica atualização
	for {
		payload, _ := json.Marshal(posto_criado)
		token := client.Publish("postos/"+posto_criado.ID, 0, false, payload)
		token.Wait()
		fmt.Println("Publicado estado do posto.")
		time.Sleep(5 * time.Second)
	}
}

func cadastrarPosto() {
	// fmt.Println("Cadastro do Posto")
	// fmt.Print("Digite o ID do posto: ")
	// fmt.Scanln(&id)
	// fmt.Print("Digite a latitude do posto: ")
	// fmt.Scanln(&latitude)
	// fmt.Print("Digite a longitude do posto: ")
	// fmt.Scanln(&longitude)

	posto_criado = modelo.NovoPosto("posto1", 15, 20)
	onSubmit(posto_criado);
}

func selecionarServidorManual() string {
	var ip string
	var porta string

	fmt.Println("Conexão com servidor MQTT")
	fmt.Print("Digite o IP do servidor (ex: 127.0.0.1): ")
	fmt.Scanln(&ip)
	fmt.Print("Digite a porta (ex: 1883): ")
	fmt.Scanln(&porta)

	return fmt.Sprintf("tcp://%s:%s", ip, porta)
}

func onSubmit(posto modelo.Posto) {
	postData, err := json.Marshal(posto) //converte para json
	if err != nil {
		fmt.Println("Erro ao codificar JSON:", err)
		return
	}

	resp, err := http.Post("http://localhost:8080/cadastrar", "application/json", bytes.NewBuffer(postData))
	if err != nil {
		fmt.Println("Erro ao enviar requisição:", err)
		return
	}
	resp.Body.Close()

	fmt.Println("Status:", resp.Status)
}
