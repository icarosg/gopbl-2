package types

import (
	"fmt"
)

type Topics int

const (
	Consult Topics = iota
	Reserve
	Select
	Birth
	Death
)

var TopicNames = map[Topics]string{
	Consult: "consult",
	Reserve: "reserve",
	Select:  "select",
	Birth:   "birth",
	Death:   "death",
}

func (t Topics) String() string {
	return TopicNames[t]
}

type MqttClientTypes int

const (
	StationClientType MqttClientTypes = iota
	CarClientType
	CompanyClientType
)

var MqttClientTypeNames = map[MqttClientTypes]string{
	StationClientType: "station",
	CarClientType:     "car",
	CompanyClientType: "company",
}

func (m MqttClientTypes) String() string {
	return MqttClientTypeNames[m]
}

type MQTT_Message struct {
	Topic   string `json:"topic"`
	Message []byte `json:"message"`
}

type CarInfo struct {
	CarId int `json:"car_id"`
}

type RoutesMessage struct {
	City1 string `json:"city1"`
	City2 string `json:"city2"`
}

type SelectRouteMessage struct {
	Car   Car   `json:"car"`
	Route Route `json:"route"`
}

type RoutesList struct {
	Routes []Route `json:"routes"`
}

// STATION TOPICS
func StationBirthTopic(serverIP string) string {
	// Birth of a station in serverIP
	return Birth.String() + StationClientType.String() + serverIP
}

func StationDeathTopic(serverIP string) string {
	// Death of a station in serverIP
	return Death.String() + StationClientType.String() + serverIP
}

func StationConsultTopic(serverIP string, stationID int) string {
	// Consult a station in serverIP with stationID
	return Consult.String() + StationClientType.String() + serverIP + fmt.Sprintf("%d", stationID)
}

func StationReserveTopic(serverIP string, stationID int) string {
	// Reserve a station in serverIP with stationID
	return Reserve.String() + StationClientType.String() + serverIP + fmt.Sprintf("%d", stationID)
}

// CAR TOPICS
func CarBirthTopic(serverIP string) string {
	// Birth of a Car in serverIP
	return Birth.String() + CarClientType.String() + serverIP
}

func CarDeathTopic(serverIP string) string {
	// Death of a Car in serverIP
	return Death.String() + CarClientType.String() + serverIP
}

func CarConsultTopic(serverIP string, CarID int) string {
	// Consult a Car in serverIP with CarID
	return Consult.String() + CarClientType.String() + serverIP + fmt.Sprintf("%d", CarID)
}

func CarReserveTopic(serverIP string, CarID int) string {
	// Reserve a Car in serverIP with CarID
	return Reserve.String() + CarClientType.String() + serverIP + fmt.Sprintf("%d", CarID)
}

func CarSelectRouteTopic(serverIP string, CarID int) string {
	// Select a route for a Car in serverIP with CarID
	return Select.String() + CarClientType.String() + serverIP + fmt.Sprintf("%d", CarID)
}
