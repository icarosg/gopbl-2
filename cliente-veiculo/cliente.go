package main

import (
	"fmt"
	"gopbl-2/modelo"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"bytes"
	"encoding/json"
	"net/http"
)

var veiculo modelo.Veiculo
var cadastrado bool = false
var endereco string

var idPostos []string

var client mqtt.Client
var opts *mqtt.ClientOptions

func main() {
	endereco = selecionarServidorManual() // para o http

	menu() //conecta ao servidor após o cadastro do veículo

	//select {} // mantém o cliente vivo e aguardando interações
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

func cadastrarVeiculo() {
	var id string
	var lat float64
	var long float64
	fmt.Print("Digite o ID do veículo: ")
	fmt.Scanln(&id)

	fmt.Print("Digite a Latitude do veículo: ")
	fmt.Scanln(&lat)

	fmt.Print("Digite a Longitude do veículo: ")
	fmt.Scanln(&long)

	veiculo = modelo.NovoVeiculo(id, lat, long)
	cadastrado = true
	fmt.Println("Veículo cadastrado:", veiculo)

	conectarAoBroker() //----------------------------------------------------------------------------------------------------
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

func onSubmit(idPostos []string, reservar bool) {
	putData, erro := json.Marshal(struct {
		IDPostos []string `json:"idPostos"`
		Reservar bool     `json:"reservar"`
	}{
		IDPostos: idPostos,
		Reservar: reservar,
	})
	if erro != nil {
		fmt.Println("Erro ao codificar JSON:", erro)
		return
	}

	req, erro := http.NewRequest(http.MethodPut, endereco+"/reservar", bytes.NewBuffer(putData))
	if erro != nil {
		fmt.Println("erro ao criar requisição:", erro)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, erro := client.Do(req)
	if erro != nil {
		fmt.Println("erro ao enviar requisição:", erro)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)
}

func listarPostos() []modelo.Posto {
	resp, err := http.Get(endereco + "/postosDisponiveis")
	if err != nil {
		fmt.Println("Erro ao consultar:", err)
		return []modelo.Posto{}
	}
	defer resp.Body.Close()

	var postos []modelo.Posto
	err = json.NewDecoder(resp.Body).Decode(&postos)
	if err != nil {
		fmt.Println("Erro ao decodificar resposta:", err)
		return []modelo.Posto{}
	}

	fmt.Printf("\n\nPostos disponíveis:\n")
	for _, p := range postos {
		fmt.Printf("- %s (%f, %f) - Servidor: %s\n", p.ID, p.Latitude, p.Longitude, p.ServidorOrigem)
	}

	return postos
}

func procurarPostosParaReserva(postos []modelo.Posto) {
	var reserva string

	fmt.Println("Digite 1 ou pressione Enter para sair")
	fmt.Println("Digite o ID dos postos que deseja reservar, um de cada vez: ")

	for {
		reserva = "1"
		fmt.Scanln(&reserva)

		if reserva == "1" || reserva == "" {
			break
		}

		encontrado := false
		for _, p := range postos {
			if p.ID == reserva {
				idPostos = append(idPostos, p.ID)
				encontrado = true
				fmt.Println("Posto adicionado!")
				break
			}
		}

		if !encontrado {
			fmt.Println("Posto não encontrado.")
		}
	}

	if len(idPostos) > 0 {
		fmt.Println(idPostos)
		onSubmit(idPostos, true)
	} else {
		fmt.Println("Nenhum posto selecionado.")
	}
}

func menu() {
	for {
		fmt.Println("\nMenu de Ações:")
		fmt.Println("1 - Cadastrar veículo")
		fmt.Println("2 - Atualizar posição do veículo")
		fmt.Println("3 - Consultar postos disponíveis")
		fmt.Println("4 - Reservar posto")
		fmt.Println("5 - Finalizar viagem")
		var opcao int
		fmt.Print("Escolha uma opção: ")
		fmt.Scanln(&opcao)

		switch opcao {
		case 1:
			fmt.Println("cadastrar veiculo")
			cadastrarVeiculo()
		case 2:
			if !cadastrado {
				fmt.Println("Veículo não cadastrado!")
				continue
			}
			fmt.Println("Digite a nova Latitude do veículo:")
			var lat float64
			fmt.Scanln(&lat)
			fmt.Println("Digite a nova Longitude do veículo:")
			var long float64
			fmt.Scanln(&long)
			veiculo.Latitude = lat
			veiculo.Longitude = long
			fmt.Println("Veículo atualizado:", veiculo)
		case 3:
			listarPostos()
		case 4:
			if len(idPostos) == 0 || idPostos == nil {
				postos := listarPostos()

				if len(postos) > 0 {
					procurarPostosParaReserva(postos)
				} else {
					fmt.Println("Postos não encontrado.")
				}
			} else {
				fmt.Println("Você já possui reservas em andamento. Finalize a viagem para reservar novamente.")
			}
		case 5:
			if len(idPostos) > 0 {
				onSubmit(idPostos, false)
				idPostos = []string{}
			} else {
				fmt.Println("Você não possui reservas.")
			}
		default:
			fmt.Println("Opção inválida.")
		}
	}
}
