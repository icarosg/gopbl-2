package main

import (
	"time"
)

type Payment struct {
	PaymentID int       `json:"payment_id"`
	From      int       `json:"from"` // CarID que pagou a estação
	To        int       `json:"to"`   // StationID que recebeu o pagamento
	Value     int       `json:"value"`
	TimeStamp time.Time `json:"timestamp"`
}
