package usecase

import (
	"api/model"
	"api/repository"
)

type RouteUsecase struct {
	routeRepo repository.RouteRepository
}

func NewRouteUsecase(routeRepo *repository.RouteRepository) RouteUsecase {
	return RouteUsecase{
		routeRepo: *routeRepo, // Desreferencia o ponteiro
	}
}

func (ru *RouteUsecase) CreateRoute(route *model.Route) error {
	return ru.routeRepo.CreateRoute(route)
}

func (ru *RouteUsecase) GetRoutesBetweenCities(startCity, endCity string) ([]model.Route, error) {
	return ru.routeRepo.GetRoutesBetweenCities(startCity, endCity)
}
