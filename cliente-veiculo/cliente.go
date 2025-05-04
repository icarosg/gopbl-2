package main

import (
	"encoding/json"
	"fmt"
	"gopbl-2/modelo"
	"net/http"
	//mqtt "github.com/eclipse/paho.mqtt.golang"
)

var veiculo modelo.Veiculo
var cadastrado bool = false
var endereco string

func main() {
	endereco = selecionarServidorManual()

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
			resp, erro := http.Get(endereco + "/postos")
			if erro != nil {
				fmt.Println("Erro ao consultar:", erro)
				continue
			}

			var postos []modelo.Posto
			erro = json.NewDecoder(resp.Body).Decode(&postos)
			resp.Body.Close()

			if erro != nil {
				fmt.Println("Erro ao decodificar resposta:", erro)
				continue
			}

			fmt.Printf("\n\nPostos disponíveis:\n")
			for _, p := range postos {
				fmt.Printf("- %s (%f, %f)\n", p.ID, p.Latitude, p.Longitude)
			}

		default:
			fmt.Println("Opção inválida.")
		}
	}
}
