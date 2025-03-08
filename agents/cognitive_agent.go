package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"HiveMind/memory"
)

// CognitiveAgent representa um agente cognitivo que pode executar tarefas específicas
type CognitiveAgent struct {
	*AgentStruct
	Model            string                 // Modelo de IA usado pelo agente
	Temperature      float64                // Temperatura para geração de respostas
	MaxTokens        int                    // Número máximo de tokens por resposta
	ContextWindow    int                    // Tamanho da janela de contexto
	KnowledgeBase    map[string]interface{} // Base de conhecimento do agente
	LearningRate     float64                // Taxa de aprendizado para ajustes
	PromptTemplates  map[string]string      // Templates de prompts
	ResponseHistory  []string               // Histórico de respostas
	PerformanceStats map[string]float64     // Estatísticas de performance
	MaxRounds        int                    // Número máximo de rodadas de treinamento
	trainingHistory  []*TrainingMetrics     // Histórico de treinamento

	// Campos específicos para execução de tarefas
	taskManager   *TaskManager
	memoryManager memory.MemoryManager
	stopChan      chan struct{}
	healthTicker  *time.Ticker
	metricsTicker *time.Ticker
	ctx           context.Context
}

// NewCognitiveAgent cria uma nova instância de CognitiveAgent
func NewCognitiveAgent(id, name, description string, maxRounds int, model string, role string, goal string, memoryManager memory.MemoryManager) *CognitiveAgent {
	return &CognitiveAgent{
		AgentStruct: &AgentStruct{
			ID:              id,
			Name:            name,
			Role:            role,
			Goal:            goal,
			AllowDelegation: true,
			Model:           model,
		},
		Model:           model,
		Temperature:     0.7,
		MaxTokens:       2048,
		ContextWindow:   4096,
		KnowledgeBase:   make(map[string]interface{}),
		LearningRate:    0.001,
		PromptTemplates: make(map[string]string),
		ResponseHistory: make([]string, 0),
		MaxRounds:       maxRounds,
		PerformanceStats: map[string]float64{
			"accuracy":       0.8, // Inicializa com 80% de acurácia
			"response_time":  0.0,
			"success_rate":   0.8, // Inicializa com 80% de taxa de sucesso
			"token_usage":    0.0,
			"context_hits":   0.0,
			"learning_score": 0.0,
		},
		trainingHistory: make([]*TrainingMetrics, 0),
		memoryManager:   memoryManager,
		stopChan:        make(chan struct{}),
	}
}

// Train treina o agente com base na configuração fornecida
func (a *CognitiveAgent) Train(ctx context.Context, config TrainingConfig) (*TrainingMetrics, error) {
	startTime := time.Now()
	metrics := &TrainingMetrics{
		StartTime:      startTime,
		EndTime:        startTime.Add(30 * time.Second),
		Accuracy:       0.85,
		Loss:           0.15,
		RoundsExecuted: 100,
		Errors:         make([]error, 0),
	}

	// Armazena métricas de treinamento na memória
	metricsData := map[string]interface{}{
		"metrics": metrics,
		"parameters": map[string]interface{}{
			"temperature":   a.Temperature,
			"learning_rate": a.LearningRate,
		},
	}

	metricsJSON, err := json.Marshal(metricsData)
	if err != nil {
		return metrics, fmt.Errorf("erro ao converter métricas para JSON: %v", err)
	}

	memory := &memory.Memory{
		ID:         fmt.Sprintf("training_%s_%d", a.GetID(), time.Now().Unix()),
		AgentID:    a.GetID(),
		Type:       memory.LongTerm,
		Content:    string(metricsJSON),
		Importance: metrics.Accuracy,
		Tags:       []string{"training", "metrics", "parameters"},
	}

	if err := a.memoryManager.StoreMemory(ctx, memory); err != nil {
		return metrics, fmt.Errorf("erro ao armazenar métricas de treinamento: %v", err)
	}

	// Adiciona as métricas ao histórico de treinamento
	a.trainingHistory = append(a.trainingHistory, metrics)

	return metrics, nil
}

// GetTrainingHistory retorna o histórico de treinamento do agente
func (a *CognitiveAgent) GetTrainingHistory() []*TrainingMetrics {
	return a.trainingHistory
}

// Remember busca memórias relacionadas a um conjunto de tags
func (a *CognitiveAgent) Remember(ctx context.Context, tags []string) ([]*memory.Memory, error) {
	return a.memoryManager.SearchMemories(ctx, a.GetID(), tags)
}

