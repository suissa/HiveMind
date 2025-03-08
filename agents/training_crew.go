package agents

import (
	"fmt"
	"sync"
	"time"

	"HiveMind/agents/memory"
)

// TrainingProject contém os detalhes do projeto de treinamento
type TrainingProject struct {
	Name           string
	Description    string
	Objectives     []string
	TargetAudience []string
	Duration       time.Duration
	Difficulty     string
	Prerequisites  []string
}

// TrainingResults contém os resultados do workflow de treinamento
type TrainingResults struct {
	Training string
	Chapters string
	Feedback string
}

// TrainingCrew representa uma equipe de agentes de treinamento
type TrainingCrew struct {
	*BaseCrew
	memoryManager memory.MemoryManager
	trainingAgent *CognitiveAgent
	chapterAgent  *CognitiveAgent
	feedbackAgent *CognitiveAgent
	accountAgent  *CognitiveAgent
	mu            sync.RWMutex
}

// NewTrainingCrew cria uma nova equipe de treinamento
func NewTrainingCrew(memoryManager memory.MemoryManager) *TrainingCrew {
	return &TrainingCrew{
		BaseCrew:      NewBaseCrew(),
		memoryManager: memoryManager,
	}
}

// AddAgent adiciona um agente à equipe
func (c *TrainingCrew) AddAgent(agent *CognitiveAgent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.BaseCrew.AddAgent(agent)

	switch agent.GetRole() {
	case "training":
		c.trainingAgent = agent
	case "chapter":
		c.chapterAgent = agent
	case "feedback":
		c.feedbackAgent = agent
	case "account":
		c.accountAgent = agent
	}
}

// ExecuteWorkflow executa o workflow de treinamento
func (c *TrainingCrew) ExecuteWorkflow(project *TrainingProject) (*TrainingResults, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.trainingAgent == nil || c.chapterAgent == nil || c.feedbackAgent == nil || c.accountAgent == nil {
		return nil, fmt.Errorf("equipe incompleta: todos os agentes são necessários")
	}

	// Emite evento de início do workflow
	c.EmitEvent(Event{
		Type:      EventWorkflowUpdate,
		Timestamp: time.Now(),
		Source:    "training_crew",
		Data: map[string]interface{}{
			"action":    "workflow_start",
			"project":   project.Name,
			"objective": project.Description,
		},
	})

	startTime := time.Now()

	// Simula a execução do workflow
	training := "Estratégia de treinamento focada em aprendizado prático"
	chapters := "Capítulos estruturados com exercícios progressivos"
	feedback := "Sistema de feedback personalizado baseado em desempenho"

	// Emite evento de conclusão do workflow
	c.EmitEvent(Event{
		Type:      EventWorkflowUpdate,
		Timestamp: time.Now(),
		Source:    "training_crew",
		Data: map[string]interface{}{
			"action":   "workflow_complete",
			"duration": time.Since(startTime).String(),
			"project":  project.Name,
			"results": map[string]interface{}{
				"Training": training,
				"Chapters": chapters,
				"Feedback": feedback,
			},
		},
	})

	return &TrainingResults{
		Training: training,
		Chapters: chapters,
		Feedback: feedback,
	}, nil
}

// GetProjectStatus retorna o status atual do projeto
func (c *TrainingCrew) GetProjectStatus() *ProjectStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Simula o status do projeto
	status := &ProjectStatus{
		Progress:       85.0,
		CompletedTasks: 17,
		TotalTasks:     20,
	}

	// Emite evento de atualização do projeto
	c.EmitEvent(Event{
		Type:      EventProjectUpdate,
		Timestamp: time.Now(),
		Source:    "training_crew",
		Data: map[string]interface{}{
			"action":          "status_update",
			"progress":        status.Progress,
			"completed_tasks": status.CompletedTasks,
			"total_tasks":     status.TotalTasks,
			"elapsed_time":    "2h30m",
			"remaining_time":  "30m",
		},
	})

	return status
}
