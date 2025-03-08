package tools

import "time"

// Transaction representa uma transação a ser analisada
type Transaction struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Timestamp       time.Time              `json:"timestamp"`
	Type            string                 `json:"type"`
	Status          string                 `json:"status"`
	Device          *DeviceInfo            `json:"device,omitempty"`
	Location        *LocationInfo          `json:"location,omitempty"`
	PaymentMethod   *PaymentMethodInfo     `json:"payment_method,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// DeviceInfo contém informações sobre o dispositivo usado na transação
type DeviceInfo struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	OS           string `json:"os"`
	Browser      string `json:"browser"`
	IP           string `json:"ip"`
	Fingerprint  string `json:"fingerprint"`
	IsTor        bool   `json:"is_tor"`
	IsVPN        bool   `json:"is_vpn"`
	IsProxy      bool   `json:"is_proxy"`
}

// LocationInfo contém informações de localização
type LocationInfo struct {
	Country     string  `json:"country"`
	City        string  `json:"city"`
	PostalCode  string  `json:"postal_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
}

// PaymentMethodInfo contém informações do método de pagamento
type PaymentMethodInfo struct {
	Type        string `json:"type"`
	Last4       string `json:"last4"`
	Brand       string `json:"brand"`
	Country     string `json:"country"`
	IssuingBank string `json:"issuing_bank"`
	IsExpired   bool   `json:"is_expired"`
}

// FraudScore representa a pontuação de risco de fraude
type FraudScore struct {
	Score       float64                `json:"score"`          // 0-100, onde 100 é risco máximo
	Risk        string                 `json:"risk"`           // "LOW", "MEDIUM", "HIGH"
	Triggers    []string              `json:"triggers"`       // Razões que contribuíram para o score
	Confidence  float64                `json:"confidence"`     // 0-1, confiança na análise
	Details     map[string]interface{} `json:"details"`        // Detalhes adicionais
	SuggestedAction string            `json:"suggested_action"` // Ação sugerida
}

// FraudAnalysisResult representa o resultado da análise de fraude
type FraudAnalysisResult struct {
	TransactionID string      `json:"transaction_id"`
	Score        FraudScore  `json:"score"`
	IsAccepted   bool        `json:"is_accepted"`
	ReviewNeeded bool        `json:"review_needed"`
	RulesTriggered []string  `json:"rules_triggered"`
	Timestamp    time.Time   `json:"timestamp"`
	ProcessingTime string    `json:"processing_time"`
	Error        string      `json:"error,omitempty"`
}

// FraudDetectionRule representa uma regra de detecção de fraude
type FraudDetectionRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Conditions  []RuleCondition        `json:"conditions"`
	Score       float64                `json:"score"`
	Action      string                 `json:"action"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RuleCondition representa uma condição para uma regra
type RuleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// FraudDetectionOptions representa as opções para detecção de fraude
type FraudDetectionOptions struct {
	EnableMLDetection bool     `json:"enable_ml_detection"`
	EnableRules      bool     `json:"enable_rules"`
	CustomRules      []FraudDetectionRule `json:"custom_rules,omitempty"`
	Threshold        float64  `json:"threshold"`
	Timeout         int      `json:"timeout"`
}

// FraudDetector é a interface que todas as ferramentas de detecção de fraude devem implementar
type FraudDetector interface {
	// AnalyzeTransaction analisa uma transação em busca de fraudes
	AnalyzeTransaction(transaction Transaction, options FraudDetectionOptions) (*FraudAnalysisResult, error)

	// BatchAnalyze analisa múltiplas transações em lote
	BatchAnalyze(transactions []Transaction, options FraudDetectionOptions) ([]*FraudAnalysisResult, error)

	// AddRule adiciona uma nova regra de detecção
	AddRule(rule FraudDetectionRule) error

	// UpdateRule atualiza uma regra existente
	UpdateRule(ruleID string, rule FraudDetectionRule) error

	// DeleteRule remove uma regra
	DeleteRule(ruleID string) error

	// GetRules retorna todas as regras configuradas
	GetRules() ([]FraudDetectionRule, error)

	// GetStatistics retorna estatísticas do detector
	GetStatistics() (map[string]interface{}, error)
} 