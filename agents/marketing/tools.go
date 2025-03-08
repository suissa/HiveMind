package marketing

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// ToolConfig contém a configuração de uma ferramenta
type ToolConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	APIKey      string   `yaml:"api_key,omitempty"`
	MaxResults  int      `yaml:"max_results,omitempty"`
	Timeout     int      `yaml:"timeout,omitempty"`
	Libraries   []string `yaml:"libraries,omitempty"`
	MaxMemory   string   `yaml:"max_memory,omitempty"`
	Templates   []string `yaml:"templates,omitempty"`
	DataSources []string `yaml:"data_sources,omitempty"`
	Techniques  []string `yaml:"techniques,omitempty"`
}

// ToolsConfig contém a configuração de todas as ferramentas
type ToolsConfig struct {
	WebSearch           ToolConfig `yaml:"web_search"`
	DataAnalysis        ToolConfig `yaml:"data_analysis"`
	ProjectAnalysis     ToolConfig `yaml:"project_analysis"`
	RequirementsMapping ToolConfig `yaml:"requirements_mapping"`
	StrategyPlanning    ToolConfig `yaml:"strategy_planning"`
	MarketAnalysis      ToolConfig `yaml:"market_analysis"`
	CreativeIdeation    ToolConfig `yaml:"creative_ideation"`
	AudienceAnalysis    ToolConfig `yaml:"audience_analysis"`
	Copywriting         ToolConfig `yaml:"copywriting"`
	ContentOptimization ToolConfig `yaml:"content_optimization"`
}

// LoadToolsConfig carrega a configuração das ferramentas do arquivo YAML
func LoadToolsConfig(filename string) (*ToolsConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração das ferramentas: %v", err)
	}

	// Substitui variáveis de ambiente
	content := os.ExpandEnv(string(data))

	var config ToolsConfig
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configuração das ferramentas: %v", err)
	}

	return &config, nil
}

// GetToolConfig retorna a configuração de uma ferramenta específica
func (c *ToolsConfig) GetToolConfig(toolName string) (*ToolConfig, error) {
	switch strings.ToLower(toolName) {
	case "web_search":
		return &c.WebSearch, nil
	case "data_analysis":
		return &c.DataAnalysis, nil
	case "project_analysis":
		return &c.ProjectAnalysis, nil
	case "requirements_mapping":
		return &c.RequirementsMapping, nil
	case "strategy_planning":
		return &c.StrategyPlanning, nil
	case "market_analysis":
		return &c.MarketAnalysis, nil
	case "creative_ideation":
		return &c.CreativeIdeation, nil
	case "audience_analysis":
		return &c.AudienceAnalysis, nil
	case "copywriting":
		return &c.Copywriting, nil
	case "content_optimization":
		return &c.ContentOptimization, nil
	default:
		return nil, fmt.Errorf("ferramenta não encontrada: %s", toolName)
	}
}

// GetToolsByCategory retorna as ferramentas de uma categoria específica
func (c *ToolsConfig) GetToolsByCategory(category string) []ToolConfig {
	var tools []ToolConfig

	switch strings.ToLower(category) {
	case "research":
		tools = append(tools, c.WebSearch, c.DataAnalysis, c.MarketAnalysis)
	case "planning":
		tools = append(tools, c.ProjectAnalysis, c.RequirementsMapping, c.StrategyPlanning)
	case "creative":
		tools = append(tools, c.CreativeIdeation, c.Copywriting, c.ContentOptimization)
	case "analysis":
		tools = append(tools, c.DataAnalysis, c.MarketAnalysis, c.AudienceAnalysis)
	}

	return tools
}

// GetAllTools retorna todas as ferramentas disponíveis
func (c *ToolsConfig) GetAllTools() []ToolConfig {
	return []ToolConfig{
		c.WebSearch,
		c.DataAnalysis,
		c.ProjectAnalysis,
		c.RequirementsMapping,
		c.StrategyPlanning,
		c.MarketAnalysis,
		c.CreativeIdeation,
		c.AudienceAnalysis,
		c.Copywriting,
		c.ContentOptimization,
	}
}
