package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TrendPredictorImpl implementa a interface TrendPredictor
type TrendPredictorImpl struct {
	series     map[string]TimeSeries
	statistics map[string]interface{}
	dataPath   string
	mu         sync.RWMutex
}

// NewTrendPredictor cria uma nova instância do TrendPredictor
func NewTrendPredictor() (*TrendPredictorImpl, error) {
	predictor := &TrendPredictorImpl{
		series:     make(map[string]TimeSeries),
		statistics: make(map[string]interface{}),
		dataPath:   "trend_data",
	}

	// Criar diretório de dados se não existir
	if err := os.MkdirAll(predictor.dataPath, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de dados: %v", err)
	}

	// Carregar séries existentes
	if err := predictor.loadSeries(); err != nil {
		return nil, err
	}

	return predictor, nil
}

// AddTimeSeries adiciona uma nova série temporal
func (p *TrendPredictorImpl) AddTimeSeries(series TimeSeries) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.series[series.ID]; exists {
		return fmt.Errorf("série já existe: %s", series.ID)
	}

	p.series[series.ID] = series
	return p.saveSeries(series.ID)
}

// UpdateTimeSeries atualiza uma série temporal existente
func (p *TrendPredictorImpl) UpdateTimeSeries(seriesID string, series TimeSeries) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.series[seriesID]; !exists {
		return fmt.Errorf("série não encontrada: %s", seriesID)
	}

	p.series[seriesID] = series
	return p.saveSeries(seriesID)
}

// DeleteTimeSeries remove uma série temporal
func (p *TrendPredictorImpl) DeleteTimeSeries(seriesID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.series[seriesID]; !exists {
		return fmt.Errorf("série não encontrada: %s", seriesID)
	}

	delete(p.series, seriesID)
	return os.Remove(filepath.Join(p.dataPath, seriesID+".json"))
}

// GetTimeSeries retorna uma série temporal específica
func (p *TrendPredictorImpl) GetTimeSeries(seriesID string) (*TimeSeries, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	series, exists := p.series[seriesID]
	if !exists {
		return nil, fmt.Errorf("série não encontrada: %s", seriesID)
	}

	return &series, nil
}

// ListTimeSeries lista todas as séries temporais disponíveis
func (p *TrendPredictorImpl) ListTimeSeries() ([]TimeSeries, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	series := make([]TimeSeries, 0, len(p.series))
	for _, s := range p.series {
		series = append(series, s)
	}
	return series, nil
}

// PredictTrend realiza a predição de tendência para uma série
func (p *TrendPredictorImpl) PredictTrend(seriesID string, options PredictionOptions) (*TrendPredictionResult, error) {
	startTime := time.Now()

	// Obter série temporal
	series, err := p.GetTimeSeries(seriesID)
	if err != nil {
		return nil, err
	}

	// Verificar número mínimo de pontos
	if len(series.DataPoints) < options.MinDataPoints {
		return nil, fmt.Errorf("série precisa ter pelo menos %d pontos", options.MinDataPoints)
	}

	// Realizar análise de tendência
	analysis := p.analyzeTrend(series)

	// Gerar previsões
	predictions, err := p.generatePredictions(series, options)
	if err != nil {
		return nil, err
	}

	// Analisar padrões
	patterns, _ := p.AnalyzePatterns(seriesID)

	result := &TrendPredictionResult{
		SeriesID:       seriesID,
		Analysis:       analysis,
		Predictions:    predictions,
		Patterns:       patterns,
		ProcessingTime: time.Since(startTime).String(),
	}

	// Atualizar estatísticas
	p.updateStatistics(result)

	return result, nil
}

// AnalyzePatterns analisa padrões em uma série temporal
func (p *TrendPredictorImpl) AnalyzePatterns(seriesID string) ([]Pattern, error) {
	series, err := p.GetTimeSeries(seriesID)
	if err != nil {
		return nil, err
	}

	patterns := make([]Pattern, 0)

	// Detectar sazonalidade
	if seasonal := p.detectSeasonality(series); seasonal != nil {
		patterns = append(patterns, *seasonal)
	}

	// Detectar ciclos
	if cycle := p.detectCycles(series); cycle != nil {
		patterns = append(patterns, *cycle)
	}

	// Detectar outliers
	outliers := p.detectOutliers(series)
	patterns = append(patterns, outliers...)

	return patterns, nil
}

