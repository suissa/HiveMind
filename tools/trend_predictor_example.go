package tools

import (
	"fmt"
	"math"
	"time"
)

// ExampleTrendPredictor demonstra o uso básico do preditor de tendências
func ExampleTrendPredictor() {
	// Criar nova instância do preditor
	predictor, err := NewTrendPredictor()
	if err != nil {
		fmt.Printf("Erro ao criar preditor: %v\n", err)
		return
	}

	// Criar uma série temporal de exemplo (preços de ações)
	series := TimeSeries{
		ID:          "PETR4",
		Name:        "Petrobras PN",
		Description: "Preços diários das ações da Petrobras",
		Unit:        "BRL",
		Tags:        []string{"ações", "petróleo", "energia"},
	}

	// Gerar dados sintéticos para exemplo
	startDate := time.Now().AddDate(0, -1, 0) // Último mês
	for i := 0; i < 30; i++ {
		// Simular tendência de alta com alguma volatilidade
		baseValue := 30.0 + float64(i)*0.5
		noise := math.Sin(float64(i)*0.5) * 2
		
		series.DataPoints = append(series.DataPoints, DataPoint{
			Timestamp: startDate.AddDate(0, 0, i),
			Value:     baseValue + noise,
			Labels:    []string{"daily"},
		})
	}

	// Adicionar série temporal
	if err := predictor.AddTimeSeries(series); err != nil {
		fmt.Printf("Erro ao adicionar série: %v\n", err)
		return
	}

	// Configurar opções de predição
	options := PredictionOptions{
		Method:          "arima",
		Horizon:         7 * 24 * time.Hour, // Prever próxima semana
		Interval:        24 * time.Hour,     // Previsões diárias
		ConfidenceLevel: 0.95,
		MinDataPoints:   10,
	}

	// Realizar predição
	result, err := predictor.PredictTrend(series.ID, options)
	if err != nil {
		fmt.Printf("Erro ao prever tendência: %v\n", err)
		return
	}

	// Imprimir resultados
	fmt.Println("\nAnálise de Tendência:")
	fmt.Printf("Direção: %s\n", result.Analysis.Direction)
	fmt.Printf("Força: %.2f\n", result.Analysis.Strength)
	fmt.Printf("Confiança: %.2f\n", result.Analysis.Confidence)
	fmt.Printf("Taxa de Mudança: %.2f%% ao dia\n", result.Analysis.ChangeRate)
	fmt.Printf("Volatilidade: %.2f\n", result.Analysis.Volatility)
	fmt.Printf("Sazonalidade Detectada: %v\n", result.Analysis.SeasonalityDetected)

	fmt.Println("\nPrevisões:")
	for _, pred := range result.Predictions {
		fmt.Printf("Data: %s\n", pred.Timestamp.Format("2006-01-02"))
		fmt.Printf("Valor Previsto: %.2f (%.2f - %.2f)\n", pred.Value, pred.LowerBound, pred.UpperBound)
		fmt.Printf("Confiança: %.2f\n", pred.Confidence)
		fmt.Println()
	}

	if len(result.Patterns) > 0 {
		fmt.Println("\nPadrões Detectados:")
		for _, pattern := range result.Patterns {
			fmt.Printf("Tipo: %s\n", pattern.Type)
			fmt.Printf("Período: %s - %s\n", pattern.StartTime.Format("2006-01-02"), pattern.EndTime.Format("2006-01-02"))
			fmt.Printf("Descrição: %s\n", pattern.Description)
			fmt.Printf("Confiança: %.2f\n", pattern.Confidence)
			fmt.Println()
		}
	}
}

// ExampleTrendPredictorAdvanced demonstra recursos avançados do preditor
func ExampleTrendPredictorAdvanced() {
	predictor, err := NewTrendPredictor()
	if err != nil {
		fmt.Printf("Erro ao criar preditor: %v\n", err)
		return
	}

	// Criar múltiplas séries temporais
	series := []TimeSeries{
		{
			ID:          "IBOV",
			Name:        "Índice Bovespa",
			Description: "Pontos do Índice Bovespa",
			Unit:        "pontos",
			Tags:        []string{"índice", "mercado"},
		},
		{
			ID:          "DOLAR",
			Name:        "Dólar Comercial",
			Description: "Cotação do dólar comercial",
			Unit:        "BRL",
			Tags:        []string{"câmbio", "moeda"},
		},
	}

	// Gerar dados sintéticos para cada série
	startDate := time.Now().AddDate(0, -2, 0) // Últimos 2 meses
	for _, s := range series {
		for i := 0; i < 60; i++ {
			var baseValue float64
			if s.ID == "IBOV" {
				baseValue = 100000.0 + float64(i)*100
			} else {
				baseValue = 5.0 + float64(i)*0.02
			}
			
			noise := math.Sin(float64(i)*0.3) * baseValue * 0.01
			
			s.DataPoints = append(s.DataPoints, DataPoint{
				Timestamp: startDate.AddDate(0, 0, i),
				Value:     baseValue + noise,
				Labels:    []string{"daily"},
			})
		}

		if err := predictor.AddTimeSeries(s); err != nil {
			fmt.Printf("Erro ao adicionar série %s: %v\n", s.ID, err)
			continue
		}
	}

	// Testar diferentes métodos de previsão
	methods := []string{"arima", "prophet", "lstm"}
	horizons := []time.Duration{
		7 * 24 * time.Hour,  // 1 semana
		30 * 24 * time.Hour, // 1 mês
	}

	for _, series := range series {
		fmt.Printf("\nAnálise para %s (%s)\n", series.Name, series.ID)
		fmt.Println("----------------------------------------")

		for _, method := range methods {
			for _, horizon := range horizons {
				options := PredictionOptions{
					Method:          method,
					Horizon:         horizon,
					Interval:        24 * time.Hour,
					ConfidenceLevel: 0.95,
					MinDataPoints:   30,
				}

				result, err := predictor.PredictTrend(series.ID, options)
				if err != nil {
					fmt.Printf("Erro ao prever tendência com %s para %v: %v\n", method, horizon, err)
					continue
				}

				fmt.Printf("\nMétodo: %s, Horizonte: %v\n", method, horizon)
				fmt.Printf("Direção: %s (Força: %.2f, Confiança: %.2f)\n",
					result.Analysis.Direction, result.Analysis.Strength, result.Analysis.Confidence)
				fmt.Printf("Primeira Previsão: %.2f (%.2f - %.2f)\n",
					result.Predictions[0].Value,
					result.Predictions[0].LowerBound,
					result.Predictions[0].UpperBound)
				fmt.Printf("Última Previsão: %.2f (%.2f - %.2f)\n",
					result.Predictions[len(result.Predictions)-1].Value,
					result.Predictions[len(result.Predictions)-1].LowerBound,
					result.Predictions[len(result.Predictions)-1].UpperBound)
			}
		}
	}

	// Imprimir estatísticas gerais
	stats, err := predictor.GetStatistics()
	if err != nil {
		fmt.Printf("Erro ao obter estatísticas: %v\n", err)
		return
	}

	fmt.Println("\nEstatísticas Gerais:")
	fmt.Printf("Total de Predições: %d\n", stats["total_predictions"])
	fmt.Printf("Última Predição: %v\n", stats["last_prediction"])
	if avgError, ok := stats["average_error"]; ok {
		fmt.Printf("Erro Médio: %.2f\n", avgError)
	}
} 