package agents

import (
	"time"
)

// TaskStatus representa o estado de uma tarefa
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusComplete  TaskStatus = "complete"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Task representa uma tarefa a ser executada por um agente
type Task struct {
	ID           string                 // Identificador único da tarefa
	Type         string                 // Tipo da tarefa
	Description  string                 // Descrição da tarefa
	Input        map[string]interface{} // Dados de entrada
	Output       map[string]interface{} // Dados de saída
	Status       TaskStatus             // Estado atual da tarefa
	Priority     int                    // Prioridade da tarefa (maior = mais prioritário)
	CreatedAt    time.Time              // Data de criação
	StartedAt    *time.Time             // Data de início
	FinishedAt   *time.Time             // Data de conclusão
	AssignedTo   string                 // ID do agente designado
	Error        error                  // Erro ocorrido durante execução
	Retries      int                    // Número de tentativas realizadas
	MaxRetries   int                    // Número máximo de tentativas permitidas
	Timeout      time.Duration          // Tempo máximo de execução
	Dependencies []string               // IDs das tarefas que precisam ser concluídas antes
}

// NewTask cria uma nova tarefa
func NewTask(id, taskType, description string, input map[string]interface{}) *Task {
	return &Task{
		ID:           id,
		Type:         taskType,
		Description:  description,
		Input:        input,
		Output:       make(map[string]interface{}),
		Status:       TaskStatusPending,
		Priority:     1,
		CreatedAt:    time.Now(),
		MaxRetries:   3,
		Timeout:      5 * time.Minute,
		Dependencies: make([]string, 0),
	}
}

// SetPriority define a prioridade da tarefa
func (t *Task) SetPriority(priority int) {
	t.Priority = priority
}

// SetTimeout define o timeout da tarefa
func (t *Task) SetTimeout(timeout time.Duration) {
	t.Timeout = timeout
}

// SetMaxRetries define o número máximo de tentativas
func (t *Task) SetMaxRetries(maxRetries int) {
	t.MaxRetries = maxRetries
}

// AddDependency adiciona uma dependência
func (t *Task) AddDependency(taskID string) {
	t.Dependencies = append(t.Dependencies, taskID)
}

// SetOutput define o resultado da tarefa
func (t *Task) SetOutput(output map[string]interface{}) {
	t.Output = output
}

// SetError define o erro da tarefa
func (t *Task) SetError(err error) {
	t.Error = err
}

// Start marca o início da execução da tarefa
func (t *Task) Start() {
	now := time.Now()
	t.StartedAt = &now
	t.Status = TaskStatusRunning
}

// Complete marca a tarefa como concluída
func (t *Task) Complete() {
	now := time.Now()
	t.FinishedAt = &now
	t.Status = TaskStatusComplete
}

// Fail marca a tarefa como falha
func (t *Task) Fail(err error) {
	now := time.Now()
	t.FinishedAt = &now
	t.Status = TaskStatusFailed
	t.Error = err
}

// Cancel marca a tarefa como cancelada
func (t *Task) Cancel() {
	now := time.Now()
	t.FinishedAt = &now
	t.Status = TaskStatusCancelled
}

// CanRetry verifica se a tarefa pode ser reexecutada
func (t *Task) CanRetry() bool {
	return t.Status == TaskStatusFailed && t.Retries < t.MaxRetries
}

// HasTimedOut verifica se a tarefa excedeu o timeout
func (t *Task) HasTimedOut() bool {
	if t.StartedAt == nil {
		return false
	}
	return time.Since(*t.StartedAt) > t.Timeout
}

// IsPending verifica se a tarefa está pendente
func (t *Task) IsPending() bool {
	return t.Status == TaskStatusPending
}

// IsRunning verifica se a tarefa está em execução
func (t *Task) IsRunning() bool {
	return t.Status == TaskStatusRunning
}

// IsComplete verifica se a tarefa foi concluída
func (t *Task) IsComplete() bool {
	return t.Status == TaskStatusComplete
}

// IsFailed verifica se a tarefa falhou
func (t *Task) IsFailed() bool {
	return t.Status == TaskStatusFailed
}

// IsCancelled verifica se a tarefa foi cancelada
func (t *Task) IsCancelled() bool {
	return t.Status == TaskStatusCancelled
}

// Duration retorna a duração da execução da tarefa
func (t *Task) Duration() time.Duration {
	if t.StartedAt == nil {
		return 0
	}
	if t.FinishedAt == nil {
		return time.Since(*t.StartedAt)
	}
	return t.FinishedAt.Sub(*t.StartedAt)
}
