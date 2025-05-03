package main

import (
	"encoding/json"
	"fmt"
	"gopbl-2/modelo"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	//"log"
	"time"
)

var veiculo modelo.Veiculo
var cadastrado bool = false

var postos_servidor_A = make(map[string]modelo.Posto)
var postos_servidor_B = make(map[string]modelo.Posto)
var postos_servidor_C = make(map[string]modelo.Posto)

var rota []modelo.Posto

//var postos_servidor_A []modelo.Posto
// var postos_servidor_B []modelo.Posto
// var postos_servidor_C []modelo.Posto

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	switch msg.Topic() {
	case "topic/consulta-cliente-1-A":
		err := json.Unmarshal(msg.Payload(), &postos_servidor_A)
		if err != nil {
			fmt.Printf("Error decoding payload: %v\n", err)
		}
	case "topic/consulta-cliente-1-B":
		err := json.Unmarshal(msg.Payload(), &postos_servidor_B)
		if err != nil {
			fmt.Printf("Error decoding payload: %v\n", err)
		}
	case "topic/consulta-cliente-1-C":
		err := json.Unmarshal(msg.Payload(), &postos_servidor_C)
		if err != nil {
			fmt.Printf("Error decoding payload: %v\n", err)
		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
	fmt.Println("conecetei")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
	fmt.Println("desconectei")
}

func main() {
	var broker = "localhost"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername("emqx")
	opts.SetPassword("public")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sub(client)
	//publish(client)
	go menu(client)
	//defer client.Disconnect(250)
}

func publish(client mqtt.Client) {
	//num := 10
	token := client.Publish("topic/receba", 0, false, "receba")
	token.Wait()
	// for i := 0; i < num; i++ {
	// 	text := fmt.Sprintf("Message %d", i)
	// 	token := client.Publish("topic/test", 0, false, text)
	// 	token.Wait()
	// 	time.Sleep(time.Second)
	// }
}

func sub(client mqtt.Client) {
	// topic := "topic/test"
	// token := client.Subscribe(topic, 1, nil)
	//client.Subscribe("topic/testar", 1, nil)
	//token.Wait()
	topic := "topic/consulta-cliente-1-A"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()

	topic = "topic/consulta-cliente-1-B"
	token = client.Subscribe(topic, 1, nil)
	token.Wait()

	topic = "topic/consulta-cliente-1-C"
	token = client.Subscribe(topic, 1, nil)
	token.Wait()
	//fmt.Printf("Subscribed to topic: %s", topic)
}

func menu(client mqtt.Client) {
	for {
		fmt.Println("1 - cadastrar veiculo")
		fmt.Println("2 - atualizar posiçao veiculo")
		fmt.Println("3 - consultar postos de recarga")
		fmt.Println("4 - selecionar postos e enviar rota")
		fmt.Println("5 - terminar viagem e liberar postos")
		var opcao int
		fmt.Scanf("%d", &opcao)
		switch opcao {
		case 1:
			fmt.Println("digite o ID do veiculo")
			var id string
			fmt.Scanf("%s", &id)
			fmt.Println("digite a Latitude do veiculo")
			var lat float64
			fmt.Scanf("%f", &lat)
			fmt.Println("digite a Longitude do veiculo")
			var long float64
			fmt.Scanf("%f", &long)
			veiculo = modelo.NovoVeiculo(id, lat, long)
			fmt.Println(veiculo)
		case 2:
			if !cadastrado {
				fmt.Println("veiculo nao cadastrado")
				continue
			}
			fmt.Println("digite a nova Latitude do veiculo")
			var lat float64
			fmt.Scanf("%f", &lat)
			fmt.Println("digite a nova Longitude do veiculo")
			var long float64
			fmt.Scanf("%f", &long)
			veiculo.Latitude = lat
			veiculo.Longitude = long
			fmt.Println(veiculo)
		case 3:
			if !cadastrado {
				fmt.Println("veiculo nao cadastrado")
				continue
			}
			token := client.Publish("topic/pedido-consulta-cliente-1-A", 0, false, "postos-server-A")
			token.Wait()
			token = client.Publish("topic/pedido-consulta-cliente-1-B", 0, false, "postos-server-B")
			token.Wait()
			token = client.Publish("topic/pedido-consulta-cliente-1-C", 0, false, "postos-server-C")
			token.Wait()

			fmt.Println("consultando postos de recarga...")
			time.Sleep(2 * time.Second)
			fmt.Println("postos de recarga encontrados:")
			fmt.Println("****************************")
			fmt.Println("Postos do servidor A:")
			for i := range postos_servidor_A {
				p := postos_servidor_A[i]
				fmt.Println(" ", p)
			}
			fmt.Println("****************************")
			fmt.Println("Postos do servidor B:")
			for i := range postos_servidor_B {
				p := postos_servidor_B[i]
				fmt.Println(" ", p)
			}
			fmt.Println("****************************")
			fmt.Println("Postos do servidor C:")
			for i := range postos_servidor_C {
				p := postos_servidor_C[i]
				fmt.Println(" ", p)
			}
		case 4:
			if !cadastrado {
				fmt.Println("veiculo nao cadastrado")
				continue
			}
			fmt.Print("informe sua rota de servidores? Ex: A,B,C: ")
			var input string
			fmt.Scanf("%s", &input)
			sequencia := strings.Split(strings.ToUpper(input), ",")

			var posto modelo.Posto
			for _, servidor := range sequencia {
				switch servidor {
				case "A":
					fmt.Print("Digite o ID do posto do servidor A: ")
					var id string
					fmt.Scanf("%s", &id)
					posto = postos_servidor_A[id]
					if posto.ID == "" {
						fmt.Println("Posto não encontrado no servidor A")
						return
					}
					fmt.Println("Posto do servidor A encontrado:", posto)
					rota = append(rota, posto)

				case "B":
					fmt.Print("Digite o ID do posto do servidor B: ")
					var id string
					fmt.Scanf("%s", &id)
					posto = postos_servidor_B[id]
					if posto.ID == "" {
						fmt.Println("Posto não encontrado no servidor B")
						return
					}
					fmt.Println("Posto do servidor B encontrado:", posto)
					rota = append(rota, posto)
				case "C":
					fmt.Print("Digite o ID do posto do servidor C: ")
					var id string
					fmt.Scanf("%s", &id)
					posto = postos_servidor_C[id]
					if posto.ID == "" {
						fmt.Println("Posto não encontrado no servidor C")
						return
					}
					fmt.Println("Posto do servidor C encontrado:", posto)
					rota = append(rota, posto)
				default:
					fmt.Println("Servidor inválido:", servidor)
					return
				}

			}
			req := modelo.ReqAtomica{
				Veiculo: veiculo,
				Posto1:  rota[0],
				Posto2:  rota[1],
				Posto3:  rota[2],
			}
			client.Publish("topic/reqAtomica", 0, false, req)
			fmt.Println("Todos os postos foram processados com sucesso.")
		}
	}

}
