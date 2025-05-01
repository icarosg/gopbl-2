package modelo

// definir rotas para: consulta de pontos de recarga disponíveis;
// reserva de pontos de recarga;
// registro de recargas realizadas.

import (
	//"fmt"
	"math/rand"
)

type Veiculo struct {
	ID                  string
	Latitude            float64
	Longitude           float64
	Bateria             float64
	IsCarregando        bool
	IsDeslocandoAoPosto bool
	Pagamentos          []Pagamento
}

func NovoVeiculo(id string, inicialLat float64, inicialLong float64) Veiculo {
	return Veiculo{
		ID:                  id,
		Latitude:            inicialLat,
		Longitude:           inicialLong,
		Bateria:             100.0, // começa com bateria cheia
		IsCarregando:        false,
		IsDeslocandoAoPosto: false,
	}
}

// func adicionarPagamento(v *Veiculo, p PagamentoJson) {

// 	*v.Pagamentos = append(*v.Pagamentos, p)

// }

func AtualizarLocalizacao(v *Veiculo) {
	if !v.IsDeslocandoAoPosto && !v.IsCarregando {
		v.Latitude += float64(rand.Intn(11) - 5) //valor entre 5 e -5
		v.Longitude += float64(rand.Intn(11) - 5)

		// fmt.Println("_________________________________________________________________________________________________")
		// fmt.Printf("localalizacao atual do veiculo: latitude %.4f e longitude %.4f\n", v.Latitude, v.Longitude)
		// fmt.Println("_________________________________________________________________________________________________")
		DiminuirNivelBateria(v)
	}
}

func DiminuirNivelBateria(v *Veiculo) {
	if !v.IsCarregando {
		// diminui a bateria entre 3.0 e 1.0 por atualização
		v.Bateria -= rand.Float64()*3.0 + 1.0
		if v.Bateria < 10 {
			v.Bateria = 10
		}
	}
}

func DeslocarParaPosto(v *Veiculo, p *Posto) {
	if v.Latitude < p.Latitude {
		if p.Latitude-v.Latitude <= 5 {
			v.Latitude = p.Latitude
		} else {
			v.Latitude += 5
		}
	} else if v.Latitude > p.Latitude {
		if v.Latitude-p.Latitude <= 5 {
			v.Latitude = p.Latitude
		} else {
			v.Latitude -= 5
		}
	}

	if v.Longitude < p.Longitude {
		if p.Longitude-v.Longitude <= 5 {
			v.Longitude = p.Longitude
		} else {
			v.Longitude += 5
		}
	} else if v.Longitude > p.Longitude {
		if v.Longitude-p.Longitude <= 5 {
			v.Longitude = p.Longitude
		} else {
			v.Longitude -= 5
		}
	}

	if !v.IsCarregando {
		// diminui a bateria entre 3.0 e 1.0 por atualização
		v.Bateria -= rand.Float64()*3.0 + 1.0
		if v.Bateria < 10 {
			v.Bateria = 10
		}
	}
}

func GetNivelBateriaAoChegarNoPosto(v Veiculo, p *Posto) float64 {
	for v.Latitude != p.Latitude || v.Longitude != p.Longitude {
		if v.Latitude < p.Latitude {
			if p.Latitude-v.Latitude <= 5 {
				v.Latitude = p.Latitude
			} else {
				v.Latitude += 5
			}
		} else if v.Latitude > p.Latitude {
			if v.Latitude-p.Latitude <= 5 {
				v.Latitude = p.Latitude
			} else {
				v.Latitude -= 5
			}
		}

		if v.Longitude < p.Longitude {
			if p.Longitude-v.Longitude <= 5 {
				v.Longitude = p.Longitude
			} else {
				v.Longitude += 5
			}
		} else if v.Longitude > p.Longitude {
			if v.Longitude-p.Longitude <= 5 {
				v.Longitude = p.Longitude
			} else {
				v.Longitude -= 5
			}
		}

		if !v.IsCarregando {
			// diminui a bateria entre 3.0 e 1.0 por atualização
			v.Bateria -= rand.Float64()*3.0 + 1.0
			if v.Bateria < 10 {
				v.Bateria = 10
			}
		}
	}

	return v.Bateria
}