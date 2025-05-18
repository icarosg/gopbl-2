package main

import (
	"gopbl-2/db"
	"gopbl-2/modelo"
	"gopbl-2/models"

	//"gopbl-2/models"

	"context"
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	//"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"

	"bytes"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
)

var servidores = []string{
	"http://172.18.0.1:8083", //22
	"http://172.18.0.1:8085", //shell
	"http://172.18.0.1:8084", //ipiranga
}

var dbServer *db.ConexaoServidorDB
var sincronizadorMQTT *db.SincronizadorMQTT
var mqttClient mqtt.Client

func main() {
	//hostDB := getEnv("DB_HOST", "172.16.103.13")
	portaDB := 27017
	nomeServidor := "Ipiranga"
	//mqttBroker := getEnv("MQTT_BROKER", "tcp://172.18.0.1:1883")

	var erro error
	dbServer, erro = db.NovaConexaoDB(nomeServidor, "172.18.0.1", portaDB)
	if erro != nil {
		log.Fatal("Erro ao conectar ao MongoDB:", erro)
	}
	fmt.Println("Conectado ao MongoDB com sucesso!")

	configurarMQTT() // inicia conexão MQTT

	router := gin.Default() // cria o servidor Gin

	router.GET("/postosDisponiveis", postosDisponiveisHandler)
	router.POST("/cadastrar", cadastrarPostoHandler)
	router.PUT("/reservar", editarPostoHandler)

	router.Run(":8084")
}

func configurarMQTT() {
	// Configurar cliente MQTT
	opts := mqtt.NewClientOptions().AddBroker("tcp://172.18.0.1:1884")
	opts.SetClientID("servidor-ipiranga-mqtt")
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Erro ao conectar ao broker MQTT: %v", token.Error())
		return
	}
	fmt.Println("Conectado ao broker MQTT com sucesso!")

	nomeServidor := "Ipiranga"

	// Inscrever nos tópicos gerais para manter compatibilidade com clientes existentes
	mqttClient.Subscribe(modelo.TopicPostosDisponiveis, 1, handleListarPostos)
	mqttClient.Subscribe(modelo.TopicCadastrarPosto, 1, handleCadastrarPosto)
	mqttClient.Subscribe(modelo.TopicReservarPosto, 1, handleReservarPosto)

	// Inscrever nos tópicos específicos para este servidor
	topicDisponiveis := modelo.GetTopicServidor(nomeServidor, "disponiveis")
	topicCadastrar := modelo.GetTopicServidor(nomeServidor, "cadastrar")
	topicReservar := modelo.GetTopicServidor(nomeServidor, "reservar")

	mqttClient.Subscribe(topicDisponiveis, 1, handleListarPostos)
	mqttClient.Subscribe(topicCadastrar, 1, handleCadastrarPosto)
	mqttClient.Subscribe(topicReservar, 1, handleReservarPosto)

	fmt.Printf("Servidor inscrito em tópicos específicos: %s, %s, %s\n",
		topicDisponiveis, topicCadastrar, topicReservar)
}

func handleListarPostos(client mqtt.Client, msg mqtt.Message) {
	// Extrair o ID do cliente da mensagem
	var request struct {
		ClientID string `json:"clientId"`
	}

	if err := json.Unmarshal(msg.Payload(), &request); err != nil {
		log.Printf("Erro ao decodificar solicitação: %v", err)
		return
	}

	// Se não houver ID do cliente, não podemos responder
	if request.ClientID == "" {
		log.Printf("Solicitação sem ID do cliente")
		return
	}

	// Definir o tópico de resposta específico para este cliente
	responseTopic := modelo.TopicResposta + "/" + request.ClientID

	// Consultar diretamente do banco de dados e enviar via MQTT
	collection := dbServer.PostosCollection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"disponivel": true}
	cursor, erro := collection.Find(ctx, filter)
	if erro != nil {
		log.Printf("Erro ao buscar postos disponíveis: %v", erro)
		return
	}
	defer cursor.Close(ctx)

	var postosLocais []modelo.Posto
	if erro = cursor.All(ctx, &postosLocais); erro != nil {
		log.Printf("Erro ao decodificar os dados: %v", erro)
		return
	}

	// Se solicitado, consulta outros servidores
	var todosPostos []modelo.Posto
	todosPostos = append(todosPostos, postosLocais...)

	for _, servidor := range servidores {
		if servidor == "http://172.18.0.1:8084" {
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

	// Enviar a resposta via MQTT no tópico específico do cliente
	payload, err := json.Marshal(todosPostos)
	if err != nil {
		log.Printf("Erro ao codificar postos: %v", err)
		return
	}

	token := mqttClient.Publish(responseTopic, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Erro ao publicar resposta: %v", token.Error())
	}
}

func handleCadastrarPosto(client mqtt.Client, msg mqtt.Message) {
	var novoPosto modelo.Posto
	if err := json.Unmarshal(msg.Payload(), &novoPosto); err != nil {
		log.Printf("Erro ao decodificar posto: %v", err)
		return
	}

	log.Printf("Recebida solicitação para cadastrar posto: %s no servidor %s", novoPosto.ID, novoPosto.ServidorOrigem)

	// Definir tópico de resposta para este posto/cliente
	responseTopic := modelo.TopicResposta + "/" + novoPosto.ID
	log.Printf("Usando tópico de resposta: %s", responseTopic)

	// Garante que o timestamp está atualizado e definir o servidor de origem
	novoPosto.UltimaAtualizacao = time.Now()
	novoPosto.ServidorOrigem = dbServer.Nome

	collection := dbServer.PostosCollection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": novoPosto.ID}
	var existente modelo.Posto
	erro := collection.FindOne(ctx, filter).Decode(&existente) // verifica se já existe um posto com mesmo nome

	// Preparar resposta
	resposta := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{}

	if erro == nil {
		log.Printf("Posto %s já existe", novoPosto.ID)
		resposta.Status = "erro"
		resposta.Message = "Posto já existente"
	} else {
		_, erro = collection.InsertOne(ctx, novoPosto)
		if erro != nil {
			log.Printf("Erro ao inserir posto: %v", erro)
			resposta.Status = "erro"
			resposta.Message = "Erro ao cadastrar posto"
		} else {
			log.Printf("Posto %s cadastrado com sucesso", novoPosto.ID)
			resposta.Status = "sucesso"
			resposta.Message = "Posto cadastrado com sucesso"
		}
	}

	// Enviar resposta com QoS 1 para garantir entrega pelo menos uma vez
	responsePayload, _ := json.Marshal(resposta)
	log.Printf("Enviando resposta para %s: %s", responseTopic, string(responsePayload))

	token := mqttClient.Publish(responseTopic, 1, false, responsePayload)
	if token.WaitTimeout(5*time.Second) && token.Error() != nil {
		log.Printf("Erro ao enviar resposta: %v", token.Error())
		// Tentar novamente uma vez
		token = mqttClient.Publish(responseTopic, 1, false, responsePayload)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Falha na segunda tentativa de enviar resposta: %v", token.Error())
		} else {
			log.Printf("Resposta enviada com sucesso na segunda tentativa")
		}
	} else {
		log.Printf("Resposta enviada com sucesso")
	}
}

