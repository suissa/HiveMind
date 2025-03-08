package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// BaseAgent fornece a implementação base para agentes treináveis
type BaseAgent struct {
	ID              string
	Name            string
	Description     string
	MaxRounds       int
	CurrentRound    int
	TrainingHistory []*TrainingMetrics
	State           map[string]interface{}
	mu              sync.RWMutex
}

// NewBaseAgent cria uma nova instância de BaseAgent
func NewBaseAgent(id, name, description string, maxRounds int) *BaseAgent {
	return &BaseAgent{
		ID:              id,
		Name:            name,
		Description:     description,
		MaxRounds:       maxRounds,
		TrainingHistory: make([]*TrainingMetrics, 0),
		State:           make(map[string]interface{}),
	}
}

// Train implementa o treinamento básico do agente
func (a *BaseAgent) Train(ctx context.Context, config TrainingConfig) (*TrainingMetrics, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	metrics := &TrainingMetrics{
		StartTime: time.Now(),
	}

	// Verifica se atingiu o número máximo de rounds
	if a.CurrentRound >= a.MaxRounds {
		return nil, fmt.Errorf("número máximo de rounds (%d) atingido", a.MaxRounds)
	}

	// Incrementa o contador de rounds
	a.CurrentRound++

	// Simula processo de treinamento
	select {
	case <-ctx.Done():
		metrics.EndTime = time.Now()
		metrics.Errors = append(metrics.Errors, ctx.Err())
		return metrics, ctx.Err()
	case <-time.After(config.TrainingTimeout):
		metrics.EndTime = time.Now()
		metrics.RoundsExecuted = a.CurrentRound

		// Adiciona ao histórico
		a.TrainingHistory = append(a.TrainingHistory, metrics)

		return metrics, nil
	}
}

// Validate verifica se o agente está pronto para execução
func (a *BaseAgent) Validate(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.CurrentRound == 0 {
		return fmt.Errorf("agente não foi treinado")
	}

	if len(a.TrainingHistory) == 0 {
		return fmt.Errorf("histórico de treinamento vazio")
	}

	lastMetrics := a.TrainingHistory[len(a.TrainingHistory)-1]
	if lastMetrics.Errors != nil && len(lastMetrics.Errors) > 0 {
		return fmt.Errorf("último treinamento contém erros: %v", lastMetrics.Errors)
	}

	return nil
}

// GetTrainingHistory retorna o histórico de treinamento
func (a *BaseAgent) GetTrainingHistory() []*TrainingMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Cria uma cópia do histórico para evitar condições de corrida
	history := make([]*TrainingMetrics, len(a.TrainingHistory))
	copy(history, a.TrainingHistory)

	return history
}

// SaveState salva o estado atual do agente
func (a *BaseAgent) SaveState(path string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Prepara os dados para salvar
	data := struct {
		ID              string
		Name            string
		Description     string
		MaxRounds       int
		CurrentRound    int
		TrainingHistory []*TrainingMetrics
		State           map[string]interface{}
	}{
		ID:              a.ID,
		Name:            a.Name,
		Description:     a.Description,
		MaxRounds:       a.MaxRounds,
		CurrentRound:    a.CurrentRound,
		TrainingHistory: a.TrainingHistory,
		State:           a.State,
	}

	// Serializa os dados
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar estado: %v", err)
	}

	// Salva no arquivo
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	return nil
}

// LoadState carrega um estado salvo
func (a *BaseAgent) LoadState(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Lê o arquivo
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	// Estrutura temporária para deserialização
	var data struct {
		ID              string
		Name            string
		Description     string
		MaxRounds       int
		CurrentRound    int
		TrainingHistory []*TrainingMetrics
		State           map[string]interface{}
	}

	// Deserializa os dados
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("erro ao deserializar estado: %v", err)
	}

	// Atualiza o estado do agente
	a.ID = data.ID
	a.Name = data.Name
	a.Description = data.Description
	a.MaxRounds = data.MaxRounds
	a.CurrentRound = data.CurrentRound
	a.TrainingHistory = data.TrainingHistory
	a.State = data.State

	return nil
}

// GetMaxRounds retorna o número máximo de rounds
func (a *BaseAgent) GetMaxRounds() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.MaxRounds
}

// GetCurrentRound retorna o round atual
func (a *BaseAgent) GetCurrentRound() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.CurrentRound
}

// ResetRounds reinicia a contagem de rounds
func (a *BaseAgent) ResetRounds() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.CurrentRound = 0
}
