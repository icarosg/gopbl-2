package main

import (
	"testing"
)

// Função a ser testada
func Soma(a, b int) int {
	return a + b
}

// Teste de unidade para a função Soma
func TestSoma(t *testing.T) {
	// Definição dos casos de teste para a função Soma
	tests := []struct {
		name     string // Nome do caso de teste
		a, b     int    // Valores de entrada para a função Soma
		expected int    // Resultado esperado da soma
	}{
		{"Soma de números positivos", 2, 3, 5},    // Caso de teste com dois números positivos
		{"Soma com zero", 0, 5, 5},                // Caso de teste com um dos números sendo zero
		{"Soma de números negativos", -2, -3, -5}, // Caso de teste com dois números negativos
		{"Soma de positivo e negativo", 7, -3, 4}, // Caso de teste com um número positivo e um negativo
	}

	// Itera sobre cada caso de teste definido na slice 'tests'
	for _, tt := range tests {
		// Executa cada caso de teste como um subteste
		t.Run(tt.name, func(t *testing.T) {
			// Chama a função Soma com os valores de entrada do caso de teste
			result := Soma(tt.a, tt.b)
			// Verifica se o resultado obtido é igual ao esperado
			if result != tt.expected {
				// Reporta um erro se o resultado não for o esperado
				t.Errorf("Soma(%d, %d) = %d; esperado %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
