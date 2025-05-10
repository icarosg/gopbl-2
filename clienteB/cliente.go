package main

import (
	"encoding/json"
	"fmt"
	"gopbl-2/modelo"
	//"strings"
	//"sync"
	"time"

	//"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	//"log"
	//"time"
)

var veiculo modelo.Veiculo
var cadastrado bool = false
//var mutex sync.Mutex
var confirm bool = false
var idPostos []string // Restaurado para manter reservas

var postos_servidor_A = make(map[string]modelo.Posto)
// var postos_servidor_B = make(map[string]modelo.Posto)
// var postos_servidor_C = make(map[string]modelo.Posto)

// var rota []modelo.Posto

var postos_servidor_AL []modelo.Posto
// var postos_servidor_BL []modelo.Posto
// var postos_servidor_CL []modelo.Posto

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: from topic: %s\n", msg.Topic())
	switch msg.Topic() {
	case "topic/consulta-cliente-1-B":
		postos_servidor_A = make(map[string]modelo.Posto)		
		err := json.Unmarshal(msg.Payload(), &postos_servidor_AL)
		if err != nil {
			fmt.Printf("Error decoding payload: %v\n", err)
		}
		for _, posto := range postos_servidor_AL {
			postos_servidor_A[posto.ID] = posto
		}
	// case "topic/consulta-cliente-1-B":
	// 	postos_servidor_B = make(map[string]modelo.Posto)
	// 	err := json.Unmarshal(msg.Payload(), &postos_servidor_BL)
	// 	if err != nil {
	// 		fmt.Printf("Error decoding payload: %v\n", err)
	// 	}
	// 	for _, posto := range postos_servidor_BL {
	// 		postos_servidor_B[posto.ID] = posto
	// 	}
	// case "topic/consulta-cliente-1-C":
	// 	postos_servidor_C = make(map[string]modelo.Posto)
	// 	err := json.Unmarshal(msg.Payload(), &postos_servidor_CL)
	// 	if err != nil {
	// 		fmt.Printf("Error decoding payload: %v\n", err)
	// 	}
	// 	for _, posto := range postos_servidor_CL {
	// 		postos_servidor_C[posto.ID] = posto
	// 	}
	case "topic/confirmacao-veiculo-2":
		confirm = true
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
	var broker = "192.168.0.110"
	//var broker = "26.233.72.191"
	var port = 1883
	//var port = 1883
	opts := mqtt.NewClientOptions()
	//opts.AddBroker("tcp://192.168.0.110:1883")
	var idC string
	fmt.Println("digite o ID do cliente")
	fmt.Scanln(&idC)
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(idC)
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
	topic := "topic/consulta-cliente-1-B"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()

	// topic = "topic/consulta-cliente-1-B"
	// token = client.Subscribe(topic, 1, nil)
	// token.Wait()

	// topic = "topic/consulta-cliente-1-C"
	// token = client.Subscribe(topic, 1, nil)
	// token.Wait()
	//fmt.Printf("Subscribed to topic: %s", topic)

	topic = "topic/confirmacao-veiculo-2"
	token = client.Subscribe(topic, 1, nil)
	token.Wait()

}

