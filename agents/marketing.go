package agents

import (
	"time"

	"HiveMind/agents/memory"
)

// MarketingProject representa um projeto de marketing
type MarketingProject struct {
	Name           string
	Objective      string
	TargetAudience []string
	Budget         float64
	Duration       time.Duration
	Channels       []string
	Constraints    []string
	Tasks          []TaskConfig
}

// AddTask adiciona uma tarefa ao projeto
func (p *MarketingProject) AddTask(task TaskConfig) {
	p.Tasks = append(p.Tasks, task)
}

// MarketingCrew representa uma equipe de marketing
type MarketingCrew struct {
	agents     []*CognitiveAgent
	memManager memory.MemoryManager
	emitter    *EventEmitter
	project    *MarketingProject
	startTime  time.Time
	taskStatus map[string]string
}

// NewMarketingCrew cria uma nova equipe de marketing
func NewMarketingCrew(memManager memory.MemoryManager) *MarketingCrew {
	return &MarketingCrew{
		agents:     make([]*CognitiveAgent, 0),
		memManager: memManager,
		emitter:    NewEventEmitter(),
		taskStatus: make(map[string]string),
	}
}

// AddAgent adiciona um agente à equipe
func (c *MarketingCrew) AddAgent(agent *CognitiveAgent) {
	c.agents = append(c.agents, agent)
	c.emitter.Emit(Event{
		Type:      EventAgentAction,
		Timestamp: time.Now(),
		Source:    "marketing_crew",
		Data: map[string]interface{}{
			"action":     "add_agent",
			"agent_id":   agent.GetID(),
			"agent_name": agent.GetName(),
			"agent_role": agent.GetRole(),
		},
	})
}

// OnEvent registra um listener para um tipo específico de evento
func (c *MarketingCrew) OnEvent(eventType EventType, listener EventListener) {
	c.emitter.On(eventType, listener)
}

// OnAnyEvent registra um listener para todos os tipos de eventos
func (c *MarketingCrew) OnAnyEvent(listener EventListener) {
	c.emitter.OnAny(listener)
}

// WorkflowResults contém os resultados do workflow
type WorkflowResults struct {
	Strategy string
	Campaign string
	Copy     string
}

// ExecuteWorkflow executa o workflow do projeto
func (c *MarketingCrew) ExecuteWorkflow(project *MarketingProject) (*WorkflowResults, error) {
	c.project = project
	c.startTime = time.Now()

	c.emitter.Emit(Event{
		Type:      EventWorkflowUpdate,
		Timestamp: time.Now(),
		Source:    "marketing_crew",
		Data: map[string]interface{}{
			"action":    "workflow_start",
			"project":   project.Name,
			"objective": project.Objective,
		},
	})

	// Inicializa o status das tarefas
	for _, task := range project.Tasks {
		c.taskStatus[task.ID] = task.Status
	}

	// TODO: Implementar a lógica real do workflow
	// Por enquanto, simula o processamento das tarefas
	for _, task := range project.Tasks {
		c.processTask(task)
	}

	results := &WorkflowResults{
		Strategy: "Estratégia de marketing digital focada em sustentabilidade",
		Campaign: "Campanha 'Verde é o Novo Luxo'",
		Copy:     "Descubra como luxo e sustentabilidade podem andar juntos",
	}

	c.emitter.Emit(Event{
		Type:      EventWorkflowUpdate,
		Timestamp: time.Now(),
		Source:    "marketing_crew",
		Data: map[string]interface{}{
			"action":   "workflow_complete",
			"project":  project.Name,
			"results":  results,
			"duration": time.Since(c.startTime).String(),
		},
	})

	return results, nil
}

// processTask processa uma tarefa do projeto
func (c *MarketingCrew) processTask(task TaskConfig) {
	c.emitter.Emit(Event{
		Type:      EventTaskUpdate,
		Timestamp: time.Now(),
		Source:    "marketing_crew",
		Data: map[string]interface{}{
			"action":      "task_start",
			"task_id":     task.ID,
			"task_name":   task.Name,
			"assigned_to": task.AssignedTo,
		},
	})

	// Simula o processamento da tarefa
	time.Sleep(1 * time.Second)
	c.taskStatus[task.ID] = "completed"

	c.emitter.Emit(Event{
		Type:      EventTaskUpdate,
		Timestamp: time.Now(),
		Source:    "marketing_crew",
		Data: map[string]interface{}{
			"action":      "task_complete",
			"task_id":     task.ID,
			"task_name":   task.Name,
			"assigned_to": task.AssignedTo,
		},
	})
}

// GetProjectStatus retorna o status atual do projeto
func (c *MarketingCrew) GetProjectStatus() *ProjectStatus {
	if c.project == nil {
		return &ProjectStatus{}
	}

	completedTasks := 0
	for _, status := range c.taskStatus {
		if status == "completed" {
			completedTasks++
		}
	}

	totalTasks := len(c.project.Tasks)
	progress := float64(completedTasks) / float64(totalTasks) * 100
	elapsed := time.Since(c.startTime)
	remaining := c.project.Duration - elapsed

	status := &ProjectStatus{
		Progress:       progress,
		CompletedTasks: completedTasks,
		TotalTasks:     totalTasks,
		ElapsedTime:    elapsed,
		RemainingTime:  remaining,
	}

	c.emitter.Emit(Event{
		Type:      EventProjectUpdate,
		Timestamp: time.Now(),
		Source:    "marketing_crew",
		Data: map[string]interface{}{
			"action":          "status_update",
			"progress":        status.Progress,
			"completed_tasks": status.CompletedTasks,
			"total_tasks":     status.TotalTasks,
			"elapsed_time":    status.ElapsedTime.String(),
			"remaining_time":  status.RemainingTime.String(),
		},
	})

	return status
}
