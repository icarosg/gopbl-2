package repository

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson"
// )

// // UpdateCarBattery altera o battery_level de um carro
// // func UpdateCarBattery(carID, newLevel int) error {
// // 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// // 	defer cancel()

// // 	filter := bson.M{"car_id": carID}
// // 	update := bson.M{"$set": bson.M{"battery_level": newLevel}}

// // 	res, err := GetCarCollection().UpdateOne(ctx, filter, update)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	if res.MatchedCount == 0 {
// // 		return fmt.Errorf("car_id %d não encontrado", carID)
// // 	}
// // 	fmt.Printf("🔋 Bateria do carro %d atualizada para %d%%\n", carID, newLevel)
// // 	return nil
// // }
