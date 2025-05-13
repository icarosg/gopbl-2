package main

import (
	"api/controller"
	"api/database"
	"api/mqtt_server"
	"api/repository"
	"api/usecase"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Conecta ao MongoDB
	database.ConnectDB()
	defer database.DisconnectDB()

	// Usa o banco de dados dinâmico
	db := database.Database

	// Popula o banco de dados com rotas
	database.SeedRoutes(db)

	// Configura o repositório
	stationRepo := repository.NewStationRepository(db)
	serverRepo := repository.NewServerRepository(db)
	routeRepo := repository.NewRouteRepository(db)

	// Configura o usecase
	routeUsecase := usecase.NewRouteUsecase(routeRepo)
	stationUsecase := usecase.NewStationUseCase(stationRepo)
	serverUsecase := usecase.NewServerUsecase(serverRepo)

	// Configura o controlador
	routeController := controller.NewRouteController(routeUsecase)
	stationController := controller.NewStationController(stationUsecase)
	serverController := controller.NewServerController(serverUsecase)

	server := gin.Default()
	// Rotas relacionadas às estações
	server.POST("/stations", stationController.CreateStation)
	server.GET("/stations", stationController.GetAllStations)

	// Rotas relacionadas à comunicação entre servidores
	server.POST("/servers/register", serverController.RegisterServer)
	server.GET("/servers", serverController.GetRegisteredServers)
	server.DELETE("/servers/inactive", serverController.RemoveInactiveServers)

	server.GET("/server/stations", serverController.GetStationsFromServer)
	server.POST("/server/reserve", serverController.ReserveStationOnServer)
	server.POST("/stations/reserve", stationController.ReserveStation)

	// Rotas relacionadas às rotas
	server.POST("/routes", routeController.CreateRoute)
	server.GET("/routes", routeController.GetRoutes)

	// Inicia o servidor MQTT
	go mqtt_server.MqttMain()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback para a porta 8080 se a variável de ambiente não estiver configurada
	}
	server.Run(fmt.Sprintf(":%s", port))
}
