package main

import (
	"encoding/json"
	"fmt"
	"main/client"
	"main/global"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	client, err := client.NewMQTTClient(global.PORT, global.BROKER)
	if err != nil {
		fmt.Println("Error creating MQTT client:", err)
		return
	}
	// Subscribe to the topic
	client.Subscribe(global.Consult, PrintMqttMessage)
	client.Subscribe(global.Reserve, PrintMqttMessage)
	time.Sleep(60 * time.Second)
}

func PrintMqttMessage(clint mqtt.Client, msg mqtt.Message) {
	mqttMessage := global.MQTT_Message{}
	json.Unmarshal(msg.Payload(), &mqttMessage)
	fmt.Printf("Topic: %s\n", mqttMessage.Topic.String())
	fmt.Printf("Message: %s\n", mqttMessage.Message)
}
