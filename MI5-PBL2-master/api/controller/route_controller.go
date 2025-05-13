package controller

import (
	"api/model"
	"api/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouteController struct {
	routeUsecase usecase.RouteUsecase
}

// Construtor para o RouteController
func NewRouteController(usecase usecase.RouteUsecase) RouteController {
	return RouteController{
		routeUsecase: usecase,
	}
}

// Endpoint para criar uma nova rota
func (rc *RouteController) CreateRoute(ctx *gin.Context) {
	var route model.Route

	// Faz o bind do JSON recebido para o modelo Route
	if err := ctx.BindJSON(&route); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Chama o usecase para criar a rota
	err := rc.routeUsecase.CreateRoute(&route)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Retorna a rota criada
	ctx.JSON(http.StatusOK, route)
}

// Endpoint para buscar todas as rotas com base na cidade de origem e destino final
func (rc *RouteController) GetRoutes(ctx *gin.Context) {
	startCity := ctx.Query("start_city") // Obtém a cidade de origem da query string
	endCity := ctx.Query("end_city")     // Obtém a cidade de destino da query string

	// Valida se os parâmetros foram fornecidos
	if startCity == "" || endCity == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_city e end_city são obrigatórios"})
		return
	}

	// Chama o usecase para buscar as rotas
	routes, err := rc.routeUsecase.GetRoutesBetweenCities(startCity, endCity)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Retorna as rotas encontradas
	ctx.JSON(http.StatusOK, gin.H{"routes": routes})
}