func menu(client mqtt.Client) {
	for {
		fmt.Println("1 - cadastrar veiculo")
		fmt.Println("2 - atualizar posiçao veiculo")
		fmt.Println("3 - consultar postos de recarga")
		fmt.Println("4 - selecionar postos e enviar rota")
		fmt.Println("5 - terminar viagem e liberar postos")
		fmt.Println("****************************************")
		var opcao int
		fmt.Scanln(&opcao)
		switch opcao {
		case 1:
			fmt.Println("digite o ID do veiculo")
			var id string
			fmt.Scanln(&id)
			fmt.Println("digite a Latitude do veiculo")
			var lat float64
			fmt.Scanln(&lat)
			fmt.Println("digite a Longitude do veiculo")
			var long float64
			fmt.Scanln(&long)
			veiculo = modelo.NovoVeiculo(id, lat, long)
			fmt.Println(veiculo)
			cadastrado = true
		case 2:
			if !cadastrado {
				fmt.Println("veiculo nao cadastrado")
				continue
			}
			fmt.Println("digite a nova Latitude do veiculo")
			var lat float64
			fmt.Scanln(&lat)
			fmt.Println("digite a nova Longitude do veiculo")
			var long float64
			fmt.Scanln(&long)
			veiculo.Latitude = lat
			veiculo.Longitude = long
			fmt.Println(veiculo)
		case 3:
			if !cadastrado {
				fmt.Println("veiculo nao cadastrado")
				continue
			}
			postos_servidor_A = make(map[string]modelo.Posto)
			// postos_servidor_B = make(map[string]modelo.Posto)
			// postos_servidor_C = make(map[string]modelo.Posto)
			token := client.Publish("topic/pedido-consulta-cliente-1-B", 0, false, "postos-server-A")			
			token.Wait()
			// token = client.Publish("topic/pedido-consulta-cliente-1-B", 0, false, "postos-server-B")
			// token.Wait()
			// token = client.Publish("topic/pedido-consulta-cliente-1-C", 0, false, "postos-server-C")
			// token.Wait()
			
			fmt.Println("consultando postos de recarga...")
			//time.Sleep(4 * time.Second)
			for !confirm {
				fmt.Println("consultando postos de recarga...")
				time.Sleep(2 * time.Second) // Aguarda 2 segundos antes de verificar novamente
				// Aguarda a confirmação de que os postos foram recebidos
			}
			confirm = false
			fmt.Println("postos de recarga encontrados:")
			fmt.Println("****************************")
			fmt.Println("Postos disponiveis:")
			for i := range postos_servidor_A {
				p := postos_servidor_A[i]
				fmt.Println(" ", p)
			}
			// fmt.Println("****************************")
			// fmt.Println("Postos do servidor B:")
			// for i := range postos_servidor_B {
			// 	p := postos_servidor_B[i]
			// 	fmt.Println(" ", p)
			// }
			// fmt.Println("****************************")
			// fmt.Println("Postos do servidor C:")
			// for i := range postos_servidor_C {
			// 	p := postos_servidor_C[i]
			// 	fmt.Println(" ", p)
			// }
		case 4:
			if !cadastrado {
				fmt.Println("veiculo nao cadastrado")
				continue
			}
			// fmt.Print("informe sua rota de servidores? Ex: A,B,C: ")
			// var input string
			// fmt.Scanln(&input)
			// sequencia := strings.Split(strings.ToUpper(input), ",")

			// var posto modelo.Posto
			// for _, servidor := range sequencia {
			// 	switch servidor {
			// 	case "A":
			// 		fmt.Print("Digite o ID do posto do servidor A: ")
			// 		var id string
			// 		fmt.Scanln(&id)
			// 		posto = postos_servidor_A[id]
			// 		if posto.ID == "" {
			// 			fmt.Println("Posto não encontrado no servidor A")
			// 			continue
			// 		}
			// 		fmt.Println("Posto do servidor A encontrado:", posto)
			// 		rota = append(rota, posto)

			// 	case "B":
			// 		fmt.Print("Digite o ID do posto do servidor B: ")
			// 		var id string
			// 		fmt.Scanln(&id)
			// 		posto = postos_servidor_B[id]
			// 		if posto.ID == "" {
			// 			fmt.Println("Posto não encontrado no servidor B")
			// 			continue
			// 		}
			// 		fmt.Println("Posto do servidor B encontrado:", posto)
			// 		rota = append(rota, posto)
			// 	case "C":
			// 		fmt.Print("Digite o ID do posto do servidor C: ")
			// 		var id string
			// 		fmt.Scanln(&id)
			// 		posto = postos_servidor_C[id]
			// 		if posto.ID == "" {
			// 			fmt.Println("Posto não encontrado no servidor C")
			// 			continue
			// 		}
			// 		fmt.Println("Posto do servidor C encontrado:", posto)
			// 		rota = append(rota, posto)
			// 	default:
			// 		fmt.Println("Servidor inválido:", servidor)
			// 		continue
			// 	}

			// }
			// req := modelo.ReqAtomica{
			// 	Veiculo: veiculo,
			// 	Posto1:  rota[0],
			// 	Posto2:  rota[1],
			// 	Posto3:  rota[2],
			// }
			// client.Publish("topic/reqAtomica", 0, false, req)
			// fmt.Println("Todos os postos foram processados com sucesso.")
			if len(idPostos) == 0 || idPostos == nil {
				//postos := listarPostos(client)
				listarPostos(client)
				postos := []modelo.Posto{}
				for _, posto := range postos_servidor_A {
					postos = append(postos, posto)
				}

				if len(postos) > 0 {
					procurarPostosParaReserva(postos, client)
				} else {
					fmt.Println("Postos não encontrado.")
				}
			} else {
				fmt.Println("Você já possui reservas em andamento. Finalize a viagem para reservar novamente.")
			}
		case 5:
			if !cadastrado {
				fmt.Println("veiculo nao cadastrado")
				continue
			}
			if len(idPostos) > 0 {
				data := struct {
					IDPostos []string `json:"idPostos"`
					Reservar bool     `json:"reservar"`
				}{
					IDPostos: idPostos,
					Reservar: false,
				}
			
				payload, err := json.Marshal(data)
				if err != nil {
					fmt.Println("Erro ao codificar JSON:", err)
					return
				}
                idPostos = []string{}
				client.Publish("topic/encerrar-viagem-2", 0,false, payload)
            } else {
                fmt.Println("Você não possui reservas.")
            }

		default:
			fmt.Println("opçao invalida")
		}
	}


}

func procurarPostosParaReserva(postos []modelo.Posto, client mqtt.Client) {
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
		//onSubmit(idPostos, true)
		data := struct {
			IDPostos []string `json:"idPostos"`
			Reservar bool     `json:"reservar"`
		}{
			IDPostos: idPostos,
			Reservar: true,
		}
	
		payload, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Erro ao codificar JSON:", err)
			return
		}
		client.Publish("topic/reqAtomica-2", 0, false, payload)
	} else {
		fmt.Println("Nenhum posto selecionado.")
	}
}

func listarPostos(client mqtt.Client){
	postos_servidor_A = make(map[string]modelo.Posto)
			// postos_servidor_B = make(map[string]modelo.Posto)
			// postos_servidor_C = make(map[string]modelo.Posto)
			// token := client.Publish("topic/pedido-consulta-cliente-1-A", 0, false, "postos-server-A")			
			// token.Wait()
			token := client.Publish("topic/pedido-consulta-cliente-1-B", 0, false, "postos-server-B")
			token.Wait()
			// token = client.Publish("topic/pedido-consulta-cliente-1-C", 0, false, "postos-server-C")
			// token.Wait()
			
			fmt.Println("consultando postos de recarga...")
			//time.Sleep(4 * time.Second)
			for !confirm {
				fmt.Println("consultando postos de recarga...")
				time.Sleep(2 * time.Second) // Aguarda 2 segundos antes de verificar novamente
				// Aguarda a confirmação de que os postos foram recebidos
			}
			confirm = false
			fmt.Println("postos de recarga encontrados:")
			fmt.Println("****************************")
			fmt.Println("Postos disponiveis:")
			for i := range postos_servidor_A {
				p := postos_servidor_A[i]
				fmt.Println(" ", p)
			}
}
