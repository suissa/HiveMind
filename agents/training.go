package agents

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TrainingConfig contém as configurações para treinamento dos agentes
type TrainingConfig struct {
	MaxRounds       int           // Número máximo de rounds de execução
	TrainingTimeout time.Duration // Tempo máximo para treinamento
	ValidationRatio float64       // Proporção de dados usada para validação
	MinAccuracy     float64       // Precisão mínima requerida
	BatchSize       int           // Tamanho do lote para treinamento
	LearningRate    float64       // Taxa de aprendizado
	UseHistorical   bool          // Usar dados históricos para treinamento
	SaveCheckpoints bool          // Salvar checkpoints durante treinamento
}

// TrainingMetrics armazena métricas do treinamento
type TrainingMetrics struct {
	StartTime      time.Time
	EndTime        time.Time
	Accuracy       float64
	Loss           float64
	RoundsExecuted int
	Errors         []error
}

// TrainableAgent define a interface para agentes que podem ser treinados
type TrainableAgent interface {
	// Train executa o treinamento do agente
	Train(ctx context.Context, config TrainingConfig) (*TrainingMetrics, error)

	// Validate verifica se o agente está pronto para execução
	Validate(ctx context.Context) error

	// GetTrainingHistory retorna o histórico de treinamento
	GetTrainingHistory() []*TrainingMetrics

	// SaveState salva o estado atual do agente
	SaveState(path string) error

	// LoadState carrega um estado salvo
	LoadState(path string) error
}

// AgentTrainer gerencia o treinamento de múltiplos agentes
type AgentTrainer struct {
	agents  []TrainableAgent
	config  TrainingConfig
	metrics map[TrainableAgent]*TrainingMetrics
	mu      sync.RWMutex
}

// NewAgentTrainer cria uma nova instância do AgentTrainer
func NewAgentTrainer(config TrainingConfig) *AgentTrainer {
	return &AgentTrainer{
		agents:  make([]TrainableAgent, 0),
		config:  config,
		metrics: make(map[TrainableAgent]*TrainingMetrics),
	}
}

// AddAgent adiciona um agente para treinamento
func (t *AgentTrainer) AddAgent(agent TrainableAgent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.agents = append(t.agents, agent)
}

// Train executa o treinamento de todos os agentes
func (t *AgentTrainer) Train(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Cria um WaitGroup para sincronizar o treinamento
	var wg sync.WaitGroup
	errChan := make(chan error, len(t.agents))

	// Inicia o treinamento para cada agente
	for _, agent := range t.agents {
		wg.Add(1)
		go func(a TrainableAgent) {
			defer wg.Done()

			// Executa o treinamento
			metrics, err := a.Train(ctx, t.config)
			if err != nil {
				errChan <- fmt.Errorf("erro no treinamento do agente: %v", err)
				return
			}

			// Valida o agente após treinamento
			if err := a.Validate(ctx); err != nil {
				errChan <- fmt.Errorf("erro na validação do agente: %v", err)
				return
			}

			// Armazena as métricas
			t.metrics[a] = metrics
		}(agent)
	}

	// Aguarda conclusão de todos os treinamentos
	wg.Wait()
	close(errChan)

	// Verifica se houve erros
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("erros durante o treinamento: %v", errors)
	}

	return nil
}

// GetMetrics retorna as métricas de treinamento de um agente
func (t *AgentTrainer) GetMetrics(agent TrainableAgent) *TrainingMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.metrics[agent]
}

// GetAllMetrics retorna todas as métricas de treinamento
func (t *AgentTrainer) GetAllMetrics() map[TrainableAgent]*TrainingMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Cria uma cópia do mapa para evitar condições de corrida
	metrics := make(map[TrainableAgent]*TrainingMetrics)
	for agent, metric := range t.metrics {
		metrics[agent] = metric
	}

	return metrics
}