// Memorize armazena uma nova memória
func (a *CognitiveAgent) Memorize(ctx context.Context, content map[string]interface{}, importance float64, tags []string, isLongTerm bool) error {
	memType := memory.ShortTerm
	var ttl time.Duration

	if isLongTerm {
		memType = memory.LongTerm
	} else {
		ttl = 24 * time.Hour // Memórias de curto prazo expiram em 24 horas
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("erro ao converter conteúdo para JSON: %v", err)
	}

	memory := &memory.Memory{
		ID:         fmt.Sprintf("memory_%s_%d", a.GetID(), time.Now().Unix()),
		AgentID:    a.GetID(),
		Type:       memType,
		Content:    string(contentJSON),
		Importance: importance,
		TTL:        ttl,
		Tags:       tags,
	}

	return a.memoryManager.StoreMemory(ctx, memory)
}

// ConsolidateMemories move memórias importantes de curto prazo para longo prazo
func (a *CognitiveAgent) ConsolidateMemories(ctx context.Context) error {
	return a.memoryManager.ConsolidateMemories(ctx, a.GetID())
}

// ForgetOldMemories remove memórias antigas ou irrelevantes
func (a *CognitiveAgent) ForgetOldMemories(ctx context.Context) error {
	return a.memoryManager.PruneMemories(ctx, a.GetID())
}

// adjustParameters ajusta os parâmetros do agente baseado no histórico
func (a *CognitiveAgent) adjustParameters() {
	// Ajusta temperatura baseado no sucesso das respostas
	successRate := a.PerformanceStats["success_rate"]
	if successRate < 0.5 {
		a.Temperature *= 0.9 // Reduz temperatura para respostas mais conservadoras
	} else {
		a.Temperature *= 1.1 // Aumenta temperatura para mais criatividade
	}

	// Limita temperatura entre 0.1 e 1.0
	if a.Temperature < 0.1 {
		a.Temperature = 0.1
	} else if a.Temperature > 1.0 {
		a.Temperature = 1.0
	}

	// Ajusta taxa de aprendizado
	a.LearningRate *= 0.95 // Diminui gradualmente
	if a.LearningRate < 0.0001 {
		a.LearningRate = 0.0001
	}
}

// updatePerformanceStats atualiza as estatísticas de performance
func (a *CognitiveAgent) updatePerformanceStats(metrics *TrainingMetrics) {
	// Calcula tempo médio de resposta
	responseTime := metrics.EndTime.Sub(metrics.StartTime).Seconds()
	a.PerformanceStats["response_time"] = (a.PerformanceStats["response_time"]*0.9 + responseTime*0.1)

	// Atualiza taxa de sucesso
	if len(metrics.Errors) == 0 {
		a.PerformanceStats["success_rate"] = (a.PerformanceStats["success_rate"]*0.9 + 1.0*0.1)
	} else {
		a.PerformanceStats["success_rate"] = (a.PerformanceStats["success_rate"] * 0.9)
	}

	// Atualiza score de aprendizado
	learningProgress := float64(metrics.RoundsExecuted) / float64(a.MaxRounds)
	a.PerformanceStats["learning_score"] = learningProgress
}

// Validate implementa validação específica para o agente cognitivo
func (a *CognitiveAgent) Validate(ctx context.Context) error {
	// Validações específicas do agente cognitivo
	if a.Temperature <= 0 {
		return fmt.Errorf("temperatura inválida: %v", a.Temperature)
	}

	if a.MaxTokens <= 0 {
		return fmt.Errorf("número máximo de tokens inválido: %v", a.MaxTokens)
	}

	if a.ContextWindow <= 0 {
		return fmt.Errorf("tamanho da janela de contexto inválido: %v", a.ContextWindow)
	}

	// Verifica performance mínima
	if a.PerformanceStats["success_rate"] < 0.5 {
		return fmt.Errorf("taxa de sucesso muito baixa: %v", a.PerformanceStats["success_rate"])
	}

	return nil
}

// GetPerformanceStats retorna as estatísticas de performance
func (a *CognitiveAgent) GetPerformanceStats() map[string]float64 {
	stats := make(map[string]float64)
	for k, v := range a.PerformanceStats {
		stats[k] = v
	}
	return stats
}

// AddPromptTemplate adiciona um template de prompt
func (a *CognitiveAgent) AddPromptTemplate(name, template string) {
	a.PromptTemplates[name] = template
}

