package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopbl-2/modelo"
	"net/http"
)

var veiculo modelo.Veiculo
var cadastrado bool = false
var endereco string

var client mqtt.Client
var opts *mqtt.ClientOptions

func main() {
	endereco = selecionarServidorManual() // para o http

	go menu() //conecta ao servidor após o cadastro do veículo

	select {} // mantém o cliente vivo e aguardando interações
}

func selecionarServidorManual() string {
	var ip string
	var porta string

	fmt.Println("Conexão com servidor")
	fmt.Print("Digite o IP do servidor (ex: 127.0.0.1): ")
	fmt.Scanln(&ip)
	fmt.Print("Digite a porta (ex: 8080): ")
	fmt.Scanln(&porta)

	return fmt.Sprintf("http://%s:%s", ip, porta)
}

func conectarAoBroker() {
	if veiculo.ID != "" {

		opts = mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")

		opts.SetClientID(veiculo.ID)

		client = mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Println("Erro ao conectar ao broker:", token.Error())
			return
		}

		fmt.Printf("Conectado ao broker MQTT em tcp://localhost:1883\n")
	}
}

func publicarReserva(p modelo.Posto) {
	if veiculo.ID != "" {
		payload, _ := json.Marshal(p)
		token := client.Publish("reserva/"+veiculo.ID, 0, false, payload)
		token.Wait()
		fmt.Println("Publicado a possível reserva do posto.")
	}
}

func listarPostos() []modelo.Posto {
	resp, err := http.Get(endereco + "/postos")
	if err != nil {
		fmt.Println("Erro ao consultar:", err)
		return []modelo.Posto{}
	}
	resp.Body.Close()

	var postos []modelo.Posto
	err = json.NewDecoder(resp.Body).Decode(&postos)
	if err != nil {
		fmt.Println("Erro ao decodificar resposta:", err)
		return []modelo.Posto{}
	}

	fmt.Printf("\n\nPostos disponíveis:\n")
	for _, p := range postos {
		fmt.Printf("- %s (%f, %f)\n", p.ID, p.Latitude, p.Longitude)
	}

	return postos
}

func procurarPostosParaReserva(postos []modelo.Posto) {
	var reserva string

	fmt.Println("Digite o ID dos postos que deseja reservar, em rodem: ")
	fmt.Println("Digite 1, caso deseje sair")
	for {
		fmt.Scanln(&reserva)

		if reserva == "1" || reserva == "" {
			return
		}

		for _, p := range postos {
			if (p.ID == reserva) {
				publicarReserva(p)
				return
			}
		}

	}
}

func menu() {
	for {
		fmt.Println("\nMenu de Ações:")
		fmt.Println("1 - Cadastrar veículo")
		fmt.Println("2 - Atualizar posição do veículo")
		fmt.Println("3 - Consultar postos disponíveis")
		fmt.Println("4 - Reservar posto")
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

			conectarAoBroker() //----------------------------------------------------------------------------------------------------

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
			listarPostos()
		case 4:
			postos := listarPostos()

			if len(postos) > 0 {
				procurarPostosParaReserva(postos)
			} else {
				fmt.Println("Postos não encontrado.")
			}
		default:
			fmt.Println("Opção inválida.")
		}
	}
}
