package tools

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExampleFraudDetector demonstra o uso básico do detector de fraudes
func ExampleFraudDetector() {
	// Criar nova instância do detector
	detector, err := NewFraudDetector()
	if err != nil {
		fmt.Printf("Erro ao criar detector: %v\n", err)
		return
	}

	// Criar uma transação de exemplo
	transaction := Transaction{
		ID:        "tx123",
		UserID:    "user456",
		Amount:    15000.00,
		Currency:  "BRL",
		Timestamp: time.Now(),
		Type:      "payment",
		Status:    "pending",
		Device: DeviceInfo{
			ID:          "dev789",
			Type:        "mobile",
			OS:          "Android",
			Browser:     "Chrome",
			IP:          "200.158.100.123",
			Fingerprint: "abc123xyz",
			IsVPN:      true,
			IsProxy:    false,
		},
		Location: LocationInfo{
			Country:    "BR",
			City:      "São Paulo",
			PostalCode: "01310-200",
			Latitude:   -23.5505,
			Longitude: -46.6333,
			ISP:       "Vivo",
		},
		PaymentMethod: PaymentMethodInfo{
			Type:         "credit_card",
			LastFour:     "1234",
			Brand:        "Visa",
			Country:      "BR",
			IssuingBank:  "Banco XYZ",
			IsExpired:    false,
		},
		Metadata: map[string]interface{}{
			"customer_since": "2022-01-01",
			"risk_score":    80,
		},
	}

	// Configurar opções de detecção
	options := FraudDetectionOptions{
		EnableRules:      true,
		EnableMLDetection: true,
		Threshold:        75,
		CustomRules: []FraudDetectionRule{
			{
				ID:          "custom_high_risk",
				Name:        "Alto Risco Customizado",
				Description: "Detecta transações com score de risco alto no metadata",
				Type:        "metadata",
				Conditions: []RuleCondition{
					{Field: "metadata.risk_score", Operator: ">=", Value: 75},
				},
				Score:  85,
				Action: "block",
			},
		},
	}

	// Analisar transação
	result, err := detector.AnalyzeTransaction(transaction, options)
	if err != nil {
		fmt.Printf("Erro ao analisar transação: %v\n", err)
		return
	}

	// Imprimir resultado
	printAnalysisResult(result)

	// Exemplo de análise em lote
	transactions := []Transaction{transaction}
	// Adicionar mais uma transação ao lote
	transactions = append(transactions, Transaction{
		ID:        "tx124",
		UserID:    "user456",
		Amount:    500.00,
		Currency:  "BRL",
		Timestamp: time.Now(),
		Type:      "payment",
		Status:    "pending",
		Device:    transaction.Device,
		Location:  transaction.Location,
		PaymentMethod: PaymentMethodInfo{
			Type:         "debit_card",
			LastFour:     "5678",
			Brand:        "Mastercard",
			Country:      "BR",
			IssuingBank:  "Banco ABC",
			IsExpired:    false,
		},
	})

	// Analisar lote de transações
	results, err := detector.BatchAnalyze(transactions, options)
	if err != nil {
		fmt.Printf("Erro ao analisar lote: %v\n", err)
		return
	}

	// Imprimir resultados do lote
	fmt.Println("\nResultados do lote:")
	for i, result := range results {
		fmt.Printf("\nTransação %d:\n", i+1)
		printAnalysisResult(result)
	}

	// Imprimir estatísticas
	stats, err := detector.GetStatistics()
	if err != nil {
		fmt.Printf("Erro ao obter estatísticas: %v\n", err)
		return
	}

	fmt.Println("\nEstatísticas:")
	printJSON(stats)
}

// ExampleFraudDetectorAdvanced demonstra recursos avançados do detector
func ExampleFraudDetectorAdvanced() {
	detector, err := NewFraudDetector()
	if err != nil {
		fmt.Printf("Erro ao criar detector: %v\n", err)
		return
	}

	// Adicionar uma nova regra
	newRule := FraudDetectionRule{
		ID:          "velocity_check",
		Name:        "Verificação de Velocidade",
		Description: "Detecta múltiplas transações em curto período",
		Type:        "velocity",
		Conditions: []RuleCondition{
			{Field: "time_between_transactions", Operator: "<", Value: 300}, // 5 minutos
		},
		Score:  65,
		Action: "review",
	}

	if err := detector.AddRule(newRule); err != nil {
		fmt.Printf("Erro ao adicionar regra: %v\n", err)
		return
	}

	// Listar todas as regras
	rules, err := detector.GetRules()
	if err != nil {
		fmt.Printf("Erro ao listar regras: %v\n", err)
		return
	}

	fmt.Println("Regras configuradas:")
	printJSON(rules)

	// Atualizar uma regra
	newRule.Score = 75
	if err := detector.UpdateRule(newRule.ID, newRule); err != nil {
		fmt.Printf("Erro ao atualizar regra: %v\n", err)
		return
	}

	// Deletar uma regra
	if err := detector.DeleteRule("high_amount"); err != nil {
		fmt.Printf("Erro ao deletar regra: %v\n", err)
		return
	}

	// Verificar regras atualizadas
	rules, _ = detector.GetRules()
	fmt.Println("\nRegras após modificações:")
	printJSON(rules)
}

// Função auxiliar para imprimir resultado da análise
func printAnalysisResult(result *FraudAnalysisResult) {
	fmt.Printf("ID da Transação: %s\n", result.TransactionID)
	fmt.Printf("Score de Fraude: %.2f\n", result.Score.Score)
	fmt.Printf("Nível de Risco: %s\n", result.Score.Risk)
	fmt.Printf("Ação Sugerida: %s\n", result.Score.SuggestedAction)
	fmt.Printf("Confiança: %.2f\n", result.Score.Confidence)
	fmt.Printf("Regras Acionadas: %v\n", result.RulesTriggered)
	fmt.Printf("Tempo de Processamento: %s\n", result.ProcessingTime)
	if result.Error != "" {
		fmt.Printf("Erro: %s\n", result.Error)
	}
}

// Função auxiliar para imprimir JSON formatado
func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Erro ao formatar JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
} 