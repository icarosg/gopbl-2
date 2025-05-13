package main

import (
	"fmt"
	"main/client"
	"main/global"
	"time"
)

func main() {
	mqtt, err := client.NewMQTTClient(global.PORT, global.BROKER)

	if err != nil {
		fmt.Println("Error creating MQTT client:", err)
		return
	}

	consultMessage := global.MQTT_Message{
		Topic:   global.Consult,
		Message: "Como que tão os postos ae meu patrão?",
	}

	reserveMessage := global.MQTT_Message{
		Topic:   global.Reserve,
		Message: "Me reserva um posto ae",
	}

	for {
		mqtt.Publish(consultMessage)
		fmt.Println("Message published")
		time.Sleep(1 * time.Second)

		mqtt.Publish(reserveMessage)
		fmt.Println("Message published")
		time.Sleep(1 * time.Second)
	}
}
