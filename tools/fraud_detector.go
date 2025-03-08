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

// FraudDetectorImpl implementa a interface FraudDetector
type FraudDetectorImpl struct {
	rules      []FraudDetectionRule
	statistics map[string]interface{}
	rulesPath  string
	mu         sync.RWMutex
}

// NewFraudDetector cria uma nova instância do FraudDetector
func NewFraudDetector() (*FraudDetectorImpl, error) {
	detector := &FraudDetectorImpl{
		rules:      make([]FraudDetectionRule, 0),
		statistics: make(map[string]interface{}),
		rulesPath:  "fraud_rules.json",
	}

	// Carregar regras padrão
	if err := detector.loadDefaultRules(); err != nil {
		return nil, err
	}

	// Carregar regras personalizadas se existirem
	if err := detector.loadRules(); err != nil {
		return nil, err
	}

	return detector, nil
}

// AnalyzeTransaction analisa uma transação em busca de fraudes
func (d *FraudDetectorImpl) AnalyzeTransaction(transaction Transaction, options FraudDetectionOptions) (*FraudAnalysisResult, error) {
	startTime := time.Now()

	result := &FraudAnalysisResult{
		TransactionID: transaction.ID,
		Timestamp:    time.Now(),
		RulesTriggered: make([]string, 0),
	}

	// Aplicar regras de detecção
	score := d.applyRules(transaction, options, result)

	// Aplicar detecção por ML se habilitado
	if options.EnableMLDetection {
		mlScore := d.applyMLDetection(transaction)
		score = (score + mlScore) / 2
	}

	// Calcular score final e risco
	result.Score = d.calculateFinalScore(score, transaction)

	// Determinar se a transação deve ser aceita
	result.IsAccepted = result.Score.Score < options.Threshold
	result.ReviewNeeded = result.Score.Score >= options.Threshold*0.7

	result.ProcessingTime = time.Since(startTime).String()

	// Atualizar estatísticas
	d.updateStatistics(result)

	return result, nil
}

// BatchAnalyze analisa múltiplas transações em lote
func (d *FraudDetectorImpl) BatchAnalyze(transactions []Transaction, options FraudDetectionOptions) ([]*FraudAnalysisResult, error) {
	results := make([]*FraudAnalysisResult, len(transactions))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, transaction := range transactions {
		wg.Add(1)
		go func(idx int, tx Transaction) {
			defer wg.Done()

			result, err := d.AnalyzeTransaction(tx, options)
			if err != nil {
				result = &FraudAnalysisResult{
					TransactionID: tx.ID,
					Error:        err.Error(),
				}
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, transaction)
	}

	wg.Wait()
	return results, nil
}

// AddRule adiciona uma nova regra de detecção
func (d *FraudDetectorImpl) AddRule(rule FraudDetectionRule) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Validar regra
	if err := d.validateRule(rule); err != nil {
		return err
	}

	d.rules = append(d.rules, rule)
	return d.saveRules()
}

// UpdateRule atualiza uma regra existente
func (d *FraudDetectorImpl) UpdateRule(ruleID string, rule FraudDetectionRule) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, r := range d.rules {
		if r.ID == ruleID {
			if err := d.validateRule(rule); err != nil {
				return err
			}
			d.rules[i] = rule
			return d.saveRules()
		}
	}

	return fmt.Errorf("regra não encontrada: %s", ruleID)
}

// DeleteRule remove uma regra
func (d *FraudDetectorImpl) DeleteRule(ruleID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, rule := range d.rules {
		if rule.ID == ruleID {
			d.rules = append(d.rules[:i], d.rules[i+1:]...)
			return d.saveRules()
		}
	}

	return fmt.Errorf("regra não encontrada: %s", ruleID)
}

// GetRules retorna todas as regras configuradas
func (d *FraudDetectorImpl) GetRules() ([]FraudDetectionRule, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rules := make([]FraudDetectionRule, len(d.rules))
	copy(rules, d.rules)
	return rules, nil
}

// GetStatistics retorna estatísticas do detector
func (d *FraudDetectorImpl) GetStatistics() (map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range d.statistics {
		stats[k] = v
	}
	return stats, nil
}

// Funções auxiliares

func (d *FraudDetectorImpl) loadDefaultRules() error {
	defaultRules := []FraudDetectionRule{
		{
			ID:          "high_amount",
			Name:        "Alto Valor",
			Description: "Detecta transações com valor muito alto",
			Type:        "amount",
			Conditions: []RuleCondition{
				{Field: "amount", Operator: ">=", Value: 10000},
			},
			Score:  70,
			Action: "review",
		},
		{
			ID:          "multiple_countries",
			Name:        "Múltiplos Países",
			Description: "Detecta transações do mesmo usuário em países diferentes",
			Type:        "location",
			Conditions: []RuleCondition{
				{Field: "location.country", Operator: "!=", Value: "user.last_country"},
			},
			Score:  60,
			Action: "review",
		},
		{
			ID:          "vpn_proxy",
			Name:        "VPN/Proxy",
			Description: "Detecta uso de VPN ou proxy",
			Type:        "device",
			Conditions: []RuleCondition{
				{Field: "device.is_vpn", Operator: "==", Value: true},
				{Field: "device.is_proxy", Operator: "==", Value: true},
			},
			Score:  50,
			Action: "flag",
		},
	}

	d.rules = append(d.rules, defaultRules...)
	return nil
}

