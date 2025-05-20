package main

import (
	"gopbl-2/modelo"
	"strconv"
	//"math/rand"
	"fmt"
	"os"
	"os/signal"
	"time"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var posto_criado modelo.Posto
var client mqtt.Client
var cadastroConcluido chan bool = make(chan bool, 1)

// idEnv := os.Getenv("POSTO_ID")
// latEnv, _ := strconv.ParseFloat(os.Getenv("POSTO_LAT"), 64)
// longEnv, _ := strconv.ParseFloat(os.Getenv("POSTO_LONG"), 64)
// servidorEnv := os.Getenv("POSTO_SERVIDOR")
// cidadeEnv := os.Getenv("POSTO_CIDADE")

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Conectado ao broker MQTT")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Conexão perdida: %v\n", err)
}

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Mensagem recebida no tópico: %s\n", msg.Topic())
	fmt.Printf("Conteúdo: %s\n", msg.Payload())
}

func main() {
	//opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
	brokerURL := os.Getenv("MQTT_BROKER")
	if brokerURL == "" {
		brokerURL = "tcp://172.16.201.9:1884"
	}
	opts := mqtt.NewClientOptions().AddBroker(brokerURL)
	//cadastrarPosto()
	gerarPostos()
	clientID := fmt.Sprintf("posto-cliente-%s-%d", posto_criado.ID, time.Now().UnixNano())
	opts.SetClientID(clientID)

	// Definir handlers para monitorar a conexão
	opts.SetDefaultPublishHandler(messageHandler)
	opts.SetOnConnectHandler(connectHandler)
	opts.SetConnectionLostHandler(connectLostHandler)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao conectar ao broker:", token.Error())
		return
	}

	fmt.Printf("Conectado ao broker MQTT em %s\n", "tcp://172.16.201.9:1884")
	fmt.Printf("Usando ID de cliente: %s\n", clientID)

	// Inscrever no tópico de resposta para verificar confirmação
	responseTopic := modelo.TopicResposta + "/" + posto_criado.ID
	fmt.Printf("Inscrevendo no tópico de resposta: %s\n", responseTopic)

	token := client.Subscribe(responseTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Resposta recebida no tópico: %s\n", msg.Topic())
		fmt.Printf("Conteúdo: %s\n", msg.Payload())

		var resposta struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal(msg.Payload(), &resposta); err != nil {
			fmt.Println("Erro ao decodificar resposta:", err)
			return
		}

		fmt.Printf("Resposta do servidor: %s - %s\n", resposta.Status, resposta.Message)
		cadastroConcluido <- true // Sinaliza que recebeu resposta
	})

	if token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao subscrever no tópico de resposta:", token.Error())
		return
	}

	// Aguardar um momento para garantir que a inscrição foi completada
	time.Sleep(500 * time.Millisecond)

	// Para publicação - específico para o servidor de origem
	topic := modelo.TopicCadastrarPosto
	if posto_criado.ServidorOrigem != "" {
		topic = modelo.GetTopicServidor(posto_criado.ServidorOrigem, "cadastrar")
		fmt.Printf("Enviando para tópico específico: %s\n", topic)
	}

	// Publica cadastro (apenas uma vez)
	payload, _ := json.Marshal(posto_criado)
	token = client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao publicar mensagem:", token.Error())
		return
	}

	fmt.Println("Solicitação de cadastro enviada, aguardando resposta...")

	// Aguarda confirmação ou timeout
	select {
	case <-cadastroConcluido:
		fmt.Println("Cadastro processado com sucesso.")
	case <-time.After(15 * time.Second):
		fmt.Println("Tempo esgotado aguardando resposta do servidor.")
		fmt.Println("Verificando estado da conexão...")
		fmt.Printf("Cliente conectado: %v\n", client.IsConnected())

		// Tenta uma nova publicação para diagnóstico
		diagTopic := "diagnostico/cliente/" + posto_criado.ID
		fmt.Printf("Enviando mensagem de diagnóstico para: %s\n", diagTopic)
		diagPayload := []byte("Teste de conexão")
		token = client.Publish(diagTopic, 1, false, diagPayload)
		token.Wait()
		if token.Error() != nil {
			fmt.Printf("Erro ao enviar mensagem de diagnóstico: %v\n", token.Error())
		} else {
			fmt.Println("Mensagem de diagnóstico enviada com sucesso.")
		}
	}

	// Manter a conexão ativa
	fmt.Println("\nO cliente permanecerá conectado. Pressione Ctrl+C para sair.")
	fmt.Println("Ouvindo mensagens do servidor...")

	// Configurar canal para capturar sinais de interrupção (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Esperar sinal de interrupção
	<-c
	fmt.Println("\nRecebido sinal de interrupção. Encerrando cliente...")

	// Publicar no tópico de deletação antes de desconectar
	topic = modelo.TopicCadastrarPosto
	if posto_criado.ServidorOrigem != "" {
		topic = modelo.GetTopicServidor(posto_criado.ServidorOrigem, "deletar")
		fmt.Printf("Enviando para tópico específico: %s\n", topic)
	}
	payload, _ = json.Marshal(posto_criado)
	token = client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		fmt.Println("Erro ao publicar mensagem de deletação:", token.Error())
	} else {
		fmt.Println("Mensagem de deletação publicada com sucesso.")
	}

	// Desconectar do broker
	client.Disconnect(250)
}

