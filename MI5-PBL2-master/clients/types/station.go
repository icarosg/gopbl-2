package types

type Station struct {
	ID         int    `bson:"_id,omitempty"` // ID gerado pelo MongoDB
	ServerIP   string `bson:"server_ip"`     // IP do servidor
	ReservedBy int    `bson:"reserved_by"`   // ID do cliente carro que reservou o posto
}
