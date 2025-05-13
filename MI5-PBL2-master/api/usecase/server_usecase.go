package usecase

import (
	"api/model"
	"api/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ServerUsecase struct {
	serverRepo repository.ServerRepository
}

func NewServerUsecase(serverRepo repository.ServerRepository) ServerUsecase {
	return ServerUsecase{
		serverRepo: serverRepo,
	}
}

// Registra ou atualiza um servidor
func (su *ServerUsecase) RegisterOrUpdateServer(ctx context.Context, company string, serverIP string) error {
	err = su.serverRepo.RegisterOrUpdateServer(ctx, company, serverIP)
}

// Obtém a lista de servidores registrados
func (su *ServerUsecase) GetRegisteredServers(ctx context.Context) ([]model.Server, error) {
	ips, err := su.serverRepo.GetRegisteredServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter servidores registrados: %w", err)
	}

	servers := []model.Server{}
	for _, ip := range ips {
		servers = append(servers, model.Server{
			ServerIP: ip,
		})
	}
	return servers, nil
}

// Remove servidores inativos
func (su *ServerUsecase) RemoveInactiveServers(ctx context.Context, threshold time.Duration) error {
	return su.serverRepo.RemoveInactiveServers(ctx, threshold)
}

// Consulta estações disponíveis em outro servidor
func (su *ServerUsecase) GetStationsFromServer(url string) ([]model.Station, error) {
	// Faz a requisição HTTP
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição para o servidor remoto: %w", err)
	}
	defer resp.Body.Close()

	// Lê o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta do servidor remoto: %w", err)
	}

	// Estrutura auxiliar para deserializar a resposta
	var response struct {
		Stations []model.Station `json:"stations"`
	}

	// Deserializa o JSON
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("erro ao deserializar a resposta do servidor remoto: %w", err)
	}

	return response.Stations, nil
}

// Reserva uma estação em outro servidor
func (su *ServerUsecase) ReserveStationOnServer(serverURL string, stationID int, carID int) error {
	// Cria o payload da requisição
	payload := map[string]interface{}{
		"station_id": stationID,
		"car_id":     carID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	// Constrói a URL do endpoint remoto
	url := fmt.Sprintf("%s/stations/reserve", serverURL)

	// Faz a requisição HTTP POST
	resp, err := http.Post(url, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("erro ao fazer requisição para reservar estação: %w", err)
	}
	defer resp.Body.Close()

	// Verifica o status da resposta
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Lê o corpo da resposta para depuração
		return fmt.Errorf("falha ao reservar estação, status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
