package main

import (
	"fmt"
	"gopbl-2/modelo"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var veiculo modelo.Veiculo
var cadastrado bool = false
var client mqtt.Client

func main() {
	go menu() //conecta ao servidor após o cadastro do veículo

	select {} // mantém o cliente vivo e aguardando interações
}

// Função para selecionar o servidor MQTT manualmente
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

func menu() {
	for {
		fmt.Println("\nMenu de Ações:")
		fmt.Println("1 - Cadastrar veículo")
		fmt.Println("2 - Atualizar posição do veículo")
		fmt.Println("3 - Inscrição em tópicos (Para veículo)")
		var opcao int
		fmt.Scanf("%d", &opcao)

		switch opcao {
		case 1:
			fmt.Println("Digite o ID do veículo:")
			var id string
			fmt.Scanf("%s", &id)
			fmt.Println("Digite a Latitude do veículo:")
			var lat float64
			fmt.Scanf("%f", &lat)
			fmt.Println("Digite a Longitude do veículo:")
			var long float64
			fmt.Scanf("%f", &long)
			veiculo = modelo.NovoVeiculo(id, lat, long)
			cadastrado = true
			fmt.Println("Veículo cadastrado:", veiculo)



			// conecta ao servidor MQTT apenas após o cadastro do veículo
			endereco := selecionarServidorManual()

			opts := mqtt.NewClientOptions().AddBroker(endereco)
			opts.SetClientID(veiculo.ID)

			client = mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				fmt.Println("Erro ao conectar ao broker:", token.Error())
				return
			}

			fmt.Printf("Conectado ao broker MQTT em %s com clientID: %s\n", endereco, veiculo.ID)

		case 2:
			if !cadastrado {
				fmt.Println("Veículo não cadastrado!")
				continue
			}
			fmt.Println("Digite a nova Latitude do veículo:")
			var lat float64
			fmt.Scanf("%f", &lat)
			fmt.Println("Digite a nova Longitude do veículo:")
			var long float64
			fmt.Scanf("%f", &long)
			veiculo.Latitude = lat
			veiculo.Longitude = long
			fmt.Println("Veículo atualizado:", veiculo)

		case 3:
			// inscrição em tópicos MQTT
			if !cadastrado {
				fmt.Println("Você precisa cadastrar o veículo primeiro.")
				continue
			}
			inscreverEmTopicos(client) // colocar para listar os postos

		default:
			fmt.Println("Opção inválida.")
		}
	}
}

func inscreverEmTopicos(client mqtt.Client) {
	fmt.Println("Inscrevendo em tópicos...")
	// exemplificando inscrição em tópicos
	tópicos := []string{"veiculos/posicao", "veiculos/status"}
	for _, topico := range tópicos {
		token := client.Subscribe(topico, 0, func(client mqtt.Client, msg mqtt.Message) {
			fmt.Printf("Mensagem recebida no tópico %s: %s\n", msg.Topic(), string(msg.Payload()))
		})
		token.Wait()
		fmt.Printf("Inscrito no tópico: %s\n", topico)
	}
}
