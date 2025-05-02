package modelo

import (
	"fmt"
	//"math"
	//"sort"
	//"sync"
	//"time"
)

type Posto struct {
	ID           string
	Latitude     float64
	Longitude    float64
	//mu           sync.Mutex
	Fila         []*Veiculo
	QtdFila      int
	BombaOcupada bool
}

func NovoPosto(id string, lat float64, long float64) Posto {
	fmt.Printf("Posto %s criado na localização (%.6f, %.6f)",
		id, lat, long)

	return Posto{
		ID:           id,
		Latitude:     lat,
		Longitude:    long,
		Fila:         make([]*Veiculo, 0),
		QtdFila:      0,
		BombaOcupada: false,
	}
}

func (p Posto) String() string {
	return fmt.Sprintf("Posto id:%s\n Posto localizado em (%.4f, %.4f)\n ", p.ID, p.Latitude, p.Longitude)
}

