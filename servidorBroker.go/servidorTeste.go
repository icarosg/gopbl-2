package modelo

import (
	"fmt"
	"encoding/json"
	//"time"
	
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/mochi-co/mqtt/server"
	"github.com/mochi-co/mqtt/server/events"
	"gopbl-2/modelo"
)

// ... (seus tipos PostoJson, PagamentoJson, Requisicao permanecem iguais) ...

var postos_servidor_A = make(map[string]modelo.Posto)

func startMQTTBroker() *server.Server {
	// Cria uma nova instância do broker MQTT
	s := server.New()
	
	// Configuração do broker
	tcp := &server.TCP{
		Addr: ":1883",  // Escuta na porta 1883
	}
	
	// Handlers de eventos
	s.Events.OnMessage = func(cl events.Client, pk events.Packet) (pkx events.Packet, err error) {
		fmt.Printf("Mensagem recebida no tópico: %s - Conteúdo: %s\n", pk.TopicName, pk.Payload)
		
		// Seu tratamento de mensagens personalizado
		if pk.TopicName == "topic/pedido-consulta-cliente-1-A" {
			payload, err := json.Marshal(postos_servidor_A)
			if err != nil {
				fmt.Println("Erro ao codificar postos:", err)
				return pk, nil
			}
			
			err = s.Publish("topic/consulta-cliente-1-A", payload, false)
			if err != nil {
				fmt.Println("Erro ao publicar resposta:", err)
			}
		}
		return pk, nil
	}
	
	// Inicia o broker
	err := s.Serve(tcp)
	if err != nil {
		panic(err)
	}
	
	return s
}

func main() {
	// Inicia o broker MQTT embutido
	broker := startMQTTBroker()
	defer broker.Close()

	// Configuração inicial dos postos
	postos_servidor_A["posto-1"] = modelo.Posto{
		ID:        "posto-1",
		Latitude:  50,
		Longitude: 50,
	}

	// Configura o servidor Gin
	gin.SetMode(gin.ReleaseMode)
	rota := gin.Default()
	
	// Adicione suas rotas HTTP aqui
	// rota.GET("/ping", func(c *gin.Context) {
	//     c.JSON(200, gin.H{"message": "pong"})
	// })
	
	// Menu interativo (opcional)
	go interactiveMenu(broker)
	
	fmt.Println("Servidor iniciado como broker MQTT na porta 1883")
	rota.Run("localhost:8080")
}

func interactiveMenu(broker *server.Server) {
	for {
		fmt.Println("\nMenu do Broker MQTT")
		fmt.Println("1 - Enviar mensagem de teste")
		fmt.Println("2 - Listar clientes conectados")
		fmt.Println("3 - Sair")
		
		var opcao int
		fmt.Scanln(&opcao)
		
		switch opcao {
		case 1:
			fmt.Println("Digite a mensagem:")
			var msg string
			fmt.Scanln(&msg)
			broker.Publish("topic/teste", []byte(msg), false)
			
		case 2:
			clients := broker.Clients()
			fmt.Println("Clientes conectados:")
			for _, client := range clients {
				fmt.Printf("- ID: %s, Endereço: %s\n", client.ID, client.Net.Remote)
			}
			
		case 3:
			return
		}
	}
}