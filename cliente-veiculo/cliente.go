package main

import (
	"fmt"
	"gopbl-2/modelo"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"encoding/json"
	"sync"
	"time"
)

var veiculo modelo.Veiculo
var cadastrado bool = false
var idPostos []string
var client mqtt.Client
var opts *mqtt.ClientOptions
var responseTopic string
var listarMutex sync.Mutex   // mutex para evitar chamadas concorrentes à função listarPostos
var servidorPreferido string // armazena o servidor preferido do cliente

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

	selecionarServidorPreferido() // solicitar o servidoro cliente quer fazer parte

	conectarAoBroker()
}

func selecionarServidorPreferido() {
	fmt.Println("\nSelecione o servidor preferido:")
	fmt.Println("1 - Servidor Ipiranga")
	fmt.Println("2 - Servidor 22")
	fmt.Println("3 - Servidor Shell")
	var opcao int
	fmt.Print("Escolha uma opção: ")
	fmt.Scanln(&opcao)

	switch opcao {
	case 1:
		servidorPreferido = "Ipiranga"
	case 2:
		servidorPreferido = "22"
	case 3:
		servidorPreferido = "Shell"
	default:
		fmt.Println("Opção inválida. Utilizando Ipiranga como padrão.")
		servidorPreferido = "Ipiranga"
	}
	fmt.Printf("Servidor preferido selecionado: %s\n", servidorPreferido)
}
//172.16.103.11
func conectarAoBroker() {
	if veiculo.ID != "" {
		// opts = mqtt.NewClientOptions().AddBroker("tcp://172.18.0.1:1883")
		brokerURL := os.Getenv("MQTT_BROKER")
		if brokerURL == "" {
			brokerURL = "tcp://172.18.0.1:8083"
		}
		opts = mqtt.NewClientOptions().AddBroker(brokerURL)
		opts.SetClientID(veiculo.ID)
		opts.SetCleanSession(true) // sempre começar com uma sessão limpa
		opts.SetAutoReconnect(true)

		client = mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Println("Erro ao conectar ao broker:", token.Error())
			return
		}

		// cria um tópico de resposta único para este cliente
		responseTopic = modelo.TopicResposta + "/" + veiculo.ID

		fmt.Printf("Conectado ao broker MQTT em tcp://172.18.0.1:1883\n")
	}
}

func onSubmit(reservar bool) {
	// se temos servidor preferido, usar o tópico específico
	topic := modelo.TopicReservarPosto
	if servidorPreferido != "" {
		topic = modelo.GetTopicServidor(servidorPreferido, "reservar")
	}

	done := make(chan bool, 1) // Buffer de 1 para evitar goroutine leak
	timeout := time.After(10 * time.Second)

	// define um callback de única vez que se auto-cancela após ser chamado
	reservaFalhou := false
	callback := func(client mqtt.Client, msg mqtt.Message) {
		var resultReservaFalhou bool
		if err := json.Unmarshal(msg.Payload(), &resultReservaFalhou); err != nil {
			fmt.Println("Erro ao decodificar resposta:", err)
		} else {
			reservaFalhou = resultReservaFalhou
		}

		// envia sinal no canal done
		select {
		case done <- true:
			// sinal enviado com sucesso
		default:
			// canal já fechado ou recebeu um sinal, ignorar
		}
	}

	// primeiro, cancela qualquer inscrição anterior
	if token := client.Unsubscribe(responseTopic); token != nil {
		token.Wait()
		time.Sleep(100 * time.Millisecond)
	}

	token := client.Subscribe(responseTopic, 1, callback)
	if token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao subscrever:", token.Error())
		return
	}

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

	token = client.Publish(topic, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		fmt.Println("Erro ao publicar mensagem:", token.Error())
		return
	}

	fmt.Println("Solicitação de reserva enviada para o servidor", servidorPreferido)

	// aguarda a resposta ou timeout
	select {
	case <-done:
		fmt.Println("Resposta recebida")
	case <-timeout:
		fmt.Println("Tempo esgotado aguardando resposta")
	}

	// tenta cancelar a inscrição, mas não bloquear se falhar
	go func() {
		token := client.Unsubscribe(responseTopic)
		token.Wait()
	}()

	if reservaFalhou {
		fmt.Println("A reserva falhou!")
		idPostos = []string{}
	}
}

