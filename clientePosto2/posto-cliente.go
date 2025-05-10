package main

import (
	"encoding/json"
	"fmt"
	"gopbl-2/modelo"

	//"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	//"log"
	//"time"
)

var posto modelo.Posto
var cadastrado bool = false

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	switch msg.Topic() {
	case "topic/listar-postos-2":
		postoC := modelo.PostoConsulta{
			ID:    posto.ID,
			Posto: posto,
		}
		payload, err := json.Marshal(postoC)
		if err != nil {
			fmt.Println("erro ao converter posto")
			return
		}
		client.Publish("topic/receber-posto-2", 0, false, payload)
		//token.Wait()
	case "topic/possivel-reserva-2":
		if posto.BombaOcupada {
			token := client.Publish("topic/possivel-reserva-2-server", 0, false, false)
			token.Wait()
		} else {
			token := client.Publish("topic/possivel-reserva-2-server", 0, false, false)
			token.Wait()
		}
	case "topic/reservar-vaga-2":
		var veiculo modelo.Veiculo
		err := json.Unmarshal(msg.Payload(), &veiculo)
		if err != nil {
			fmt.Println("erro ao reservar vaga, erro de conversao")
			return
		}
		posto.BombaOcupada = true
		posto.Fila = veiculo
	case "topic/liberar-vaga-2":
		posto.BombaOcupada = false
		posto.Fila = modelo.Veiculo{}
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
	var broker = "192.168.0.110" //trocar pelo ip da maquina
	//var broker = "26.233.72.191"
	var port = 1883
	opts := mqtt.NewClientOptions()
	//opts.AddBroker("tcp://192.168.0.110:1883")

	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client_posto_2")
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
	menu(client)
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

	topic := "topic/listar-postos-2"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()

	topic = "topic/possivel-reserva-2"
	token = client.Subscribe(topic, 1, nil)
	token.Wait()

	topic = "topic/reservar-vaga-2"
	token = client.Subscribe(topic, 1, nil)
	token.Wait()

	topic = "topic/liberar-vaga-2"
	token = client.Subscribe(topic, 1, nil)
	token.Wait()

	// topic = "topic/consulta-cliente-1-A"
	// token = client.Subscribe(topic, 1, nil)
	// token.Wait()

	// topic = "topic/consulta-cliente-1-B"
	// token = client.Subscribe(topic, 1, nil)
	// token.Wait()

	// topic = "topic/consulta-cliente-1-C"
	// token = client.Subscribe(topic, 1, nil)
	// token.Wait()
	//fmt.Printf("Subscribed to topic: %s", topic)
}

func menu(client mqtt.Client) {
	for {
		fmt.Println("1 - cadastrar posto")
		fmt.Println("2 - consultar bomba")
		fmt.Println("****************************************")
		var opcao int
		fmt.Scanln(&opcao)
		switch opcao {
		case 1:
			fmt.Println("digite o ID do posto")
			var id string
			fmt.Scanln(&id)
			fmt.Println("digite a Latitude do posto")
			var lat float64
			fmt.Scanln(&lat)
			fmt.Println("digite a Longitude do posto")
			var long float64
			fmt.Scanln(&long)
			posto = modelo.NovoPosto(id, lat, long)
			fmt.Println(posto)
			postoEnviado, e := json.Marshal(posto)
			if e != nil {
				fmt.Println("erro ao converter posto antes de enviar pro servidor")
				continue
			}
			cadastrado = true

			client.Publish("topic/cadastro-posto-2", 0, false, postoEnviado)
		case 2:
			if !cadastrado {
				fmt.Println("posto nao cadastrado")
				continue
			}
			if !posto.BombaOcupada {
				fmt.Println("bomba disponivel")
			} else {
				fmt.Println("posto ocupado/reservado para o veiculo: ", posto.Fila)
			}

		default:
			fmt.Println("opçao invalida")
		}
	}

}