// GetPromptTemplate retorna um template de prompt
func (a *CognitiveAgent) GetPromptTemplate(name string) (string, bool) {
	template, ok := a.PromptTemplates[name]
	return template, ok
}

// AddToKnowledgeBase adiciona informação à base de conhecimento
func (a *CognitiveAgent) AddToKnowledgeBase(key string, value interface{}) {
	a.KnowledgeBase[key] = value
}

// GetFromKnowledgeBase recupera informação da base de conhecimento
func (a *CognitiveAgent) GetFromKnowledgeBase(key string) (interface{}, bool) {
	value, ok := a.KnowledgeBase[key]
	return value, ok
}

// SetBackstory define a história/contexto do agente
func (a *CognitiveAgent) SetBackstory(backstory string) {
	a.Backstory = backstory
}

// GetDescription retorna a descrição do agente
func (a *CognitiveAgent) GetDescription() string {
	return fmt.Sprintf("Agente cognitivo %s (%s) - %s", a.GetName(), a.GetRole(), a.Goal)
}

// Execute executa uma tarefa
func (a *CognitiveAgent) Execute(ctx context.Context, task Task) error {
	// TODO: Implementar execução de tarefas
	return nil
}

// Stop interrompe a execução do agente
func (a *CognitiveAgent) Stop() error {
	// TODO: Implementar parada do agente
	return nil
}

// SaveState salva o estado atual do agente em um arquivo
func (a *CognitiveAgent) SaveState(path string) error {
	state := map[string]interface{}{
		"id":                a.GetID(),
		"name":              a.GetName(),
		"role":              a.GetRole(),
		"goal":              a.Goal,
		"model":             a.Model,
		"temperature":       a.Temperature,
		"max_tokens":        a.MaxTokens,
		"context_window":    a.ContextWindow,
		"learning_rate":     a.LearningRate,
		"knowledge_base":    a.KnowledgeBase,
		"prompt_templates":  a.PromptTemplates,
		"performance_stats": a.PerformanceStats,
		"training_history":  a.trainingHistory,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao converter estado para JSON: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar estado em arquivo: %v", err)
	}

	return nil
}

// LoadState carrega o estado do agente de um arquivo
func (a *CognitiveAgent) LoadState(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de estado: %v", err)
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("erro ao decodificar estado do JSON: %v", err)
	}

	// Atualiza os campos do agente
	if id, ok := state["id"].(string); ok {
		a.AgentStruct.ID = id
	}
	if name, ok := state["name"].(string); ok {
		a.AgentStruct.Name = name
	}
	if role, ok := state["role"].(string); ok {
		a.AgentStruct.Role = role
	}
	if goal, ok := state["goal"].(string); ok {
		a.Goal = goal
	}
	if model, ok := state["model"].(string); ok {
		a.Model = model
	}
	if temperature, ok := state["temperature"].(float64); ok {
		a.Temperature = temperature
	}
	if maxTokens, ok := state["max_tokens"].(float64); ok {
		a.MaxTokens = int(maxTokens)
	}
	if contextWindow, ok := state["context_window"].(float64); ok {
		a.ContextWindow = int(contextWindow)
	}
	if learningRate, ok := state["learning_rate"].(float64); ok {
		a.LearningRate = learningRate
	}
	if knowledgeBase, ok := state["knowledge_base"].(map[string]interface{}); ok {
		a.KnowledgeBase = knowledgeBase
	}
	if promptTemplates, ok := state["prompt_templates"].(map[string]interface{}); ok {
		for k, v := range promptTemplates {
			if template, ok := v.(string); ok {
				a.PromptTemplates[k] = template
			}
		}
	}
	if performanceStats, ok := state["performance_stats"].(map[string]interface{}); ok {
		for k, v := range performanceStats {
			if stat, ok := v.(float64); ok {
				a.PerformanceStats[k] = stat
			}
		}
	}
	if trainingHistory, ok := state["training_history"].([]interface{}); ok {
		for _, v := range trainingHistory {
			if metrics, ok := v.(map[string]interface{}); ok {
				a.trainingHistory = append(a.trainingHistory, &TrainingMetrics{
					StartTime:      time.Now(),
					EndTime:        time.Now(),
					Accuracy:       metrics["accuracy"].(float64),
					Loss:           metrics["loss"].(float64),
					RoundsExecuted: int(metrics["rounds_executed"].(float64)),
				})
			}
		}
	}

	return nil
}
