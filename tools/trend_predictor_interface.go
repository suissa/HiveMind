package tools

import "time"

// DataPoint representa um ponto de dados na série temporal
type DataPoint struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     float64     `json:"value"`
	Labels    []string    `json:"labels,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TimeSeries representa uma série temporal completa
type TimeSeries struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Unit        string      `json:"unit"`
	DataPoints  []DataPoint `json:"data_points"`
	Tags        []string    `json:"tags,omitempty"`
}

// TrendAnalysis representa a análise de tendência
type TrendAnalysis struct {
	Direction       string    `json:"direction"`        // "up", "down", "stable"
	Strength        float64   `json:"strength"`         // 0-1
	Confidence      float64   `json:"confidence"`       // 0-1
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	ChangeRate     float64   `json:"change_rate"`      // Taxa de mudança
	Volatility     float64   `json:"volatility"`       // Medida de volatilidade
	SeasonalityDetected bool `json:"seasonality_detected"`
}

// Prediction representa uma previsão futura
type Prediction struct {
	Timestamp      time.Time `json:"timestamp"`
	Value          float64   `json:"value"`
	LowerBound     float64   `json:"lower_bound"`      // Intervalo de confiança inferior
	UpperBound     float64   `json:"upper_bound"`      // Intervalo de confiança superior
	Confidence     float64   `json:"confidence"`       // 0-1
	Method         string    `json:"method"`           // Método usado para previsão
}

// TrendPredictionResult representa o resultado completo da predição
type TrendPredictionResult struct {
	SeriesID        string         `json:"series_id"`
	Analysis        TrendAnalysis  `json:"analysis"`
	Predictions     []Prediction   `json:"predictions"`
	Patterns        []Pattern      `json:"patterns,omitempty"`
	ProcessingTime  string         `json:"processing_time"`
	Error           string         `json:"error,omitempty"`
}

// Pattern representa um padrão identificado na série
type Pattern struct {
	Type        string    `json:"type"`            // "cycle", "season", "outlier"
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Confidence  float64   `json:"confidence"`
	Description string    `json:"description"`
}

// PredictionOptions representa as opções para previsão
type PredictionOptions struct {
	Method          string        `json:"method"`           // "arima", "prophet", "lstm", etc
	Horizon         time.Duration `json:"horizon"`          // Período futuro para prever
	Interval        time.Duration `json:"interval"`         // Intervalo entre previsões
	ConfidenceLevel float64       `json:"confidence_level"` // Nível de confiança (0-1)
	SeasonalPeriod  time.Duration `json:"seasonal_period"`  // Período sazonal (se aplicável)
	MinDataPoints   int           `json:"min_data_points"`  // Mínimo de pontos necessários
}

// TrendPredictor é a interface que todas as ferramentas de predição devem implementar
type TrendPredictor interface {
	// AddTimeSeries adiciona uma nova série temporal
	AddTimeSeries(series TimeSeries) error

	// UpdateTimeSeries atualiza uma série temporal existente
	UpdateTimeSeries(seriesID string, series TimeSeries) error

	// DeleteTimeSeries remove uma série temporal
	DeleteTimeSeries(seriesID string) error

	// GetTimeSeries retorna uma série temporal específica
	GetTimeSeries(seriesID string) (*TimeSeries, error)

	// ListTimeSeries lista todas as séries temporais disponíveis
	ListTimeSeries() ([]TimeSeries, error)

	// PredictTrend realiza a predição de tendência para uma série
	PredictTrend(seriesID string, options PredictionOptions) (*TrendPredictionResult, error)

	// AnalyzePatterns analisa padrões em uma série temporal
	AnalyzePatterns(seriesID string) ([]Pattern, error)

	// GetStatistics retorna estatísticas do preditor
	GetStatistics() (map[string]interface{}, error)
} 