func gerarPostos() {
	idEnv := os.Getenv("POSTO_ID")
	latEnv, _ := strconv.ParseFloat(os.Getenv("POSTO_LAT"), 64)
	longEnv, _ := strconv.ParseFloat(os.Getenv("POSTO_LONG"), 64)
	servidorEnv := os.Getenv("POSTO_SERVIDOR")
	cidadeEnv := os.Getenv("POSTO_CIDADE")
	posto_criado = modelo.NovoPosto(idEnv, latEnv, longEnv, servidorEnv, cidadeEnv)
}

func cadastrarPosto() {
	var id string
	var lat float64
	var long float64
	var servidor string
	var cidade string
	fmt.Println("Cadastro do Posto")
	fmt.Print("Digite o ID do posto: ")
	fmt.Scanln(&id)
	fmt.Print("Digite a latitude do posto: ")
	fmt.Scanln(&lat)
	fmt.Print("Digite a longitude do posto: ")
	fmt.Scanln(&long)

	fmt.Println("\nSelecione o servidor:")
	fmt.Println("1 - Servidor Ipiranga")
	fmt.Println("2 - Servidor 22")
	fmt.Println("3 - Servidor Shell")
	var opcao int
	fmt.Print("Escolha uma opção: ")
	fmt.Scanln(&opcao)
	fmt.Println("Informe a cidade do posto: ")
	fmt.Println("1 - Feira de Santana")
	fmt.Println("2 - São Gonçalo")
	fmt.Println("3 - Serrinha")
	fmt.Println("4 - Caruaru")
	fmt.Println("5 - Petrolina")
	fmt.Println("6 - Recife")
	fmt.Println("7 - Chapecó")
	fmt.Println("8 - florianópolis")
	fmt.Println("9 - Joinville")
	var op int
	fmt.Scanln(&op)
	switch op {
	case 1:
		cidade = "Feira de Santana"
	case 2:
		cidade = "são Gonçalo"
	case 3:
		cidade = "serrinha"
	case 4:
		cidade = "Caruaru"
	case 5:
		cidade = "Petrolina"
	case 6:
		cidade = "Recife"
	case 7:
		cidade = "Chapecó"
	case 8:
		cidade = "florianópolis"
	case 9:
		cidade = "Joinville"
	default:
		fmt.Println("opção invalida, cidade padrão: Feira de Santana")
		cidade = "Feira de Santana"
	}
	switch opcao {
	case 1:
		servidor = "Ipiranga"
		fmt.Println("Informe a cidade do posto: ")
		fmt.Println("1 - Feira de Santana")
		fmt.Println("2 - São Gonçalo")
		fmt.Println("3 - Serrinha")
		var op int
		fmt.Scanln(&op)
		switch op {
		case 1:
			cidade = "Feira de Santana"
		case 2:
			cidade = "são Gonçalo"
		case 3:
			cidade = "serrinha"
		default:
			fmt.Println("opção invalida, cidade padrão: Feira de Santana")
			cidade = "Feira de Santana"
		}
	case 2:
		servidor = "22"
		fmt.Println("Informe a cidade do posto: ")
		fmt.Println("1 - Caruaru")
		fmt.Println("2 - Petrolina")
		fmt.Println("3 - Recife")
		var op int
		fmt.Scanln(&op)
		switch op {
		case 1:
			cidade = "Caruaru"
		case 2:
			cidade = "Petrolina"
		case 3:
			cidade = "Recife"
		default:
			fmt.Println("opção invalida, cidade padrão: Caruaru")
			cidade = "Caruaru"
		}
	case 3:
		servidor = "Shell"
		fmt.Println("Informe a cidade do posto: ")
		fmt.Println("1 - Chapecó")
		fmt.Println("2 - florianópolis")
		fmt.Println("3 - Joinville")
		var op int
		fmt.Scanln(&op)
		switch op {
		case 1:
			cidade = "Chapecó"
		case 2:
			cidade = "florianópolis"
		case 3:
			cidade = "Joinville"
		default:
			fmt.Println("opção invalida, cidade padrão: Chapecó")
			cidade = "Chapecó"
		}
	default:
		fmt.Println("Opção inválida. Utilizando Ipiranga como padrão.")
		servidor = "Ipiranga"
	}

	fmt.Printf("Servidor selecionado: %s\n", servidor)

	posto_criado = modelo.NovoPosto(id, lat, long, servidor, cidade)
}
