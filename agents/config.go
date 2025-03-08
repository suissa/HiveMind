package agents

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Configurações do RabbitMQ
const (
	RABBITMQ_HOST = "localhost"
	RABBITMQ_PORT = 5672

	// Exchanges
	EXCHANGE_HEALTH = "health_events"
	EXCHANGE_TASK   = "task_events"

	// Filas
	QUEUE_HEALTH_MONITOR = "health_monitor"
	QUEUE_TASK_QUIZ      = "quiz_tasks"
	QUEUE_TASK_CHALLENGE = "challenge_tasks"

	// Configurações de durabilidade
	QUEUE_DURABLE     = true
	QUEUE_AUTO_DELETE = false
	QUEUE_EXCLUSIVE   = false
	QUEUE_NO_WAIT     = false

	// Configurações de mensagem
	MESSAGE_PERSISTENT = 2 // DeliveryMode 2 = persistente
)

// Configurações de escalabilidade
const (
	CPU_THRESHOLD    = 80.0 // 80% de uso de CPU
	MEMORY_THRESHOLD = 85.0 // 85% de uso de memória
	TASKS_THRESHOLD  = 100  // 100 tarefas na fila
	ERROR_THRESHOLD  = 0.05 // 5% de taxa de erro
	SCALE_COOLDOWN   = 300  // 5 minutos de cooldown entre escalas
)

// AgentConfig representa a configuração de um agente
type AgentConfig struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Role        string `yaml:"role"`
	Goal        string `yaml:"goal"`
	Model       string `yaml:"model"`
	MaxRounds   int    `yaml:"max_rounds"`
	Backstory   string `yaml:"backstory"`
}

// AgentsConfig representa a configuração de todos os agentes
type AgentsConfig struct {
	Agents []AgentConfig `yaml:"agents"`
}

// TaskConfig representa a configuração de uma tarefa
type TaskConfig struct {
	ID           string   `yaml:"id"`
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	AssignedTo   string   `yaml:"assigned_to"`
	Dependencies []string `yaml:"dependencies"`
	Priority     int      `yaml:"priority"`
	Status       string   `yaml:"status"`
	Deadline     string   `yaml:"deadline"`
}

// TasksConfig representa a configuração de todas as tarefas
type TasksConfig struct {
	Tasks []TaskConfig `yaml:"tasks"`
}

// ToolConfig representa a configuração de uma ferramenta
type ToolConfig struct {
	ID           string   `yaml:"id"`
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Capabilities []string `yaml:"capabilities"`
	Requirements []string `yaml:"requirements"`
}

// ToolsConfig representa a configuração de todas as ferramentas
type ToolsConfig struct {
	Tools map[string][]ToolConfig `yaml:"tools"`
}

// LoadAgentsConfig carrega a configuração dos agentes de um arquivo YAML
func LoadAgentsConfig(filename string) (*AgentsConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração dos agentes: %v", err)
	}

	var config AgentsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configuração dos agentes: %v", err)
	}

	return &config, nil
}

// LoadTasksConfig carrega a configuração das tarefas de um arquivo YAML
func LoadTasksConfig(filename string) (*TasksConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração das tarefas: %v", err)
	}

	var config TasksConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configuração das tarefas: %v", err)
	}

	return &config, nil
}

// LoadToolsConfig carrega a configuração das ferramentas de um arquivo YAML
func LoadToolsConfig(filename string) (*ToolsConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração das ferramentas: %v", err)
	}

	var config ToolsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configuração das ferramentas: %v", err)
	}

	return &config, nil
}
