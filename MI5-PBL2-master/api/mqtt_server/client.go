package mqtt_server

import (
	model "api/model"
	mqtt "api/mqtt"
	types "api/types"
	"encoding/json"
	"fmt"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type ServerState struct {
	ServerIP string
	Mqtt     *mqtt.MQTT
}

func MqttMain() {
	// Cria o cliente MQTT
	mqttClient, err := mqtt.NewMQTTClient(types.PORT, types.BROKER)
	if err != nil {
		fmt.Println("Error creating MQTT client:", err)
		return
	}

	serverIP := ""
	fmt.Println("Insira o IP do server/empresa:")
	fmt.Scanln(&serverIP)

	serverState := ServerState{
		ServerIP: serverIP,
		Mqtt:     mqttClient,
	}

	// Inscrição no tópico de nascimento do carro
	topic := model.CarBirthTopic(serverState.ServerIP)
	serverState.Mqtt.Subscribe(topic, func(client paho.Client, msg paho.Message) {
		// Funcao de callback
		// adiciona o carro na database e se inscreve no tópico de consulta e reserva de rotas

		mqttMessage := &model.MQTT_Message{}
		json.Unmarshal(msg.Payload(), mqttMessage)

		car := &model.Car{}
		json.Unmarshal(mqttMessage.Message, car)

		// TODO adicionar o carro (car) na database

		// Inscrição no tópico de consulta de rotas
		topic = model.CarConsultTopic(serverIP, car.GetCarID())
		serverState.Mqtt.Subscribe(topic, func(client paho.Client, msg paho.Message) {
			// Funcao de callback
			// Deve retornar uma mensagem com payload ListRoutes
			mqttMessage := &model.MQTT_Message{}
			json.Unmarshal(msg.Payload(), mqttMessage)

			routesMessage := &model.RoutesMessage{}
			json.Unmarshal(mqttMessage.Message, routesMessage)

			// city1, city2 := routesMessage.City1, routesMessage.City2
			// TODO requisitar as rotas pela API e retornar na variavel routesList
			routesList := model.RoutesList{
				Routes: []model.Route{},
			}

			payload, _ := json.Marshal(routesList)

			mqttMessage = &model.MQTT_Message{
				Topic:   model.CarConsultTopic(serverIP, car.GetCarID()),
				Message: payload,
			}
			serverState.Mqtt.Publish(*mqttMessage)
		})

		// Inscrição no tópico de reserva de rotas
		topic = model.CarReserveTopic(serverIP, car.GetCarID())
		serverState.Mqtt.Subscribe(topic, func(client paho.Client, msg paho.Message) {
			// Funcao de callback
			// Deve retornar uma mensagem com payload ListRoutes
			mqttMessage := &model.MQTT_Message{}
			json.Unmarshal(msg.Payload(), mqttMessage)

			routesMessage := &model.RoutesMessage{}
			json.Unmarshal(mqttMessage.Message, routesMessage)

			// city1, city2 := routesMessage.City1, routesMessage.City2
			// TODO requisitar as rotas pela API e retornar na variavel routesList
			routesList := model.RoutesList{
				Routes: []model.Route{},
			}

			payload, _ := json.Marshal(routesList)

			mqttMessage = &model.MQTT_Message{
				Topic:   model.CarReserveTopic(serverIP, car.GetCarID()),
				Message: payload,
			}
			serverState.Mqtt.Publish(*mqttMessage)
		})

		topic = model.CarSelectRouteTopic(serverIP, car.GetCarID())
		serverState.Mqtt.Subscribe(topic, func(client paho.Client, msg paho.Message) {
			// Funcao de callback
			// Deve retornar uma mensagem com payload ListRoutes
			mqttMessage := &model.MQTT_Message{}
			json.Unmarshal(msg.Payload(), mqttMessage)

			selectRouteMessage := &model.SelectRouteMessage{}
			json.Unmarshal(mqttMessage.Message, selectRouteMessage)
			// car := selectRouteMessage.Car
			// route := selectRouteMessage.Route
			// for _, waypoint := range route.Waypoints {
			// 	topic = model.StationReserveTopic(serverIP, waypoint)

			// 	carInfo := &model.CarInfo{
			// 		CarId: car.GetCarID(),
			// 	}
			// 	payload, _ := json.Marshal(carInfo)

			// 	mqttMessage = &model.MQTT_Message{
			// 		Topic:   topic,
			// 		Message: payload,
			// 	}

			// 	serverState.Mqtt.Publish(*mqttMessage)
			// }

			// TODO reservar a rota pela API
		})

		topic = model.CarDeathTopic(serverIP)
		serverState.Mqtt.Subscribe(topic, func(client paho.Client, msg paho.Message) {
			// Funcao de callback
			// Retira o carro da database
			mqttMessage := &model.MQTT_Message{}
			json.Unmarshal(msg.Payload(), mqttMessage)

			car := &model.Car{}
			json.Unmarshal(mqttMessage.Message, car)

			serverState.Mqtt.Client.Unsubscribe(
				model.CarConsultTopic(serverIP, car.GetCarID()),
				model.CarReserveTopic(serverIP, car.GetCarID()),
				model.CarSelectRouteTopic(serverIP, car.GetCarID()),
			)

			// TODO retirar o carro da database
		})
	})

	// Inscrição no tópico de nascimento de um posto
	topic = model.StationBirthTopic(serverIP)
	serverState.Mqtt.Subscribe(topic, func(client paho.Client, msg paho.Message) {
		// Funcao de callback
		// Adiciona o posto na database
		mqttMessage := &model.MQTT_Message{}
		json.Unmarshal(msg.Payload(), mqttMessage)
		station := &model.Station{}
		json.Unmarshal(mqttMessage.Message, station)

		// TODO adicionar o posto (station) na database

		// Inscrição no tópico de nascimento de um posto
		topic = model.StationDeathTopic(serverIP)
		serverState.Mqtt.Subscribe(topic, func(client paho.Client, msg paho.Message) {
			// Funcao de callback
			// Retira o posto da database

			mqttMessage := &model.MQTT_Message{}
			json.Unmarshal(msg.Payload(), mqttMessage)
			station := &model.Station{}
			json.Unmarshal(mqttMessage.Message, station)

			// TODO retirar o posto (station) da database
		})
	})

	// Mantem o cliente MQTT ativo até o usuário encerrar
	fmt.Println("Enter para encerra o server")
	fmt.Scanln()
}
