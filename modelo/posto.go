package modelo

import (
	"fmt"
	"math"
	"sort"
	//"sync"
	"time"
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

func ReservarVaga(p *Posto, v *Veiculo) bool {
	// retorna true se caso seja atualização da posição do veículo na fila
	// p.mu.Lock()
	// defer p.mu.Unlock()

	veiculoJaEstaNaFila := false

	for i := range p.Fila {
		if p.Fila[i].ID == v.ID {
			veiculoJaEstaNaFila = true
			p.Fila[i] = v
		}
	}

	if veiculoJaEstaNaFila {
		return true
	} else {
		p.Fila = append(p.Fila, v)
		p.QtdFila++
		fmt.Printf("\n\nPosto %s: Veículo %s adicionado à fila de espera. Posição: %d\n\n", p.ID, v.ID, p.QtdFila)
		return false
	}
}

func LiberarVaga(p *Posto) {
	// p.mu.Lock()
	// defer p.mu.Unlock()

	p.BombaOcupada = false
	fmt.Printf("Posto %s com a bomba liberada.\n", p.ID)

	if len(p.Fila) > 0 {
		for i := range p.Fila {
			if p.Fila[i].Latitude == p.Latitude && p.Fila[i].Longitude == p.Longitude {
				carregarBateriaVeiculo := p.Fila[i]

				p.Fila = append(p.Fila[:i], p.Fila[i+1:]...) // remove o veículo da fila; índice do primeiro elemento a ser removido e índice após o último elemento a ser removido

				CarregarBateria(carregarBateriaVeiculo, p)
				fmt.Printf("Posto %s: Veículo %s removido da fila e iniciando carregamento\n", p.ID, carregarBateriaVeiculo.ID)

				p.BombaOcupada = true
				break
			}
		}
	}
}

func GetPosFila(v Veiculo, p *Posto) int {
	var pos int = 0
	for i := range p.Fila {
		if p.Fila[i].ID == v.ID {
			pos = i
		}
	}
	return pos
}

func GetBombaDisponivel(p *Posto) bool {
	// p.mu.Lock()
	// defer p.mu.Unlock()

	return p.BombaOcupada
}

func GetLocalizacaoPosto(p *Posto) (float64, float64) {
	// p.mu.Lock()
	// defer p.mu.Unlock()

	return p.Latitude, p.Longitude
}

func PararCarregamentoBateria(v *Veiculo) {
	v.IsCarregando = false
	v.IsDeslocandoAoPosto = false
	v.Bateria = 100.0
	fmt.Printf("[%s] Carregamento concluído em: %s | Nível de bateria: %.2f%%\n", v.ID, time.Now().Format("02/01/2006 15:04:05"), v.Bateria)
}

func CarregarBateria(v *Veiculo, p *Posto) {
	v.IsCarregando = true
	v.IsDeslocandoAoPosto = false
	tempoInicio := time.Now()
	fmt.Printf("[%s] Carregamento iniciado em: %s | Nível de bateria inicial: %.2f%%\n", v.ID, tempoInicio.Format("02/01/2006 15:04:05"), v.Bateria)
	p.BombaOcupada = true
	// Goroutine para parar o carregamento após 1 minuto por cada 1% de bateria que falta
	go func() {
		time.Sleep(time.Duration(100-v.Bateria) * time.Second) // Espera 1 segundo por cada 1% de bateria que falta

		//v.IsCarregando = false
		PararCarregamentoBateria(v)
		p.BombaOcupada = false
	}()
}

func ArrumarPosicaoFila(p *Posto) {
	// p.mu.Lock()
	// defer p.mu.Unlock()

	//fmt.Println("POSTO", p.Fila)

	for i := range p.Fila {
		veiculo := p.Fila[i]
		if !veiculo.IsCarregando && veiculo.Latitude == p.Latitude && veiculo.Longitude == p.Longitude && !p.BombaOcupada {
			// Se o veículo já está no posto e a bomba não está ocupada, inicia o carregamento
			//carrega o primeiro veiculo da fila e o remove da fila para poder arrumar a fila
			CarregarBateria(veiculo, p)
			p.Fila = append(p.Fila[:i], p.Fila[i+1:]...) // remove o veículo da fila; índice do primeiro elemento a ser removido e índice após o último elemento a ser removido
			fmt.Printf("Posto %s: Veículo %s removido da fila e iniciando carregamento\n", p.ID, veiculo.ID)
			p.QtdFila--
			break
		}
	}

	//Ordena a fila baseado no tempo total estimado (ordem crescente)
	sort.Slice(p.Fila, func(i, j int) bool {
		tempoI, _ := CalcularTempoTotalVeiculo(p, p.Fila[i])
		tempoJ, _ := CalcularTempoTotalVeiculo(p, p.Fila[j])
		return tempoI < tempoJ // Menor tempo vem primeiro
	})
}

func CalcularTempoTotalVeiculo(p *Posto, v *Veiculo) (time.Duration, int) {
	if v.Latitude == p.Latitude && v.Longitude == p.Longitude {
		return TempoEstimado(p, 0)
	}
	tempoViagem := time.Duration(math.Abs(v.Latitude-p.Latitude)+math.Abs(v.Longitude-p.Longitude)) * 15 * time.Second
	return TempoEstimado(p, tempoViagem)
}

func TempoEstimado(p *Posto, tempoDistanciaVeiculo time.Duration) (time.Duration, int) {
	tempo_total := tempoDistanciaVeiculo
	posicao_na_fila := -1

	// se não houver veículos na fila, retorna apenas o tempo de chegada
	if len(p.Fila) == 0 {
		return tempo_total, 0
	}

	for i := range p.Fila {
		veiculo := p.Fila[i]
		var tempo_carregamento time.Duration

		if veiculo.Latitude == p.Latitude && veiculo.Longitude == p.Longitude {
			tempo_carregamento = time.Duration(100-veiculo.Bateria) * time.Minute
		} else {
			nivelBateriaAoChegarNoPosto := GetNivelBateriaAoChegarNoPosto(*veiculo, p)
			tempo_carregamento = time.Duration(100-nivelBateriaAoChegarNoPosto) * time.Minute
		}

		tempo_ate_posto_veiculo_fila := time.Duration(math.Abs(veiculo.Latitude-p.Latitude)+math.Abs(veiculo.Longitude-p.Longitude)) * 15 * time.Second

		tempo_total += tempo_carregamento

		// se o veículo atual chegar antes que este veículo da fila, é inserido a frente dele
		if tempoDistanciaVeiculo < tempo_ate_posto_veiculo_fila {
			posicao_na_fila = i
			break
		}
	}

	// Se não encontrou uma posição na fila, será o último
	if posicao_na_fila == -1 {
		posicao_na_fila = len(p.Fila)
	}

	return tempo_total, posicao_na_fila
}