// GetStatistics retorna estatísticas do preditor
func (p *TrendPredictorImpl) GetStatistics() (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range p.statistics {
		stats[k] = v
	}
	return stats, nil
}

// Funções auxiliares

func (p *TrendPredictorImpl) loadSeries() error {
	files, err := os.ReadDir(p.dataPath)
	if err != nil {
		return fmt.Errorf("erro ao ler diretório de dados: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(p.dataPath, file.Name()))
		if err != nil {
			return fmt.Errorf("erro ao ler arquivo %s: %v", file.Name(), err)
		}

		var series TimeSeries
		if err := json.Unmarshal(data, &series); err != nil {
			return fmt.Errorf("erro ao decodificar série %s: %v", file.Name(), err)
		}

		p.series[series.ID] = series
	}

	return nil
}

func (p *TrendPredictorImpl) saveSeries(seriesID string) error {
	series := p.series[seriesID]
	data, err := json.MarshalIndent(series, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao codificar série: %v", err)
	}

	filename := filepath.Join(p.dataPath, seriesID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar série: %v", err)
	}

	return nil
}

func (p *TrendPredictorImpl) analyzeTrend(series *TimeSeries) TrendAnalysis {
	analysis := TrendAnalysis{
		StartTime: series.DataPoints[0].Timestamp,
		EndTime:   series.DataPoints[len(series.DataPoints)-1].Timestamp,
	}

	// Calcular direção e força da tendência
	firstValue := series.DataPoints[0].Value
	lastValue := series.DataPoints[len(series.DataPoints)-1].Value
	changePercent := (lastValue - firstValue) / firstValue * 100

	if math.Abs(changePercent) < 5 {
		analysis.Direction = "stable"
		analysis.Strength = 0.2
	} else if changePercent > 0 {
		analysis.Direction = "up"
		analysis.Strength = math.Min(math.Abs(changePercent)/100, 1)
	} else {
		analysis.Direction = "down"
		analysis.Strength = math.Min(math.Abs(changePercent)/100, 1)
	}

	// Calcular taxa de mudança
	timeDiff := analysis.EndTime.Sub(analysis.StartTime)
	analysis.ChangeRate = changePercent / timeDiff.Hours() * 24 // Taxa diária

	// Calcular volatilidade
	analysis.Volatility = p.calculateVolatility(series)

	// Detectar sazonalidade
	analysis.SeasonalityDetected = p.detectSeasonality(series) != nil

	// Definir confiança baseada na volatilidade e quantidade de dados
	analysis.Confidence = 1 - math.Min(analysis.Volatility, 0.8)
	if len(series.DataPoints) < 30 {
		analysis.Confidence *= 0.7
	}

	return analysis
}

func (p *TrendPredictorImpl) generatePredictions(series *TimeSeries, options PredictionOptions) ([]Prediction, error) {
	predictions := make([]Prediction, 0)
	lastPoint := series.DataPoints[len(series.DataPoints)-1]
	currentTime := lastPoint.Timestamp

	for currentTime.Before(lastPoint.Timestamp.Add(options.Horizon)) {
		currentTime = currentTime.Add(options.Interval)
		
		// Calcular previsão baseada no método escolhido
		var prediction float64
		var confidence float64
		var bounds float64

		switch options.Method {
		case "arima":
			prediction, confidence, bounds = p.arimaPredict(series, currentTime)
		case "prophet":
			prediction, confidence, bounds = p.prophetPredict(series, currentTime)
		case "lstm":
			prediction, confidence, bounds = p.lstmPredict(series, currentTime)
		default:
			prediction, confidence, bounds = p.simplePredict(series, currentTime)
		}

		predictions = append(predictions, Prediction{
			Timestamp:  currentTime,
			Value:     prediction,
			LowerBound: prediction - bounds,
			UpperBound: prediction + bounds,
			Confidence: confidence,
			Method:    options.Method,
		})
	}

	return predictions, nil
}

func (p *TrendPredictorImpl) calculateVolatility(series *TimeSeries) float64 {
	if len(series.DataPoints) < 2 {
		return 0
	}

	var sumSquaredChanges float64
	for i := 1; i < len(series.DataPoints); i++ {
		change := series.DataPoints[i].Value - series.DataPoints[i-1].Value
		sumSquaredChanges += change * change
	}

	return math.Sqrt(sumSquaredChanges / float64(len(series.DataPoints)-1))
}

func (p *TrendPredictorImpl) detectSeasonality(series *TimeSeries) *Pattern {
	// Implementação simplificada de detecção de sazonalidade
	// Aqui você pode implementar métodos mais avançados como análise de Fourier
	return nil
}

func (p *TrendPredictorImpl) detectCycles(series *TimeSeries) *Pattern {
	// Implementação simplificada de detecção de ciclos
	return nil
}

func (p *TrendPredictorImpl) detectOutliers(series *TimeSeries) []Pattern {
	outliers := make([]Pattern, 0)
	
	if len(series.DataPoints) < 3 {
		return outliers
	}

	// Calcular média e desvio padrão
	var sum, sumSquared float64
	for _, point := range series.DataPoints {
		sum += point.Value
		sumSquared += point.Value * point.Value
	}
	mean := sum / float64(len(series.DataPoints))
	stdDev := math.Sqrt(sumSquared/float64(len(series.DataPoints)) - mean*mean)

	// Detectar outliers (valores além de 3 desvios padrão)
	threshold := 3 * stdDev
	for _, point := range series.DataPoints {
		if math.Abs(point.Value-mean) > threshold {
			outliers = append(outliers, Pattern{
				Type:        "outlier",
				StartTime:   point.Timestamp,
				EndTime:     point.Timestamp,
				Confidence:  0.9,
				Description: fmt.Sprintf("Valor atípico detectado: %.2f (média: %.2f, desvio: %.2f)", point.Value, mean, stdDev),
			})
		}
	}

	return outliers
}

// Métodos de previsão

func (p *TrendPredictorImpl) simplePredict(series *TimeSeries, targetTime time.Time) (float64, float64, float64) {
	// Implementação simplificada usando média móvel
	windowSize := 5
	if len(series.DataPoints) < windowSize {
		windowSize = len(series.DataPoints)
	}

	var sum float64
	for i := len(series.DataPoints) - windowSize; i < len(series.DataPoints); i++ {
		sum += series.DataPoints[i].Value
	}
	prediction := sum / float64(windowSize)
	
	confidence := 0.6
	bounds := p.calculateVolatility(series) * 2

	return prediction, confidence, bounds
}

func (p *TrendPredictorImpl) arimaPredict(series *TimeSeries, targetTime time.Time) (float64, float64, float64) {
	// Implementar ARIMA
	return p.simplePredict(series, targetTime)
}

func (p *TrendPredictorImpl) prophetPredict(series *TimeSeries, targetTime time.Time) (float64, float64, float64) {
	// Implementar Prophet
	return p.simplePredict(series, targetTime)
}

func (p *TrendPredictorImpl) lstmPredict(series *TimeSeries, targetTime time.Time) (float64, float64, float64) {
	// Implementar LSTM
	return p.simplePredict(series, targetTime)
}

func (p *TrendPredictorImpl) updateStatistics(result *TrendPredictionResult) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.statistics["total_predictions"] = p.statistics["total_predictions"].(int) + 1
	p.statistics["last_prediction"] = time.Now()
	
	// Atualizar estatísticas de precisão se tivermos dados reais para comparar
	if actual, ok := p.statistics["actual_values"].(map[string]float64); ok {
		if actualValue, exists := actual[result.SeriesID]; exists {
			error := math.Abs(result.Predictions[0].Value - actualValue)
			p.statistics["average_error"] = (p.statistics["average_error"].(float64)*float64(p.statistics["total_predictions"].(int)-1) + error) / float64(p.statistics["total_predictions"].(int))
		}
	}
} 