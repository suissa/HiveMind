package tools

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2"
	"go.uber.org/ratelimit"
)

// BaseAPIDecorator é a interface base para todos os decorators
type BaseAPIDecorator interface {
	APITool
	GetWrapped() APITool
}

// CacheDecorator implementa cache de respostas
type CacheDecorator struct {
	wrapped APITool
	cache   *lru.Cache[string, *APIResponse]
}

// NewCacheDecorator cria um novo decorator de cache
func NewCacheDecorator(wrapped APITool, size int) (*CacheDecorator, error) {
	cache, err := lru.New[string, *APIResponse](size)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cache: %v", err)
	}

	return &CacheDecorator{
		wrapped: wrapped,
		cache:   cache,
	}, nil
}

func (d *CacheDecorator) GetWrapped() APITool {
	return d.wrapped
}

func (d *CacheDecorator) Request(options APIOptions) (*APIResponse, error) {
	// Só cacheia requisições GET
	if options.Method != "GET" {
		return d.wrapped.Request(options)
	}

	// Gerar chave do cache
	key := fmt.Sprintf("%s-%s-%v", options.Method, options.URL, options.QueryParams)

	// Verificar cache
	if cached, ok := d.cache.Get(key); ok {
		return cached, nil
	}

	// Fazer requisição
	resp, err := d.wrapped.Request(options)
	if err != nil {
		return nil, err
	}

	// Armazenar no cache se a requisição foi bem sucedida
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		d.cache.Add(key, resp)
	}

	return resp, nil
}

// CompressionDecorator implementa compressão de dados
type CompressionDecorator struct {
	wrapped APITool
}

func NewCompressionDecorator(wrapped APITool) *CompressionDecorator {
	return &CompressionDecorator{
		wrapped: wrapped,
	}
}

func (d *CompressionDecorator) GetWrapped() APITool {
	return d.wrapped
}

func (d *CompressionDecorator) Request(options APIOptions) (*APIResponse, error) {
	// Comprimir corpo da requisição se existir
	if options.Body != nil {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)

		bodyData, err := json.Marshal(options.Body)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar corpo: %v", err)
		}

		_, err = gz.Write(bodyData)
		if err != nil {
			return nil, fmt.Errorf("erro ao comprimir dados: %v", err)
		}
		gz.Close()

		options.Body = buf.Bytes()
		if options.Headers == nil {
			options.Headers = make(map[string]string)
		}
		options.Headers["Content-Encoding"] = "gzip"
	}

	return d.wrapped.Request(options)
}

// MetricsDecorator implementa métricas e logging
type MetricsDecorator struct {
	wrapped     APITool
	totalCalls  int64
	errorCalls  int64
	totalTime   time.Duration
	mutex       sync.RWMutex
}

func NewMetricsDecorator(wrapped APITool) *MetricsDecorator {
	return &MetricsDecorator{
		wrapped: wrapped,
	}
}

func (d *MetricsDecorator) GetWrapped() APITool {
	return d.wrapped
}

func (d *MetricsDecorator) Request(options APIOptions) (*APIResponse, error) {
	startTime := time.Now()

	resp, err := d.wrapped.Request(options)

	d.mutex.Lock()
	d.totalCalls++
	d.totalTime += time.Since(startTime)
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		d.errorCalls++
	}
	d.mutex.Unlock()

	return resp, err
}

func (d *MetricsDecorator) GetMetrics() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	var errorRate float64
	if d.totalCalls > 0 {
		errorRate = float64(d.errorCalls) / float64(d.totalCalls) * 100
	}

	var avgTime time.Duration
	if d.totalCalls > 0 {
		avgTime = d.totalTime / time.Duration(d.totalCalls)
	}

	return map[string]interface{}{
		"total_calls":     d.totalCalls,
		"error_calls":     d.errorCalls,
		"error_rate":      errorRate,
		"average_time_ms": avgTime.Milliseconds(),
		"total_time_ms":   d.totalTime.Milliseconds(),
	}
}

// CircuitBreakerDecorator implementa o padrão Circuit Breaker
type CircuitBreakerDecorator struct {
	wrapped            APITool
	failureThreshold   int
	resetTimeout       time.Duration
	failures          int
	lastFailure       time.Time
	state             string // closed, open, half-open
	mutex             sync.RWMutex
}

func NewCircuitBreakerDecorator(wrapped APITool, failureThreshold int, resetTimeout time.Duration) *CircuitBreakerDecorator {
	return &CircuitBreakerDecorator{
		wrapped:          wrapped,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:           "closed",
	}
}

func (d *CircuitBreakerDecorator) GetWrapped() APITool {
	return d.wrapped
}

func (d *CircuitBreakerDecorator) Request(options APIOptions) (*APIResponse, error) {
	d.mutex.Lock()
	state := d.state
	if state == "open" {
		if time.Since(d.lastFailure) > d.resetTimeout {
			d.state = "half-open"
			state = "half-open"
		}
	}
	d.mutex.Unlock()

	if state == "open" {
		return nil, fmt.Errorf("circuit breaker está aberto")
	}

	resp, err := d.wrapped.Request(options)

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		d.failures++
		d.lastFailure = time.Now()

		if d.failures >= d.failureThreshold {
			d.state = "open"
		}
	} else if state == "half-open" {
		d.state = "closed"
		d.failures = 0
	}

	return resp, err
}

// RateLimitDecorator implementa limitação de taxa de requisições
type RateLimitDecorator struct {
	wrapped APITool
	limiter ratelimit.Limiter
}

func NewRateLimitDecorator(wrapped APITool, rps int) *RateLimitDecorator {
	return &RateLimitDecorator{
		wrapped: wrapped,
		limiter: ratelimit.New(rps),
	}
}

func (d *RateLimitDecorator) GetWrapped() APITool {
	return d.wrapped
}

func (d *RateLimitDecorator) Request(options APIOptions) (*APIResponse, error) {
	d.limiter.Take()
	return d.wrapped.Request(options)
}

// TimeoutDecorator implementa timeout personalizado para requisições
type TimeoutDecorator struct {
	wrapped      APITool
	defaultTimeout time.Duration
}

func NewTimeoutDecorator(wrapped APITool, defaultTimeout time.Duration) *TimeoutDecorator {
	return &TimeoutDecorator{
		wrapped:        wrapped,
		defaultTimeout: defaultTimeout,
	}
}

func (d *TimeoutDecorator) GetWrapped() APITool {
	return d.wrapped
}

func (d *TimeoutDecorator) Request(options APIOptions) (*APIResponse, error) {
	// Usar timeout das opções se fornecido, senão usar o padrão
	timeout := d.defaultTimeout
	if options.Timeout > 0 {
		timeout = options.Timeout
	}

	// Criar contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Criar canal para resultado
	type result struct {
		response *APIResponse
		err      error
	}
	done := make(chan result, 1)

	// Executar requisição em goroutine
	go func() {
		resp, err := d.wrapped.Request(options)
		done <- result{resp, err}
	}()

	// Aguardar resultado ou timeout
	select {
	case <-ctx.Done():
		return &APIResponse{
			Error:        fmt.Sprintf("timeout após %v", timeout),
			ResponseTime: timeout,
		}, fmt.Errorf("timeout após %v", timeout)
	case r := <-done:
		return r.response, r.err
	}
} 