func (d *FraudDetectorImpl) loadRules() error {
	data, err := os.ReadFile(d.rulesPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("erro ao ler regras: %v", err)
	}

	var rules []FraudDetectionRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("erro ao decodificar regras: %v", err)
	}

	d.rules = append(d.rules, rules...)
	return nil
}

func (d *FraudDetectorImpl) saveRules() error {
	data, err := json.MarshalIndent(d.rules, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao codificar regras: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(d.rulesPath), 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %v", err)
	}

	if err := os.WriteFile(d.rulesPath, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar regras: %v", err)
	}

	return nil
}

func (d *FraudDetectorImpl) validateRule(rule FraudDetectionRule) error {
	if rule.ID == "" {
		return fmt.Errorf("ID da regra é obrigatório")
	}
	if rule.Score < 0 || rule.Score > 100 {
		return fmt.Errorf("score deve estar entre 0 e 100")
	}
	if len(rule.Conditions) == 0 {
		return fmt.Errorf("regra deve ter pelo menos uma condição")
	}
	return nil
}

func (d *FraudDetectorImpl) applyRules(tx Transaction, options FraudDetectionOptions, result *FraudAnalysisResult) float64 {
	if !options.EnableRules {
		return 0
	}

	var totalScore float64
	var appliedRules int

	for _, rule := range d.rules {
		if d.evaluateRule(tx, rule) {
			totalScore += rule.Score
			result.RulesTriggered = append(result.RulesTriggered, rule.ID)
			appliedRules++
		}
	}

	if appliedRules > 0 {
		return totalScore / float64(appliedRules)
	}
	return 0
}

func (d *FraudDetectorImpl) evaluateRule(tx Transaction, rule FraudDetectionRule) bool {
	for _, condition := range rule.Conditions {
		if !d.evaluateCondition(tx, condition) {
			return false
		}
	}
	return true
}

func (d *FraudDetectorImpl) evaluateCondition(tx Transaction, condition RuleCondition) bool {
	value := d.getFieldValue(tx, condition.Field)
	switch condition.Operator {
	case "==":
		return value == condition.Value
	case "!=":
		return value != condition.Value
	case ">=":
		return d.compareNumeric(value, condition.Value) >= 0
	case "<=":
		return d.compareNumeric(value, condition.Value) <= 0
	case ">":
		return d.compareNumeric(value, condition.Value) > 0
	case "<":
		return d.compareNumeric(value, condition.Value) < 0
	}
	return false
}

func (d *FraudDetectorImpl) getFieldValue(tx Transaction, field string) interface{} {
	// Implementar acesso a campos aninhados
	return nil
}

func (d *FraudDetectorImpl) compareNumeric(a, b interface{}) float64 {
	var valA, valB float64

	switch v := a.(type) {
	case int:
		valA = float64(v)
	case float64:
		valA = v
	}

	switch v := b.(type) {
	case int:
		valB = float64(v)
	case float64:
		valB = v
	}

	return valA - valB
}

func (d *FraudDetectorImpl) applyMLDetection(tx Transaction) float64 {
	// Implementar detecção baseada em ML
	// Esta é uma implementação simplificada
	score := 0.0

	// Verificar padrões de comportamento
	if tx.Amount > getAverageAmount(tx.UserID) {
		score += 20
	}

	// Verificar localização suspeita
	if isUnusualLocation(tx) {
		score += 30
	}

	// Verificar velocidade de transações
	if isHighFrequency(tx) {
		score += 25
	}

	return score
}

func (d *FraudDetectorImpl) calculateFinalScore(score float64, tx Transaction) FraudScore {
	fraudScore := FraudScore{
		Score:      score,
		Confidence: 0.8,
		Details:    make(map[string]interface{}),
		Triggers:   make([]string, 0),
	}

	// Determinar nível de risco
	if score >= 80 {
		fraudScore.Risk = "HIGH"
		fraudScore.SuggestedAction = "block"
	} else if score >= 50 {
		fraudScore.Risk = "MEDIUM"
		fraudScore.SuggestedAction = "review"
	} else {
		fraudScore.Risk = "LOW"
		fraudScore.SuggestedAction = "accept"
	}

	// Adicionar detalhes
	fraudScore.Details["transaction_amount"] = tx.Amount
	fraudScore.Details["user_history"] = getUserHistory(tx.UserID)

	return fraudScore
}

func (d *FraudDetectorImpl) updateStatistics(result *FraudAnalysisResult) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.statistics["total_transactions"] = d.statistics["total_transactions"].(int) + 1
	if !result.IsAccepted {
		d.statistics["blocked_transactions"] = d.statistics["blocked_transactions"].(int) + 1
	}
	if result.ReviewNeeded {
		d.statistics["reviewed_transactions"] = d.statistics["reviewed_transactions"].(int) + 1
	}
}

// Funções auxiliares para ML (implementações simplificadas)

func getAverageAmount(userID string) float64 {
	return 1000 // Implementar cálculo real
}

func isUnusualLocation(tx Transaction) bool {
	return false // Implementar verificação real
}

func isHighFrequency(tx Transaction) bool {
	return false // Implementar verificação real
}

func getUserHistory(userID string) map[string]interface{} {
	return map[string]interface{}{
		"total_transactions": 10,
		"average_amount":    500,
		"last_transaction":  time.Now().Add(-24 * time.Hour),
	}
} 