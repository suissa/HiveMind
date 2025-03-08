package marketing

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// AgentConfig contém a configuração de um agente
type AgentConfig struct {
	Name          string  `yaml:"name"`
	Role          string  `yaml:"role"`
	Goal          string  `yaml:"goal"`
	Backstory     string  `yaml:"backstory"`
	Model         string  `yaml:"model"`
	Temperature   float64 `yaml:"temperature"`
	MaxTokens     int     `yaml:"max_tokens"`
	ContextWindow int     `yaml:"context_window"`
}

// TaskConfig contém a configuração de uma tarefa
type TaskConfig struct {
	Description    string   `yaml:"description"`
	ExpectedOutput string   `yaml:"expected_output"`
	Context        []string `yaml:"context"`
	Tools          []string `yaml:"tools"`
}

// AgentsConfig contém a configuração de todos os agentes
type AgentsConfig struct {
	LeadMarketAnalyst        AgentConfig `yaml:"lead_market_analyst"`
	ChiefMarketingStrategist AgentConfig `yaml:"chief_marketing_strategist"`
	CreativeContentCreator   AgentConfig `yaml:"creative_content_creator"`
}

// TasksConfig contém a configuração de todas as tarefas
type TasksConfig struct {
	ResearchTask             TaskConfig `yaml:"research_task"`
	ProjectUnderstandingTask TaskConfig `yaml:"project_understanding_task"`
	MarketingStrategyTask    TaskConfig `yaml:"marketing_strategy_task"`
	CampaignIdeaTask         TaskConfig `yaml:"campaign_idea_task"`
	CopyCreationTask         TaskConfig `yaml:"copy_creation_task"`
}

// LoadAgentsConfig carrega a configuração dos agentes do arquivo YAML
func LoadAgentsConfig(filename string) (*AgentsConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração dos agentes: %v", err)
	}

	var config AgentsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configuração dos agentes: %v", err)
	}

	return &config, nil
}

// LoadTasksConfig carrega a configuração das tarefas do arquivo YAML
func LoadTasksConfig(filename string) (*TasksConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração das tarefas: %v", err)
	}

	var config TasksConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configuração das tarefas: %v", err)
	}

	return &config, nil
}
