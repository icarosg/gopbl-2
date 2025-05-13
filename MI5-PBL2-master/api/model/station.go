package model

type Station struct {
	StationID int    `bson:"station_id"`
	Company   string `bson:"company"` // Nome da empresa
	InUseBy   int    `bson:"in_use"`  // CarID
}
