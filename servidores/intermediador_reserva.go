package intermediador_reserva

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"gopbl-2/modelo"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var intermediadorMutex sync.Mutex
var ultimaReservaIntermediada time.Time

func IniciarIntermediadorReserva(mqttClient mqtt.Client) {
	mqttClient.Subscribe(modelo.TopicReservaIntermediador, 1, handleReservaIntermediador)
	log.Printf("Intermediador de reserva inscrito em %s", modelo.TopicReservaIntermediador)
}

func handleReservaIntermediador(client mqtt.Client, msg mqtt.Message) {
	intermediadorMutex.Lock()
	defer intermediadorMutex.Unlock()

	delta := time.Since(ultimaReservaIntermediada)
	if delta < 300*time.Millisecond {
		time.Sleep(300*time.Millisecond - delta)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[Intermediador] Erro ao decodificar payload: %v", err)
		return
	}
	destino, ok := data["destino"].(string)
	if !ok || destino == "" {
		log.Printf("[Intermediador] Payload sem campo 'destino' válido: %v", data)
		return
	}
	token := client.Publish(destino, 1, false, msg.Payload())
	token.Wait()
	if token.Error() != nil {
		log.Printf("[Intermediador] Erro ao publicar no destino %s: %v", destino, token.Error())
		return
	}
	ultimaReservaIntermediada = time.Now()
	log.Printf("[Intermediador] Reserva intermediada e publicada em %s", destino)
}
