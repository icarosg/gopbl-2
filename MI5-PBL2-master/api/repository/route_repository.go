package repository

import (
	"api/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type RouteRepository struct {
	collection *mongo.Collection
}

func NewRouteRepository(db *mongo.Database) *RouteRepository {
	return &RouteRepository{
		collection: db.Collection("routes"),
	}
}

func (r *RouteRepository) CreateRoute(route *model.Route) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, route)
	return err
}

// Busca todas as rotas entre duas cidades
func (r *RouteRepository) GetRoutesBetweenCities(startCity, endCity string) ([]model.Route, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"start_city": startCity, "end_city": endCity}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var routes []model.Route
	if err := cursor.All(ctx, &routes); err != nil {
		return nil, err
	}

	return routes, nil
}
