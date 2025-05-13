package controller

import (
	"api/usecase"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ServerController struct {
	serverUsecase usecase.ServerUsecase
}

func NewServerController(usecase usecase.ServerUsecase) ServerController {
	return ServerController{
		serverUsecase: usecase,
	}
}

func (sc *ServerController) RegisterServer(ctx *gin.Context) {
	var request struct {
		Company  string `json:"company"`
		ServerIP string `json:"server_ip"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := sc.serverUsecase.RegisterOrUpdateServer(context.Background(), request.Company, request.ServerIP)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Servidor registrado com sucesso"})
}

// Endpoint para obter a lista de servidores registrados
func (sc *ServerController) GetRegisteredServers(ctx *gin.Context) {
	servers, err := sc.serverUsecase.GetRegisteredServers(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"servers": servers})
}

// Endpoint para remover servidores inativos
func (sc *ServerController) RemoveInactiveServers(ctx *gin.Context) {
	threshold := 10 * time.Minute // Exemplo: servidores inativos por mais de 10 minutos
	err := sc.serverUsecase.RemoveInactiveServers(context.Background(), threshold)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Servidores inativos removidos com sucesso"})
}

func (sc *ServerController) GetStationsFromServer(ctx *gin.Context) {
	serverURL := ctx.Query("server_url")
	if serverURL == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "server_url is required"})
		return
	}

	stations, err := sc.serverUsecase.GetStationsFromServer(serverURL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"stations": stations})
}

func (sc *ServerController) ReserveStationOnServer(ctx *gin.Context) {
	// Captura os parâmetros da query string
	serverURL := ctx.Query("server_url")
	stationID := ctx.Query("station_id")
	carID := ctx.Query("car_id")

	// Valida os parâmetros
	if serverURL == "" || stationID == "" || carID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "server_url, station_id, and car_id are required"})
		return
	}

	// Converte stationID e carID para inteiros
	stationIDInt, err := strconv.Atoi(stationID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "station_id must be an integer"})
		return
	}

	carIDInt, err := strconv.Atoi(carID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "car_id must be an integer"})
		return
	}

	// Chama o usecase para realizar a reserva
	err = sc.serverUsecase.ReserveStationOnServer(serverURL, stationIDInt, carIDInt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Station reserved successfully"})
}
