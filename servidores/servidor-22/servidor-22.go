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

	//mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"

	"bytes"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
)

var servidores = []string{
	"http://localhost:8080",
	"http://localhost:8082",
	"http://localhost:8084",
}

var dbServer *db.ConexaoServidorDB
var sincronizadorMQTT *db.SincronizadorMQTT

func main() {
	hostDB := getEnv("DB_HOST", "localhost")
	portaDB := 27017
	nomeServidor := "22"
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

	router := gin.Default() // cria o servidor Gin

	// registra rota
	router.GET("/postosDisponiveis", postosDisponiveisHandler)
	router.POST("/cadastrar", cadastrarPostoHandler)
	router.PUT("/reservar", editarPostoHandler)

	router.Run(":8082")
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
			if servidor == "http://localhost:8082" {
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
				if servidor == "http://localhost:8082" {
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