func handleReservarPosto(client mqtt.Client, msg mqtt.Message) {
	var data models.ReservaData
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("Erro ao decodificar dados de reserva: %v", err)
		return
	}

	responseTopic := modelo.TopicResposta + "/" + data.ClientID

	disponibilidade := make(map[string]bool)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"id": bson.M{"$in": data.IDPostos}}
	cursor, err := dbServer.PostosCollection.Find(ctx, filter)
	if err == nil {
		var postos []modelo.Posto
		if err = cursor.All(ctx, &postos); err == nil {
			for _, posto := range postos {
				disponibilidade[posto.ID] = posto.Disponivel
			}
		}
	}
	cursor.Close(ctx)

	// consulta disponibilidade em outros servidores
	for _, servidor := range servidores {
		if servidor == "http://172.18.0.1:8084" {
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
					disponibilidade[posto.ID] = posto.Disponivel
				}
			}
		}
	}

	if data.Reservar { // verifica disponibilidade de todos os postos
		for _, id := range data.IDPostos {
			if !disponibilidade[id] {
				log.Printf("Nem todos os postos estão disponíveis")
				// Enviar a resposta via MQTT no tópico específico do cliente
				reservaFalhou := true
				payload, err := json.Marshal(reservaFalhou)
				if err != nil {
					log.Printf("Erro ao falha de reserva: %v", err)
					return
				}

				token := mqttClient.Publish(responseTopic, 1, false, payload)
				token.Wait()
				if token.Error() != nil {
					log.Printf("Erro ao publicar resposta de falha de reserva: %v", token.Error())
				}
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
			log.Printf("Erro ao atualizar os postos: %v", err)
			return
		}

		//atualiza nos outros servidores
		for _, servidor := range servidores {
			if servidor == "http://172.18.0.1:8084" {
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

		// Enviar a resposta via MQTT no tópico específico do cliente
		reservaFalhou := false
		payload, err := json.Marshal(reservaFalhou)
		if err != nil {
			log.Printf("Erro ao falha de reserva: %v", err)
			return
		}

		token := mqttClient.Publish(responseTopic, 1, false, payload)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Erro ao publicar resposta de reserva concluida: %v", token.Error())
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
			log.Printf("Erro ao atualizar os postos: %v", err)
			return
		}

		//atualiza nos outros servidores
		for _, servidor := range servidores {
			if servidor == "http://172.18.0.1:8084" {
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

		// Enviar a resposta via MQTT no tópico específico do cliente
		reservaFalhou := false
		payload, err := json.Marshal(reservaFalhou)
		if err != nil {
			log.Printf("Erro ao codificar resposta de finalização de reserva: %v", err)
			return
		}

		token := mqttClient.Publish(responseTopic, 1, false, payload)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Erro ao publicar resposta de falha de finalizar reserva: %v", token.Error())
		}
	}
}

// obter variáveis de ambiente com valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func postosDisponiveisHandler(c *gin.Context) {
	if c.Query("consultarOutrosServidores") != "false" {
		// consulta outros servidores também

		var todosPostos []modelo.Posto

		for _, servidor := range servidores {
			if servidor == "http://172.18.0.1:8084" {
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

		filter := bson.M{"disponivel": true}
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

	filter := bson.M{"disponivel": true}
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
				disponibilidade[posto.ID] = posto.Disponivel
			}
		}
	}
	cursor.Close(ctx)

	// consulta disponibilidade em outros servidores
	for _, servidor := range servidores {
		if servidor == "http://172.18.0.1:8084" {
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
					disponibilidade[posto.ID] = posto.Disponivel
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
				if servidor == "http://172.18.0.1:8084" {
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
				if servidor == "http://172.18.0.1:8084" {
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
