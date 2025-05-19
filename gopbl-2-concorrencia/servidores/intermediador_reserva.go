package intermediador_reserva

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"gopbl-2/modelo"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	mu                     sync.Mutex
	ultimaReservaIntermediada time.Time
	bloqueioAtivo          bool
)

func IniciarIntermediadorReserva(mqttClient mqtt.Client) {
	mqttClient.Subscribe(modelo.TopicReservaIntermediador, 1, handleReservaIntermediador)
	mqttClient.Subscribe(modelo.TopicReservaCancelamento, 1, handleCancelamentoReserva)
	log.Printf("Intermediador de reserva inscrito em %s", modelo.TopicReservaIntermediador)
}

func handleReservaIntermediador(client mqtt.Client, msg mqtt.Message) {
	mu.Lock()
	defer mu.Unlock()

	agora := time.Now()
	delta := agora.Sub(ultimaReservaIntermediada)

	// Se o bloqueio ainda está ativo ou passou menos de 300ms, recusar
	if bloqueioAtivo || delta < 300*time.Millisecond {
		log.Printf("[Intermediador] Reserva rejeitada: tentativa muito próxima da última ou bloqueio ativo.")
		resposta := map[string]interface{}{
			"sucesso": false,
			"motivo":  "Não foi possível reservar. Tente novamente.",
		}
		resp, _ := json.Marshal(resposta)
		client.Publish(modelo.TopicReservaResposta, 1, false, resp)
		return
	}

	// Marca o início de uma nova tentativa
	bloqueioAtivo = true
	ultimaReservaIntermediada = agora

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

	// Publica a tentativa de reserva
	token := client.Publish(destino, 1, false, msg.Payload())
	token.Wait()
	if token.Error() != nil {
		log.Printf("[Intermediador] Erro ao publicar no destino %s: %v", destino, token.Error())
		return
	}
	log.Printf("[Intermediador] Reserva publicada para %s", destino)

	// Define um timeout para liberar o bloqueio automaticamente após 10s
	go func() {
		select {
		case <-time.After(10 * time.Second):
			mu.Lock()
			bloqueioAtivo = false
			mu.Unlock()
			log.Println("[Intermediador] Bloqueio liberado por timeout")
		}
	}()
}

func handleCancelamentoReserva(client mqtt.Client, msg mqtt.Message) {
	mu.Lock()
	defer mu.Unlock()

	var cancelado bool
	if err := json.Unmarshal(msg.Payload(), &cancelado); err != nil {
		log.Printf("[Cancelamento] Erro ao decodificar cancelamento: %v", err)
		return
	}

	if cancelado {
		bloqueioAtivo = false
		log.Println("[Cancelamento] Bloqueio liberado manualmente via mensagem MQTT")
	}
}