func listarPostos() []modelo.Posto {
	// usar mutex para garantir apenas uma operação por vez
	listarMutex.Lock()
	defer listarMutex.Unlock()

	var postos []modelo.Posto
	done := make(chan bool, 1) // Buffer de 1 para evitar goroutine leak
	timeout := time.After(10 * time.Second)

	// define um callback de única vez que se auto-cancela após ser chamado
	callback := func(client mqtt.Client, msg mqtt.Message) {
		var result []modelo.Posto
		if err := json.Unmarshal(msg.Payload(), &result); err != nil {
			fmt.Println("Erro ao decodificar resposta:", err)
		} else {
			postos = result
		}

		// envia sinal no canal done
		select {
		case done <- true:
			// sinal enviado com sucesso
		default:
			// canal já fechado ou recebeu um sinal, ignorar
		}
	}

	// primeiro, cancela qualquer inscrição anterior
	if token := client.Unsubscribe(responseTopic); token != nil {
		token.Wait()
		time.Sleep(100 * time.Millisecond)
	}

	token := client.Subscribe(responseTopic, 1, callback)
	if token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao subscrever:", token.Error())
		return []modelo.Posto{}
	}

	// determina o tópico para publicar
	topic := modelo.TopicPostosDisponiveis
	if servidorPreferido != "" {
		topic = modelo.GetTopicServidor(servidorPreferido, "disponiveis")
	}

	// preparar a solicitação com ID do cliente
	requestData := struct {
		ClientID string `json:"clientId"`
	}{
		ClientID: veiculo.ID,
	}

	requestPayload, _ := json.Marshal(requestData)

	// Publicar a solicitação
	token = client.Publish(topic, 1, false, requestPayload)
	if token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao enviar solicitação:", token.Error())
		client.Unsubscribe(responseTopic)
		return []modelo.Posto{}
	}

	fmt.Printf("Solicitação enviada para o servidor %s\n", servidorPreferido)

	// aguarda a resposta ou timeout
	select {
	case <-done:
		fmt.Println("Resposta recebida")
	case <-timeout:
		fmt.Println("Tempo esgotado aguardando resposta")
	}

	// tenta cancelar a inscrição, mas não bloquear se falhar
	go func() {
		token := client.Unsubscribe(responseTopic)
		token.Wait()
	}()
	rotasGeradas := montarRotas(postos)
	// Exibir os postos
	if len(rotasGeradas) > 0 {
		fmt.Printf("\n\nRotas disponíveis:\n")
		for id, rota := range rotasGeradas {
			fmt.Printf("Rota %d: %v\n", id, rota)
		}
	} else {
		fmt.Println("Nenhuma rota disponível encontrada")
	}

	return postos
}

func montarRotas(postos []modelo.Posto) map[int][]modelo.Posto {
	var postosFSA []modelo.Posto
	var postosSonga []modelo.Posto
	var postosSerrinha []modelo.Posto
	var rotas = make(map[int][]modelo.Posto)
	var quantidadeRotas int = 0

	for _, posto := range postos {
		switch posto.Cidade {
		case "Feira de Santana":
			postosFSA = append(postosFSA, posto)
		case "Serrinha":
			postosSerrinha = append(postosSerrinha, posto)
		case "São Gonçalo":
			postosSonga = append(postosSonga, posto)
		}
	}

	var todosPostos []modelo.Posto
	todosPostos = append(todosPostos, postosFSA...)
	todosPostos = append(todosPostos, postosSonga...)
	todosPostos = append(todosPostos, postosSerrinha...)

	// Rotas com 1 posto
	for _, p := range todosPostos {
		rotas[quantidadeRotas] = []modelo.Posto{p}
		quantidadeRotas++
	}

	// Rotas com 2 postos (ordem importa, cidades diferentes)
	for i := 0; i < len(todosPostos); i++ {
		for j := 0; j < len(todosPostos); j++ {
			if i != j && todosPostos[i].Cidade != todosPostos[j].Cidade {
				rotas[quantidadeRotas] = []modelo.Posto{todosPostos[i], todosPostos[j]}
				quantidadeRotas++
			}
		}
	}

	// Rotas com 3 postos (ordem importa, cidades diferentes)
	for i := 0; i < len(todosPostos); i++ {
		for j := 0; j < len(todosPostos); j++ {
			for k := 0; k < len(todosPostos); k++ {
				if i != j && i != k && j != k {
					// Verifica se as 3 cidades são distintas
					ci := todosPostos[i].Cidade
					cj := todosPostos[j].Cidade
					ck := todosPostos[k].Cidade
					if ci != cj && ci != ck && cj != ck {
						rotas[quantidadeRotas] = []modelo.Posto{todosPostos[i], todosPostos[j], todosPostos[k]}
						quantidadeRotas++
					}
				}
			}
		}
	}

	return rotas
}





