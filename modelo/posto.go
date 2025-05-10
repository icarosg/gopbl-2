package modelo

import (
	"fmt"
	//"math"
	//"sort"
	//"sync"
	"time"
)

type Posto struct {
	ID           string
	Latitude     float64
	Longitude    float64
	//mu           sync.Mutex
	Fila         Veiculo
	QtdFila      int
	BombaOcupada bool
	UltimaAtualizacao time.Time `bson:"ultimaAtualizacao" json:"ultimaAtualizacao"`
	ServidorOrigem    string    `bson:"servidorOrigem" json:"servidorOrigem"`
}

func NovoPosto(id string, lat float64, long float64) Posto {
	fmt.Printf("Posto %s criado na localização (%.6f, %.6f)",
		id, lat, long)

	return Posto{
		ID:           id,
		Latitude:     lat,
		Longitude:    long,
		QtdFila:      0,
		BombaOcupada: false,
	}
}

func (p Posto) String() string {
	return fmt.Sprintf("Posto id:%s\n Servidor de Origem:%s\n Posto localizado em (%.4f, %.4f)\n bomba ocupada:  %t\n", p.ID, p.ServidorOrigem, p.Latitude, p.Longitude,p.BombaOcupada)
}

