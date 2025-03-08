package agents

import (
	"context"
)

// Agent define a interface básica para todos os agentes
type Agent interface {
	// Identificação
	GetID() string
	GetName() string
	GetDescription() string
	GetRole() string

	// Treinamento
	Train(ctx context.Context, config TrainingConfig) (*TrainingMetrics, error)
	GetTrainingHistory() []*TrainingMetrics

	// Estado
	SaveState(path string) error
	LoadState(path string) error
	Validate(ctx context.Context) error

	// Execução
	Execute(ctx context.Context, task Task) error
	Stop() error
}

// Agent representa a estrutura base de um agente
type AgentStruct struct {
	ID              string
	Name            string
	Role            string
	Goal            string
	AllowDelegation bool
	Model           string
	Backstory       string
}

// GetID retorna o ID do agente
func (a *AgentStruct) GetID() string {
	return a.ID
}

// GetName retorna o nome do agente
func (a *AgentStruct) GetName() string {
	return a.Name
}

// GetRole retorna o papel do agente
func (a *AgentStruct) GetRole() string {
	return a.Role
}

// Clone cria uma cópia do agente
func (a *AgentStruct) Clone() *AgentStruct {
	return &AgentStruct{
		ID:              a.ID,
		Name:            a.Name,
		Role:            a.Role,
		Goal:            a.Goal,
		AllowDelegation: a.AllowDelegation,
		Model:           a.Model,
		Backstory:       a.Backstory,
	}
}
