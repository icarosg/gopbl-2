package main

import (
	"fmt"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"

	"net/http"
	// "os"
	"gopbl-2/db"
	"gopbl-2/modelo"
	"gopbl-2/models"

	//"sync"

	"bytes"
	"time"

	// "io"
	"log"
	// "math"
	// "net"
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
)

type PostoJson struct {
	ID              string  `json:"id"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	QuantidadeFila  int     `json:"quantidade de carros na fila"`
	Disponibilidade bool    `json:"bomba disponivel"`
}

type PagamentoJson struct {
	ID_veiculo string  `json:"id_veiculo"`
	Valor      float64 `json:"valor"`
	Posto      string  `json:"id_posto"`
}

type Requisicao struct {
	Comando string          `json:"comando"`
	Dados   json.RawMessage `json:"dados"`
}

// var mutex sync.Mutex
var postos_servidor_A = make(map[string]modelo.Posto)
var confirmacoes []bool

// var postosChanA = make(chan modelo.PostoConsulta, 10) // buffer de 10
var servidores = []string{
	"http://localhost:8080",
	"http://localhost:8082",
	"http://localhost:8084",
}

var dbServer *db.ConexaoServidorDB
var sincronizadorMQTT *db.SincronizadorMQTT

var messagePubHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	//fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	switch msg.Topic() {
	case "topic/pedido-consulta-cliente-1-A":
		token := client.Publish("topic/listar-postos", 0, false, "listar postos")
		token.Wait()

		// payload, err := json.Marshal(postos_servidor_A)
		// if err != nil {
		// 	fmt.Println("erro ao codificar o dic com os postos")
		// 	return
		// }
		// token = mqttClient.Publish("topic/consulta-cliente-1-A", 0, false, payload)
		// token.Wait()
		// fmt.Println("postos enviados para o cliente 1-A")
	case "topic/receber-posto":
		var posto modelo.PostoConsulta
		err := json.Unmarshal(msg.Payload(), &posto)
		if err != nil {
			fmt.Println("erro ao decodificar o posto recebido")
			return
		}
		postos_servidor_A[posto.ID] = posto.Posto
		
		var todosPostos []modelo.Posto
		//todosPostos = append(todosPostos, postos...)

		for _, servidor := range servidores {
			// if servidor == "http://localhost:8080" {
			// 	continue
			// }

			resp, err := http.Get(servidor + "/postosDisponiveis?consultarOutrosServidores=false")
			if err != nil {
				log.Printf("Erro ao consultar servidor %s: %v", servidor, err)
				continue
			}

			var postos []modelo.Posto
			err = json.NewDecoder(resp.Body).Decode(&postos)
			resp.Body.Close()

			if err != nil {
				log.Printf("Erro ao decodificar resposta do servidor %s: %v", servidor, err)
				continue
			}

			todosPostos = append(todosPostos, postos...)
		}
		fmt.Println(todosPostos)
		payload, err := json.Marshal(todosPostos)
		if err != nil {
			fmt.Println("erro ao codificar o dic com os postos")
		}
		token := mqttClient.Publish("topic/consulta-cliente-1-A", 0, false, payload)
		token.Wait()
		fmt.Println("postos enviados para o cliente 1-A")
		client.Publish("topic/confirmacao-veiculo", 0, false, "")
		//postosChanA <- posto
		fmt.Println("recebi o posto")
	case "topic/reqAtomica":
		
		handleReservarPosto(client, msg)
		

		// var req modelo.ReqAtomica
		// err := json.Unmarshal(msg.Payload(), &req)
		// if err != nil {
		// 	fmt.Println("erro ao decodificar a req atomica")
		// 	return
		// }
		// token := mqttClient.Publish("topic/possivel-reserva", 0, false, req.Veiculo)
		// token.Wait()
		// var all bool = true
		// for i := range confirmacoes {
		// 	if !confirmacoes[i] {
		// 		all = false
		// 		break
		// 	}
		// }
		// if !all {
		// 	fmt.Println("algum postou nao pode ser reservado")
		// 	return
		// }
		// Se todos os postos puderem ser reservados, enviar a confirmação
		// token = mqttClient.Publish("topic/reservar-vaga", 0, false, req.Veiculo)
		// token.Wait()
	case "topic/possivel-reserva-server":
		var confirm bool
		err := json.Unmarshal(msg.Payload(), &confirm)
		if err != nil {
			fmt.Println("erro ao receber a possivel confirmacao")
			return
		}
		confirmacoes = append(confirmacoes, confirm)
	case "topic/cadastro-posto":
		fmt.Println("recebi o cadastro do posto")
		handleCadastrarPosto(client, msg)
	case "topic/encerrar-viagem":
		handleReservarPosto(client, msg)
	}
}

var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

// Variável global para o cliente MQTT
var mqttClient MQTT.Client

// func main() {
// 	postos_servidor_A["posto-1"] = modelo.Posto{
// 		ID:        "posto-1",
// 		Latitude:  50,
// 		Longitude: 50,
// 	}
// 	setupMQTT()
// 	defer mqttClient.Disconnect(250)

// 	gin.DisableBindValidation()
// 	gin.SetMode(gin.ReleaseMode)
// 	rota := gin.Default()
// 	// rota.GET("/ping", func(c *gin.Context) {
// 	// 	c.JSON(200, gin.H{
// 	// 		"message": "pong",
// 	// 	})
// 	// })
// 	rota.Run("localhost:8080")
// 	fmt.Println("Servidor iniciado e conectado ao MQTT Broker")

// }

// obter variáveis de ambiente com valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func handleReservarPosto(client MQTT.Client, msg MQTT.Message) {
	var data models.ReservaData
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("Erro ao decodificar dados de reserva: %v", err)
		return
	}

	disponibilidade := make(map[string]bool)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"id": bson.M{"$in": data.IDPostos}}
	cursor, err := dbServer.PostosCollection.Find(ctx, filter)
	if err == nil {
		var postos []modelo.Posto
		if err = cursor.All(ctx, &postos); err == nil {
			for _, posto := range postos {
				disponibilidade[posto.ID] = !posto.BombaOcupada
			}
		}
	}
	cursor.Close(ctx)

	// consulta disponibilidade em outros servidores
	for _, servidor := range servidores {
		if servidor == "http://localhost:8080" {
			continue
		}

		resp, err := http.Get(servidor + "/postosDisponiveis?consultarOutrosServidores=false")
		if err != nil {
			log.Printf("Erro ao consultar no servidor: %v", err)
			continue
		}
		defer resp.Body.Close()

		var postos []modelo.Posto
		err = json.NewDecoder(resp.Body).Decode(&postos)
		if err != nil {
			log.Printf("Erro ao decodificar resposta: %v", err)
			continue
		}

		for _, idPostoRequisicao := range data.IDPostos {
			for _, posto := range postos {
				if idPostoRequisicao == posto.ID {
					disponibilidade[posto.ID] = !posto.BombaOcupada
				}
			}
		}
	}

	if data.Reservar { // verifica disponibilidade de todos os postos
		for _, id := range data.IDPostos {
			if !disponibilidade[id] {
				log.Printf("Nem todos os postos estão disponíveis")
				return
			}
		}

		// todos postos estão disponíveis, então é reservado
		filtro := bson.M{
			"id":             bson.M{"$in": data.IDPostos},
			"servidorOrigem": dbServer.Nome,
		}
		update := bson.M{"$set": bson.M{
			"bombaocupada":        true,
			"ultimaAtualizacao": time.Now(),
		}}

		_, err = dbServer.PostosCollection.UpdateMany(ctx, filtro, update)
		if err != nil {
			log.Printf("Erro ao atualizar os postos: %v", err)
			return
		}

		//atualiza nos outros servidores
		for _, servidor := range servidores {
			if servidor == "http://localhost:8080" {
				continue
			}

			putData, erro := json.Marshal(struct {
				IDPostos []string `json:"idPostos"`
				Reservar bool     `json:"reservar"`
			}{
				IDPostos: data.IDPostos,
				Reservar: data.Reservar,
			})
			if erro != nil {
				log.Printf("Erro ao codificar JSON: %v", erro)
				continue
			}

			req, erro := http.NewRequest(http.MethodPut, servidor+"/reservar?consultarOutrosServidores=false", bytes.NewBuffer(putData))
			if erro != nil {
				log.Printf("Erro ao criar requisição: %v", erro)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, erro := client.Do(req)
			if erro != nil {
				log.Printf("Erro ao enviar requisição: %v", erro)
				continue
			}
			defer resp.Body.Close()
		}

		log.Printf("Postos reservados com sucesso")
	} else { //finalizar viagem
		filtro := bson.M{
			"id":             bson.M{"$in": data.IDPostos},
			"servidorOrigem": dbServer.Nome,
		}
		update := bson.M{"$set": bson.M{
			"bombaocupada":        false,
			"ultimaAtualizacao": time.Now(),
		}}

		_, err = dbServer.PostosCollection.UpdateMany(ctx, filtro, update)
		if err != nil {
			log.Printf("Erro ao atualizar os postos: %v", err)
			return
		}

		//atualiza nos outros servidores
		for _, servidor := range servidores {
			if servidor == "http://localhost:8080" {
				continue
			}

			putData, erro := json.Marshal(struct {
				IDPostos []string `json:"idPostos"`
				Reservar bool     `json:"reservar"`
			}{
				IDPostos: data.IDPostos,
				Reservar: data.Reservar,
			})
			if erro != nil {
				log.Printf("Erro ao codificar JSON: %v", erro)
				continue
			}

			req, erro := http.NewRequest(http.MethodPut, servidor+"/reservar?consultarOutrosServidores=false", bytes.NewBuffer(putData))
			if erro != nil {
				log.Printf("Erro ao criar requisição: %v", erro)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, erro := client.Do(req)
			if erro != nil {
				log.Printf("Erro ao enviar requisição: %v", erro)
				continue
			}
			defer resp.Body.Close()
		}

		log.Printf("Postos liberados com sucesso")
	}
}

func main() {
	hostDB := getEnv("DB_HOST", "localhost")
	portaDB := 27017
	nomeServidor := "Estado-Bahia"
	//mqttBroker := getEnv("MQTT_BROKER", "tcp://localhost:1883")

	var erro error
	dbServer, erro = db.NovaConexaoDB(nomeServidor, hostDB, portaDB)
	if erro != nil {
		log.Fatal("Erro ao conectar ao MongoDB:", erro)
	}
	fmt.Println("Conectado ao MongoDB com sucesso!")

	// sincronizadorMQTT, erro = db.NovoSincronizadorMQTT(
	// 	dbServer,
	// 	mqttBroker,
	// 	nomeServidor,
	// 	1*time.Minute,
	// )
	// if erro != nil {
	// 	log.Printf("Não foi possível iniciar o sincronizador MQTT: %v", erro)
	// } else {
	// 	sincronizadorMQTT.IniciarSincronizacao()
	// 	fmt.Println("Sincronizador MQTT iniciado com sucesso!")
	// }

	// // fecha a conexão quando o programa terminar
	// defer func() {
	// 	dbServer.Fechar()
	// 	if sincronizadorMQTT != nil {
	// 		sincronizadorMQTT.Fechar()
	// 	}
	// }()

	// Iniciar conexão MQTT
	//configurarMQTT()
	setupMQTT()

	gin.DisableBindValidation()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default() // cria o servidor Gin

	// registra rota
	router.GET("/postosDisponiveis", postosDisponiveisHandler)
	router.POST("/cadastrar", cadastrarPostoHandler)
	router.PUT("/reservar", editarPostoHandler)

	router.Run(":8080")
}

func subscribeToTopics() {
	// Exemplo de subscription
	topic := "topic/receba"
	//token := mqttClient.Subscribe(topic, 1, nil)

	token := mqttClient.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)

	topic = "topic/pedido-consulta-cliente-1-A"
	token = mqttClient.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)

	topic = "topic/receber-posto"
	token = mqttClient.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)

	topic = "topic/reqAtomica"
	token = mqttClient.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)

	topic = "topic/cadastro-posto"
	token = mqttClient.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)

	topic = "topic/encerrar-viagem"
	token = mqttClient.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("subscribe to topic: %s\n", topic)


}

func publishMessage(topic string, message string) {
	token := mqttClient.Publish(topic, 0, false, message)
	token.Wait()
	time.Sleep(time.Second)
}

func setupMQTT() {
	var broker = "192.168.0.110"
	var port = 1884
	opts := MQTT.NewClientOptions()
	//opts.AddBroker("tcp://192.168.0.110:1883")
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_server")
	opts.SetUsername("emqx")
	opts.SetPassword("public")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	mqttClient = MQTT.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Subscrever aos tópicos necessários
	subscribeToTopics()

	// go func() {
	// 	for {
	// 		fmt.Println("Menu")
	// 		fmt.Println("1 - enviar mensagem pro cliente")
	// 		var opcao int
	// 		fmt.Scanln(&opcao)
	// 		switch opcao {
	// 		case 1:
	// 			fmt.Println("Digite a mensagem")
	// 			var mensagem string
	// 			fmt.Scanln(&mensagem)
	// 			mqttClient.Publish("topic/testar", 0, false, mensagem)
	// 		}
	// 	}

	// }()

}

func postosDisponiveisHandler(c *gin.Context) {
	fmt.Println("Entrou na consulta de postos disponíveis")
	if c.Query("consultarOutrosServidores") != "false" {
		// consulta outros servidores também

		var todosPostos []modelo.Posto

		for _, servidor := range servidores {
			if servidor == "http://localhost:8080" {
				continue
			}

			resp, err := http.Get(servidor + "/postosDisponiveis?consultarOutrosServidores=false")
			if err != nil {
				log.Printf("Erro ao consultar servidor %s: %v", servidor, err)
				continue
			}

			var postos []modelo.Posto
			err = json.NewDecoder(resp.Body).Decode(&postos)
			resp.Body.Close()

			if err != nil {
				log.Printf("Erro ao decodificar resposta do servidor %s: %v", servidor, err)
				continue
			}

			todosPostos = append(todosPostos, postos...)
		}

		// busca e adiciona os postos deste servidor
		collection := dbServer.PostosCollection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"bombaocupada": false}
		cursor, erro := collection.Find(ctx, filter)
		if erro != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar postos disponíveis"})
			return
		}
		defer cursor.Close(ctx)

		var postosLocais []modelo.Posto
		if erro = cursor.All(ctx, &postosLocais); erro != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao decodificar os dados"})
			return
		}

		todosPostos = append(todosPostos, postosLocais...)

		c.JSON(http.StatusOK, todosPostos)
		return
	}

	// apenas retorna os postos locais, caso a consulta venha de outro servidor
	collection := dbServer.PostosCollection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"bombaocupada": false}
	cursor, erro := collection.Find(ctx, filter)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar postos disponíveis"})
		return
	}
	defer cursor.Close(ctx)

	var postosLocais []modelo.Posto
	if erro = cursor.All(ctx, &postosLocais); erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao decodificar os dados"})
		return
	}

	c.JSON(http.StatusOK, postosLocais)
}

func cadastrarPostoHandler(c *gin.Context) {
	var novoPosto modelo.Posto

	if erro := c.ShouldBindJSON(&novoPosto); erro != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	// garante que o timestamp está atualizado e definir o servidor de origem
	novoPosto.UltimaAtualizacao = time.Now()
	novoPosto.ServidorOrigem = dbServer.Nome

	collection := dbServer.PostosCollection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": novoPosto.ID}
	var existente modelo.Posto
	erro := collection.FindOne(ctx, filter).Decode(&existente) // verifica se já existe um posto com mesmo nome
	if erro == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Posto já existente"})
		return
	}

	_, erro = collection.InsertOne(ctx, novoPosto)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao inserir o posto"})
		return
	}

	if sincronizadorMQTT != nil { // notificar o sincronizador sobre a alteração
		sincronizadorMQTT.NotificarAlteracao(&novoPosto)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Posto cadastrado com sucesso"})
}

func handleCadastrarPosto(client MQTT.Client, msg MQTT.Message) {
	var novoPosto modelo.Posto
	if err := json.Unmarshal(msg.Payload(), &novoPosto); err != nil {
		log.Printf("Erro ao decodificar posto: %v", err)
		return
	}

	// garante que o timestamp está atualizado e definir o servidor de origem
	novoPosto.UltimaAtualizacao = time.Now()
	novoPosto.ServidorOrigem = dbServer.Nome

	collection := dbServer.PostosCollection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": novoPosto.ID}
	var existente modelo.Posto
	erro := collection.FindOne(ctx, filter).Decode(&existente) // verifica se já existe um posto com mesmo nome
	if erro == nil {
		log.Printf("Posto %s já existe", novoPosto.ID)
		return
	}

	_, erro = collection.InsertOne(ctx, novoPosto)
	if erro != nil {
		log.Printf("Erro ao inserir posto: %v", erro)
		return
	}

	log.Printf("Posto %s cadastrado com sucesso", novoPosto.ID)
}

func editarPostoHandler(c *gin.Context) {
	var data models.ReservaData

	if erro := c.ShouldBindJSON(&data); erro != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido."})
		return
	}

	disponibilidade := make(map[string]bool)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if c.Query("consultarOutrosServidores") == "false" { // somente atualiza neste servidor
		// todos postos estão disponíveis, então é reservado
		filtro := bson.M{
			"id":             bson.M{"$in": data.IDPostos},
			"servidorOrigem": dbServer.Nome,
		}
		update := bson.M{"$set": bson.M{
			"disponivel":        !data.Reservar,
			"ultimaAtualizacao": time.Now(),
		}}

		_, err := dbServer.PostosCollection.UpdateMany(ctx, filtro, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar os postos"})
			return
		}
		return
	}

	filter := bson.M{"id": bson.M{"$in": data.IDPostos}}
	cursor, err := dbServer.PostosCollection.Find(ctx, filter)
	if err == nil {
		var postos []modelo.Posto
		if err = cursor.All(ctx, &postos); err == nil {
			for _, posto := range postos {
				disponibilidade[posto.ID] = posto.BombaOcupada
			}
		}
	}
	cursor.Close(ctx)

	// consulta disponibilidade em outros servidores
	for _, servidor := range servidores {
		if servidor == "http://localhost:8084" {
			continue
		}

		resp, err := http.Get(servidor + "/postosDisponiveis?consultarOutrosServidores=false")
		if err != nil {
			fmt.Println("Erro ao consultar no servidor", err)
			continue
		}
		defer resp.Body.Close()

		var postos []modelo.Posto
		err = json.NewDecoder(resp.Body).Decode(&postos)
		if err != nil {
			fmt.Println("Erro ao decodificar resposta:", err)
			continue
		}

		for _, idPostoRequisicao := range data.IDPostos {
			for _, posto := range postos {
				if idPostoRequisicao == posto.ID {
					disponibilidade[posto.ID] = posto.BombaOcupada
				}
			}
		}
	}

	if data.Reservar { // verifica disponibilidade de todos os postos
		for _, id := range data.IDPostos {
			if !disponibilidade[id] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Nem todos os postos estão disponíveis"})
				return
			}
		}

		// todos postos estão disponíveis, então é reservado
		filtro := bson.M{
			"id":             bson.M{"$in": data.IDPostos},
			"servidorOrigem": dbServer.Nome,
		}
		update := bson.M{"$set": bson.M{
			"disponivel":        false,
			"ultimaAtualizacao": time.Now(),
		}}

		_, err = dbServer.PostosCollection.UpdateMany(ctx, filtro, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar os postos"})
			return
		}

		//atualiza nos outros servidores
		if c.Query("consultarOutrosServidores") != "false" {
			for _, servidor := range servidores {
				if servidor == "http://localhost:8084" {
					continue
				}

				putData, erro := json.Marshal(struct {
					IDPostos []string `json:"idPostos"`
					Reservar bool     `json:"reservar"`
				}{
					IDPostos: data.IDPostos,
					Reservar: data.Reservar,
				})
				if erro != nil {
					fmt.Println("Erro ao codificar JSON:", erro)
					continue
				}

				req, erro := http.NewRequest(http.MethodPut, servidor+"/reservar?consultarOutrosServidores=false", bytes.NewBuffer(putData))
				if erro != nil {
					fmt.Println("erro ao criar requisição:", erro)
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				resp, erro := client.Do(req)
				if erro != nil {
					fmt.Println("erro ao enviar requisição:", erro)
					continue
				}
				defer resp.Body.Close()
			}
		}

		// notifica o sincronizador sobre as alterações
		if sincronizadorMQTT != nil {
			for _, id := range data.IDPostos {
				var posto modelo.Posto
				filterSingle := bson.M{"id": id, "servidorOrigem": dbServer.Nome}
				if err := dbServer.PostosCollection.FindOne(ctx, filterSingle).Decode(&posto); err == nil {
					sincronizadorMQTT.NotificarAlteracao(&posto)
				}
			}
		}
	} else { //finalizar viagem
		filtro := bson.M{
			"id":             bson.M{"$in": data.IDPostos},
			"servidorOrigem": dbServer.Nome,
		}
		update := bson.M{"$set": bson.M{
			"disponivel":        true,
			"ultimaAtualizacao": time.Now(),
		}}

		_, err = dbServer.PostosCollection.UpdateMany(ctx, filtro, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar os postos"})
			return
		}

		//atualiza nos outros servidores
		if c.Query("consultarOutrosServidores") != "false" {
			for _, servidor := range servidores {
				if servidor == "http://localhost:8084" {
					continue
				}

				putData, erro := json.Marshal(struct {
					IDPostos []string `json:"idPostos"`
					Reservar bool     `json:"reservar"`
				}{
					IDPostos: data.IDPostos,
					Reservar: data.Reservar,
				})
				if erro != nil {
					fmt.Println("Erro ao codificar JSON:", erro)
					continue
				}

				req, erro := http.NewRequest(http.MethodPut, servidor+"/reservar?consultarOutrosServidores=false", bytes.NewBuffer(putData))
				if erro != nil {
					fmt.Println("erro ao criar requisição:", erro)
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				resp, erro := client.Do(req)
				if erro != nil {
					fmt.Println("erro ao enviar requisição:", erro)
					continue
				}
				defer resp.Body.Close()
			}
		}

		// notificar o sincronizador sobre as alterações
		if sincronizadorMQTT != nil {
			for _, id := range data.IDPostos {
				var posto modelo.Posto
				filterSingle := bson.M{"id": id, "servidorOrigem": dbServer.Nome}
				if err := dbServer.PostosCollection.FindOne(ctx, filterSingle).Decode(&posto); err == nil {
					sincronizadorMQTT.NotificarAlteracao(&posto)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Postos atualizados com sucesso"})
}
