package model

import "time"

type Server struct {
	ServerIP  string    `bson:"server_ip"` // IP do servidor
	Company   string    `bson:"company"`   // Nome da empresa
	Timestamp time.Time `bson:"timestamp"` // Última atualização
}
