package types

import (
	"fmt"
	"math/rand"
)

type Car struct {
	CarID            int   `bson:"car_id"`
	BatteryLevel     int   `bson:"battery_level"`      // 0-100%
	BatteryDrainRate int   `bson:"battery_drain_rate"` // % por segundo
	ReservedRoute    Route `bson:"reserved_route"`     // Rota reservada
}

// Functions
func GetNewRandomCar() Car {
	fmt.Println("Insira o ID do carro:")
	carID := 0
	fmt.Scanln(&carID)
	// Criar um novo carro com as coordenadas geradas
	return Car{
		CarID:            carID,              // Exemplo: ID do carro entre 0 e 999
		BatteryLevel:     rand.Intn(50) + 50, // Bateria cheia inicialmente (50-100%)
		BatteryDrainRate: rand.Intn(3) + 1,   // % por KM (1-3%)
		ReservedRoute:    Route{},
	}
}

func (c *Car) PrintState(paymentID int) {
	println("Car ID:", c.CarID)
	println("Battery Level:", c.BatteryLevel)
	println("Battery Drain Rate:", c.BatteryDrainRate)
	c.ReservedRoute.PrintRoute()
}

func (c *Car) GetCarID() int {
	return c.CarID
}

func (c *Car) GetBatteryLevel() int {
	return c.BatteryLevel
}

func (c *Car) GetBatteryDrainRate() int {
	return c.BatteryDrainRate
}

func (c *Car) GetReservedRoute() Route {
	return c.ReservedRoute
}

func (c *Car) SetReservedRoute(route Route) {
	c.ReservedRoute = route
}

func (c *Car) SetBatteryLevel(batteryLevel int) {
	c.BatteryLevel = batteryLevel
}

func (c *Car) SetBatteryDrainRate(batteryDrainRate int) {
	c.BatteryDrainRate = batteryDrainRate
}

func (c *Car) SetCarID(carID int) {
	c.CarID = carID
}
