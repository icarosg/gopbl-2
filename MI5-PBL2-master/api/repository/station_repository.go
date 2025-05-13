package repository

import (
	"api/model"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StationRepository struct {
	collection *mongo.Collection
}

// mudar a collection para o mongo
func NewStationRepository(db *mongo.Database) StationRepository {
	return StationRepository{
		collection: db.Collection("stations"),
	}
}
func (sr *StationRepository) CreateStation(station model.Station) (int, error) {
	// Insere a estação na coleção
	_, err := sr.collection.InsertOne(context.TODO(), station)
	if err != nil {
		return 0, fmt.Errorf("erro ao criar estação: %w", err)
	}

	// Retorna o ID da estação
	return station.StationID, nil
}

func (sr *StationRepository) GetAllStations(ctx context.Context, company string) ([]model.Station, error) {
	// Define o filtro para a consulta
	filter := bson.M{}
	if company != "" {
		filter["company"] = company
	}

	// Realiza a consulta no MongoDB
	cursor, err := sr.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar estações: %w", err)
	}
	defer cursor.Close(ctx)

	// Decodifica todos os documentos encontrados
	var stations []model.Station
	if err := cursor.All(ctx, &stations); err != nil {
		return nil, fmt.Errorf("erro ao decodificar estações: %w", err)
	}

	return stations, nil
}
func (sr *StationRepository) ClearStations(ctx context.Context) error {
	err := sr.collection.Drop(ctx)
	if err != nil {
		return fmt.Errorf("erro ao limpar a coleção de estações: %w", err)
	}
	return nil
}
func (sr *StationRepository) UpdateStation(ctx context.Context, station model.Station) error {
	filter := bson.M{"station_id": station.StationID}
	update := bson.M{"$set": station}

	_, err := sr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar estação: %w", err)
	}

	return nil
}
