package tools

import (
	"fmt"
	"time"
)

// Exemplo de como criar um cliente de API com todos os decorators
func ExampleAPIClient() {
	// Criar o cliente base
	baseClient := NewAPIClient()

	// Adicionar cache (100 itens)
	cachedClient, err := NewCacheDecorator(baseClient, 100)
	if err != nil {
		panic(err)
	}

	// Adicionar compressão
	compressedClient := NewCompressionDecorator(cachedClient)

	// Adicionar timeout global de 30 segundos
	timeoutClient := NewTimeoutDecorator(compressedClient, 30*time.Second)

	// Adicionar métricas
	metricsClient := NewMetricsDecorator(timeoutClient)

	// Adicionar circuit breaker (3 falhas, reset em 1 minuto)
	circuitClient := NewCircuitBreakerDecorator(metricsClient, 3, time.Minute)

	// Adicionar rate limit (10 requisições por segundo)
	finalClient := NewRateLimitDecorator(circuitClient, 10)

	// Exemplo de uso com timeout específico para esta requisição
	resp, err := finalClient.Request(APIOptions{
		Method: "GET",
		URL:    "https://api.exemplo.com/data",
		Auth: &APIAuth{
			Type:  "bearer",
			Token: "seu_token_aqui",
		},
		QueryParams: map[string]string{
			"page": "1",
			"size": "10",
		},
		Timeout:    5 * time.Second,  // Sobrescreve o timeout padrão de 30 segundos
		RetryCount: 3,
	})

	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Tempo de resposta: %v\n", resp.ResponseTime)

	// Obter métricas
	if metrics, ok := metricsClient.GetMetrics(); ok {
		fmt.Printf("Métricas:\n")
		for key, value := range metrics {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
}

// Exemplo de uso individual dos decorators
func ExampleIndividualDecorators() {
	baseClient := NewAPIClient()

	// Exemplo com apenas timeout
	timeoutClient := NewTimeoutDecorator(baseClient, 10*time.Second)
	timeoutClient.Request(APIOptions{
		Method: "GET",
		URL:    "https://api.exemplo.com/slow-endpoint",
	})

	// Exemplo com timeout e cache
	cachedClient, _ := NewCacheDecorator(baseClient, 100)
	timeoutCachedClient := NewTimeoutDecorator(cachedClient, 5*time.Second)
	timeoutCachedClient.Request(APIOptions{
		Method: "GET",
		URL:    "https://api.exemplo.com/cached-data",
		Timeout: 2*time.Second, // Timeout específico para esta requisição
	})

	// Exemplo com timeout e rate limit
	rateLimitedClient := NewRateLimitDecorator(baseClient, 5)
	timeoutRateLimitedClient := NewTimeoutDecorator(rateLimitedClient, 15*time.Second)
	timeoutRateLimitedClient.Request(APIOptions{
		Method: "POST",
		URL:    "https://api.exemplo.com/limited-endpoint",
		Body: map[string]interface{}{
			"data": "exemplo",
		},
	})
}