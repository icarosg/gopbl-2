package db

import (
	"context"
	"encoding/json"
	"fmt"
	"gopbl-2/modelo"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TopicPrefix          = "sincronizacao/postos/"
	TopicPostoAtualizado = TopicPrefix + "atualizado"
	TopicPostoRemovido   = TopicPrefix + "removido"
	TopicSolicitacao     = TopicPrefix + "solicitacao"
)

// gerencia a sincronização via MQTT entre diferentes bases de dados
type SincronizadorMQTT struct {
	Conexao       *ConexaoServidorDB
	Cliente       mqtt.Client
	ServidorID    string
	IntervaloSync time.Duration
}

func NovoSincronizadorMQTT(conexao *ConexaoServidorDB, brokerURL, servidorID string, intervalo time.Duration) (*SincronizadorMQTT, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(fmt.Sprintf("sincronizador-%s", servidorID)).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(1 * time.Second).
		SetCleanSession(false).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(30 * time.Second)

	cliente := mqtt.NewClient(opts)
	token := cliente.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("erro ao conectar ao broker MQTT: %v", token.Error())
	}
	sincronizador := &SincronizadorMQTT{
		Conexao:       conexao,
		Cliente:       cliente,
		ServidorID:    servidorID,
		IntervaloSync: intervalo,
	}

	configurarHandlers(sincronizador)

	return sincronizador, nil
}

// configurarHandlers configura os callbacks para os tópicos MQTT
func configurarHandlers(s *SincronizadorMQTT) {
	// Handler para postos atualizados
	s.Cliente.Subscribe(TopicPostoAtualizado, 1, func(client mqtt.Client, msg mqtt.Message) {
		var posto modelo.Posto
		if err := json.Unmarshal(msg.Payload(), &posto); err != nil {
			log.Printf("Erro ao decodificar posto atualizado: %v", err)
			return
		}
		s.atualizarPostoLocal(&posto)
	})

	// Handler para solicitações de sincronização completa
	s.Cliente.Subscribe(TopicSolicitacao, 1, func(client mqtt.Client, msg mqtt.Message) {
		var solicitacao struct {
			ServidorID string `json:"servidorID"`
		}
		if err := json.Unmarshal(msg.Payload(), &solicitacao); err != nil {
			log.Printf("Erro ao decodificar solicitação: %v", err)
			return
		}

		if solicitacao.ServidorID != s.ServidorID { // evita responder a propria solicitação
			s.enviarTodosPostos()
		}
	})
}

// inicia o processo de sincronização periódica via MQTT
func (s *SincronizadorMQTT) IniciarSincronizacao() {
	s.SolicitarSincronizacao()

	// timer para sincronização periódica
	go func() {
		ticker := time.NewTicker(s.IntervaloSync)
		defer ticker.Stop()

		for range ticker.C {
			s.SolicitarSincronizacao()
		}
	}()
}

// solicita a todos os servidores que enviem seus postos
func (s *SincronizadorMQTT) SolicitarSincronizacao() {
	solicitacao := struct {
		ServidorID string `json:"servidorID"`
		Timestamp  int64  `json:"timestamp"`
	}{
		ServidorID: s.ServidorID,
		Timestamp:  time.Now().UnixNano(),
	}

	payload, err := json.Marshal(solicitacao)
	if err != nil {
		log.Printf("Erro ao serializar solicitação: %v", err)
		return
	}

	// publica no tópico de solicitação
	token := s.Cliente.Publish(TopicSolicitacao, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Erro ao publicar solicitação: %v", token.Error())
	}
}

func (s *SincronizadorMQTT) NotificarAlteracao(posto *modelo.Posto) {
	payload, err := json.Marshal(posto)
	if err != nil {
		log.Printf("Erro ao serializar posto: %v", err)
		return
	}

	// publica no tópico de atualização
	token := s.Cliente.Publish(TopicPostoAtualizado, 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Erro ao publicar atualização: %v", token.Error())
	}
}

// atualiza um posto no banco de dados local
func (s *SincronizadorMQTT) atualizarPostoLocal(posto *modelo.Posto) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// se o posto não pertence a este servidor, apenas atualizamos a disponibilidade
	if posto.ServidorOrigem != s.ServidorID {
		filtro := bson.M{"id": posto.ID}

		var postoExistente modelo.Posto
		err := s.Conexao.PostosCollection.FindOne(ctx, filtro).Decode(&postoExistente)

		if err != nil { // se o posto não existe, retorna nada
			return
		}

		// atualizar apenas a disponibilidade
		update := bson.M{"$set": bson.M{
			"disponivel":        posto.Disponivel,
			"ultimaAtualizacao": posto.UltimaAtualizacao,
		}}

		_, err = s.Conexao.PostosCollection.UpdateOne(ctx, filtro, update)
		if err != nil {
			log.Printf("Erro ao atualizar disponibilidade do posto %s: %v", posto.ID, err)
		} else {
			log.Printf("Disponibilidade do posto %s atualizada localmente", posto.ID)
		}
		return
	}

	// se o posto pertence a este servidor atualiza normalmente
	filtro := bson.M{"id": posto.ID}
	opcoes := options.Update().SetUpsert(true)
	update := bson.M{"$set": posto}

	_, err := s.Conexao.PostosCollection.UpdateOne(ctx, filtro, update, opcoes)
	if err != nil {
		log.Printf("Erro ao atualizar posto local: %v", err)
	} else {
		log.Printf("Posto %s atualizado localmente", posto.ID)
	}
}

// envia todos os postos locais para sincronização
func (s *SincronizadorMQTT) enviarTodosPostos() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// buscar apenas os postos deste servidor
	filtro := bson.M{"servidorOrigem": s.ServidorID}
	cursor, err := s.Conexao.PostosCollection.Find(ctx, filtro)
	if err != nil {
		log.Printf("Erro ao buscar postos locais: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var postos []modelo.Posto
	if err := cursor.All(ctx, &postos); err != nil {
		log.Printf("Erro ao decodificar postos locais: %v", err)
		return
	}

	for i := range postos {
		s.NotificarAlteracao(&postos[i])
		time.Sleep(50 * time.Millisecond)
	}
}

// encerra a conexão com o broker MQTT
func (s *SincronizadorMQTT) Fechar() {
	if s.Cliente.IsConnected() {
		s.Cliente.Unsubscribe(TopicPostoAtualizado, TopicSolicitacao)
		s.Cliente.Disconnect(250)
	}
}
