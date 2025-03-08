package agents

import (
	"fmt"
	"sync"
	"time"
)

// TaskManager gerencia a execução de tarefas
type TaskManager struct {
	tasks      map[string]*Task
	agents     map[string]*BaseAgent
	taskQueue  []*Task
	mu         sync.RWMutex
	healthChan chan *AgentHealth
}

// AgentHealth representa o estado de saúde de um agente
type AgentHealth struct {
	AgentName      string    `json:"agent_name"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	IsProcessing   bool      `json:"is_processing"`
	CurrentTaskID  string    `json:"current_task_id"`
	ProcessingTime float64   `json:"processing_time"`
	SuccessRate    float64   `json:"success_rate"`
}

// NewTaskManager cria uma nova instância do TaskManager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks:      make(map[string]*Task),
		agents:     make(map[string]*BaseAgent),
		taskQueue:  make([]*Task, 0),
		healthChan: make(chan *AgentHealth, 100),
	}
}

// AddTask adiciona uma nova tarefa ao gerenciador
func (tm *TaskManager) AddTask(task *Task) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tasks[task.ID]; exists {
		return fmt.Errorf("tarefa já existe: %s", task.ID)
	}

	tm.tasks[task.ID] = task
	tm.taskQueue = append(tm.taskQueue, task)
	return nil
}

// GetTask retorna uma tarefa pelo ID
func (tm *TaskManager) GetTask(taskID string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, ok := tm.tasks[taskID]
	return task, ok
}

// GetNextTask retorna a próxima tarefa para um agente
func (tm *TaskManager) GetNextTask(agentID string) *Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if len(tm.taskQueue) == 0 {
		return nil
	}

	// Encontra a primeira tarefa que pode ser executada pelo agente
	for i, task := range tm.taskQueue {
		if task.AssignedTo == "" || task.AssignedTo == agentID {
			// Remove a tarefa da fila
			tm.taskQueue = append(tm.taskQueue[:i], tm.taskQueue[i+1:]...)
			task.AssignedTo = agentID
			return task
		}
	}

	return nil
}

// UpdateTaskStatus atualiza o status de uma tarefa
func (tm *TaskManager) UpdateTaskStatus(taskID string, status TaskStatus) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if task, ok := tm.tasks[taskID]; ok {
		task.Status = status
	}
}

// RegisterAgent registra um novo agente
func (tm *TaskManager) RegisterAgent(agent *BaseAgent) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.agents[agent.ID] = agent
}

// UnregisterAgent remove um agente
func (tm *TaskManager) UnregisterAgent(agentID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.agents, agentID)
}

// GetAgent retorna um agente pelo ID
func (tm *TaskManager) GetAgent(agentID string) (*BaseAgent, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	agent, ok := tm.agents[agentID]
	return agent, ok
}

// GetAllAgents retorna todos os agentes registrados
func (tm *TaskManager) GetAllAgents() []*BaseAgent {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	agents := make([]*BaseAgent, 0, len(tm.agents))
	for _, agent := range tm.agents {
		agents = append(agents, agent)
	}
	return agents
}

// EmitHealthSignal emite um sinal de saúde de um agente
func (tm *TaskManager) EmitHealthSignal(health *AgentHealth) error {
	select {
	case tm.healthChan <- health:
		return nil
	default:
		return fmt.Errorf("canal de saúde cheio")
	}
}

// GetHealthSignals retorna o canal de sinais de saúde
func (tm *TaskManager) GetHealthSignals() <-chan *AgentHealth {
	return tm.healthChan
}

// GetActiveAgentsCount retorna o número de agentes ativos
func (tm *TaskManager) GetActiveAgentsCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.agents)
}

// GetTotalTasksCount retorna o número total de tarefas
func (tm *TaskManager) GetTotalTasksCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.tasks)
}

// GetQueuedTasksCount retorna o número de tarefas na fila
func (tm *TaskManager) GetQueuedTasksCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.taskQueue)
}

// GetRunningTasksCount retorna o número de tarefas em execução
func (tm *TaskManager) GetRunningTasksCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	count := 0
	for _, task := range tm.tasks {
		if task.Status == TaskStatusRunning {
			count++
		}
	}
	return count
}

// GetCompletedTasksCount retorna o número de tarefas concluídas
func (tm *TaskManager) GetCompletedTasksCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	count := 0
	for _, task := range tm.tasks {
		if task.Status == TaskStatusComplete {
			count++
		}
	}
	return count
}
