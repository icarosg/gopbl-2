package main

import (
	mqtt "clients/mqtt"
	types "clients/types"
	"encoding/json"
	"fmt"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Station struct {
	StationID  int
	ServerIP   string
	ReservedBy int // ID do cliente carro que reservou o posto
	Mqtt       *mqtt.MQTT
}

func main() {
	// Input das informações do posto
	serverIP, stationID := "", 0
	fmt.Println("Insira o IP do server/empresa a qual esse posto pertence:")
	fmt.Scanln(&serverIP)
	fmt.Printf("Insira o ID do posto:")
	fmt.Scanln(&stationID)

	fmt.Printf(`Informações do posto:
	Posto ID: %d
	IP do Servidor: %s`, stationID, serverIP)

	// Cria o cliente MQTT
	mqttClient, err := mqtt.NewMQTTClient(types.PORT, types.BROKER)
	if err != nil {
		fmt.Println("Error creating MQTT client:", err)
		return
	}

	// Estado do posto
	station := Station{
		StationID:  stationID,
		ServerIP:   serverIP,
		ReservedBy: -1,
		Mqtt:       mqttClient,
	}

	// Mensagem de nascimento do posto, que informa o servidor que o posto está online
	birthMessage, err := station.BirthMessage()
	if err != nil {
		fmt.Println("Error creating birth message:", err)
		return
	}
	err = station.Mqtt.Publish(birthMessage)
	if err != nil {
		fmt.Println("Error publishing birth message:", err)
		return
	}

	// Topico para reservar o posto
	topic := types.StationReserveTopic(station.ServerIP, station.StationID)
	// Inscrição no tópico de reserva, e atribui a função de callback
	station.Mqtt.Subscribe(topic, func(client paho.Client, msg paho.Message) {
		carInfo := &types.CarInfo{}
		// Decodifica a mensagem recebida
		err := json.Unmarshal(msg.Payload(), carInfo)
		if err != nil {
			fmt.Println("Error unmarshalling car info:", err)
			return
		}
		// Atualiza o ID do carro que reservou o posto
		station.ReservedBy = carInfo.CarId
		fmt.Printf("Posto %d reservado pelo carro %d\n", station.StationID, carInfo.CarId)
	})

	// Mantem o cliente MQTT ativo até o usuário encerrar
	fmt.Println("Enter para encerra o posto")
	fmt.Scanln()
	// Mensagem de morte do posto, que informa o servidor que o posto está offline
	message, err := station.DeathMessage()
	if err != nil {
		fmt.Println("Error creating death message:", err)
		return
	}
	station.Mqtt.Publish(message)
}

func (s *Station) BirthMessage() (types.MQTT_Message, error) {
	topic := types.StationBirthTopic(s.ServerIP)

	station := &types.Station{
		ID:         s.StationID,
		ServerIP:   s.ServerIP,
		ReservedBy: s.ReservedBy,
	}

	payload, err := json.Marshal(station)
	if err != nil {
		return types.MQTT_Message{}, err
	}

	return types.MQTT_Message{
		Topic:   topic,
		Message: payload,
	}, nil
}

func (s *Station) DeathMessage() (types.MQTT_Message, error) {
	topic := types.StationDeathTopic(s.ServerIP)

	payload, err := json.Marshal(s)
	if err != nil {
		return types.MQTT_Message{}, err
	}

	return types.MQTT_Message{
		Topic:   topic,
		Message: payload,
	}, nil
}