func procurarPostosParaReserva(rotas map[int][]modelo.Posto) {
	var escolha int
	fmt.Println("\n--- Rotas Disponíveis ---")
	for idx, rota := range rotas {
		fmt.Printf("Rota %d: ", idx)
		for _, posto := range rota {
			fmt.Printf("%s ", posto.ID)
		}
		fmt.Println()
	}

	fmt.Println("\nDigite o número da rota que deseja reservar, ou -1 para sair:")
	fmt.Scanln(&escolha)

	if escolha == -1 {
		fmt.Println("Operação cancelada.")
		return
	}

	rotaEscolhida, existe := rotas[escolha]
	if !existe {
		fmt.Println("Rota inválida.")
		return
	}
	idPostos = []string{} // limpa a lista anterior
	//var idPostos []string
	for _, p := range rotaEscolhida {
		idPostos = append(idPostos, p.ID)
	}

	if len(idPostos) > 0 {
		fmt.Println(idPostos)
		//postosAtuais := listarPostos()
		concluirReserva := true //verifica se todos os postos ainda estão como não reservados
		if concluirReserva {
			onSubmit(true)
		} else {
			fmt.Printf("\n\nOs postos foram reservados por outro cliente. Tente novamente.\n\n")
			idPostos = []string{}
		}
	} else {
		fmt.Println("Nenhum posto selecionado.")
	}
}

func todosPostosPresentes(idPostos []string, postosAtuais []modelo.Posto) bool {
	dicionarioPostosAtuais := make(map[string]bool)
	for _, p := range postosAtuais {
		dicionarioPostosAtuais[p.ID] = true
	}

	// verifica se todos os IDs de idPostos estão no dicionario
	for _, id := range idPostos {
		if !dicionarioPostosAtuais[id] {
			return false
		}
	}

	return true
}

func menu() {
	for {
		fmt.Println("\nMenu de Ações:")
		fmt.Println("1 - Cadastrar veículo")
		fmt.Println("2 - Atualizar posição do veículo")
		fmt.Println("3 - Consultar postos disponíveis")
		fmt.Println("4 - Reservar posto")
		fmt.Println("5 - Finalizar viagem")
		fmt.Println("6 - Alterar servidor preferido")
		fmt.Println("7 - Sair")
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
				rotas := montarRotas(postos)
				if len(postos) > 0 {
					//procurarPostosParaReserva(postos)
					procurarPostosParaReserva(rotas)
				} else {
					fmt.Println("Postos não encontrado.")
				}
			} else {
				fmt.Println("Você já possui reservas em andamento. Finalize a viagem para reservar novamente.")
			}
		case 5:
			if len(idPostos) > 0 {
				onSubmit(false)
				idPostos = []string{}
			} else {
				fmt.Println("Você não possui reservas.")
			}
		case 6:
			selecionarServidorPreferido()
		case 7:
			if client != nil && client.IsConnected() {
				client.Disconnect(250)
			}
			return
		default:
			fmt.Println("Opção inválida.")
		}
	}
}
