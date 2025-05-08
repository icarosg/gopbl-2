package main

import (
	"fmt"
	"gopbl-2/modelo"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"encoding/json"
	"sync"
	"time"
)

var veiculo modelo.Veiculo
var cadastrado bool = false
var idPostos []string // Restaurado para manter reservas
var client mqtt.Client
var opts *mqtt.ClientOptions
var responseTopic string
var listarMutex sync.Mutex // Mutex para evitar chamadas concorrentes à função listarPostos

func main() {
	menu()
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

	conectarAoBroker()
}

func conectarAoBroker() {
	if veiculo.ID != "" {
		opts = mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
		opts.SetClientID(veiculo.ID)
		opts.SetCleanSession(true) // Sempre começar com uma sessão limpa
		opts.SetAutoReconnect(true)

		client = mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Println("Erro ao conectar ao broker:", token.Error())
			return
		}

		// Cria um tópico de resposta único para este cliente
		responseTopic = modelo.TopicResposta + "/" + veiculo.ID

		fmt.Printf("Conectado ao broker MQTT em tcp://localhost:1883\n")
	}
}

func onSubmit(idPostos []string, reservar bool) {
	data := struct {
		IDPostos []string `json:"idPostos"`
		Reservar bool     `json:"reservar"`
		ClientID string   `json:"clientId"`
	}{
		IDPostos: idPostos,
		Reservar: reservar,
		ClientID: veiculo.ID,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Erro ao codificar JSON:", err)
		return
	}

	token := client.Publish(modelo.TopicReservarPosto, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		fmt.Println("Erro ao publicar mensagem:", token.Error())
		return
	}

	fmt.Println("Solicitação de reserva enviada")
}

func listarPostos() []modelo.Posto {
	// Usar mutex para garantir apenas uma operação por vez
	listarMutex.Lock()
	defer listarMutex.Unlock()

	var postos []modelo.Posto
	done := make(chan bool, 1) // Buffer de 1 para evitar goroutine leak
	timeout := time.After(10 * time.Second)

	// Definir um callback de única vez que se auto-cancela após ser chamado
	callback := func(client mqtt.Client, msg mqtt.Message) {
		var result []modelo.Posto
		if err := json.Unmarshal(msg.Payload(), &result); err != nil {
			fmt.Println("Erro ao decodificar resposta:", err)
		} else {
			postos = result
		}

		// Enviar sinal no canal done
		select {
		case done <- true:
			// Sinal enviado com sucesso
		default:
			// Canal já fechado ou recebeu um sinal, ignorar
		}
	}

	// Primeiro, cancelar qualquer inscrição anterior
	if token := client.Unsubscribe(responseTopic); token != nil {
		token.Wait()
		// Adicionando uma pequena pausa para garantir que a inscrição anterior foi cancelada
		time.Sleep(100 * time.Millisecond)
	}

	// Inscrever-se no tópico de resposta
	token := client.Subscribe(responseTopic, 1, callback)
	if token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao subscrever:", token.Error())
		return []modelo.Posto{}
	}

	// Preparar a solicitação com ID do cliente
	requestData := struct {
		ClientID string `json:"clientId"`
	}{
		ClientID: veiculo.ID,
	}

	requestPayload, _ := json.Marshal(requestData)

	// Publicar a solicitação
	token = client.Publish(modelo.TopicPostosDisponiveis, 1, false, requestPayload)
	if token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao enviar solicitação:", token.Error())
		client.Unsubscribe(responseTopic)
		return []modelo.Posto{}
	}

	// Aguardar a resposta ou timeout
	select {
	case <-done:
		fmt.Println("Resposta recebida")
	case <-timeout:
		fmt.Println("Tempo esgotado aguardando resposta")
	}

	// Tentar cancelar a inscrição, mas não bloquear se falhar
	go func() {
		token := client.Unsubscribe(responseTopic)
		token.Wait()
	}()

	// Exibir os postos
	if len(postos) > 0 {
		fmt.Printf("\n\nPostos disponíveis:\n")
		for _, p := range postos {
			fmt.Printf("- %s (%f, %f) - Servidor: %s\n", p.ID, p.Latitude, p.Longitude, p.ServidorOrigem)
		}
	} else {
		fmt.Println("Nenhum posto disponível encontrado")
	}

	return postos
}

func procurarPostosParaReserva(postos []modelo.Posto) {
	var reserva string

	fmt.Println("Digite 1 ou pressione Enter para sair")
	fmt.Println("Digite o ID dos postos que deseja reservar, um de cada vez: ")

	idPostos = []string{} // Limpa lista anterior
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
		fmt.Println("6 - Sair")
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
		case 6:
			if client != nil && client.IsConnected() {
				client.Disconnect(250)
			}
			return
		default:
			fmt.Println("Opção inválida.")
		}
	}
}
