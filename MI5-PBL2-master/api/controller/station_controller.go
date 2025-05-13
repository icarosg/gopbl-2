package controller

import (
	"api/model"
	"api/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StationController struct {
	stationUsecase usecase.StationUsecase
}

func NewStationController(usecase usecase.StationUsecase) StationController {
	return StationController{
		stationUsecase: usecase,
	}
}

func (sc *StationController) CreateStation(ctx *gin.Context) {

	station := model.Station{}

	err := ctx.BindJSON(&station)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	station, err = sc.stationUsecase.CreateStation(station)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, station)
}

func (sc *StationController) GetAllStations(ctx *gin.Context) {
	company := ctx.Query("company") // Obtém o filtro de empresa da query string

	stations, err := sc.stationUsecase.GetAllStations(ctx, company)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"stations": stations})
}

func (sc *StationController) ReserveStation(ctx *gin.Context) {
	var request struct {
		StationID int `json:"station_id"`
		CarID     int `json:"car_id"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := sc.stationUsecase.ReserveStation(ctx, request.StationID, request.CarID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Estação reservada com sucesso"})
}
