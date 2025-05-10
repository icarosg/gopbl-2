package models

type Posto struct {
	ID        string  `json:"id" bson:"_id"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	BombaOcupada bool    `json:"reservado" bson:"reservado"`
}

type ReservaData struct {
	IDPostos []string `json:"idPostos"`
	Reservar bool     `json:"reservar"`
}