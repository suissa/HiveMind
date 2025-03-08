package marketing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"HiveMind/agents"
	"HiveMind/agents/memory"
)

// MarketStrategy representa uma estratégia de marketing
type MarketStrategy struct {
	Name     string   `json:"name"`     // Nome da estratégia
	Tactics  []string `json:"tactics"`  // Lista de táticas
	Channels []string `json:"channels"` // Lista de canais
	KPIs     []string `json:"kpis"`     // Lista de KPIs
}

// CampaignIdea representa uma ideia de campanha
type CampaignIdea struct {
	Name        string `json:"name"`        // Nome da campanha
	Description string `json:"description"` // Descrição da campanha
	Audience    string `json:"audience"`    // Público-alvo
	Channel     string `json:"channel"`     // Canal principal
}

// Copy representa um texto publicitário
type Copy struct {
	Title string `json:"title"` // Título do texto
	Body  string `json:"body"`  // Corpo do texto
}

// MarketingPostsCrew gerencia a equipe de marketing
type MarketingPostsCrew struct {
	leadMarketAnalyst        *agents.CognitiveAgent
	chiefMarketingStrategist *agents.CognitiveAgent
	creativeContentCreator   *agents.CognitiveAgent
	memoryManager            memory.MemoryManager
	ctx                      context.Context
}

// NewMarketingPostsCrew cria uma nova equipe de marketing
func NewMarketingPostsCrew(ctx context.Context, memManager memory.MemoryManager) *MarketingPostsCrew {
	// Carrega as configurações
	agentsConfig, err := LoadAgentsConfig("config/agents.yaml")
	if err != nil {
		log.Printf("Erro ao carregar configuração dos agentes: %v", err)
		agentsConfig = &AgentsConfig{
			LeadMarketAnalyst: AgentConfig{
				Name:          "Analista Líder de Mercado",
				Role:          "analista",
				Goal:          "Realizar pesquisas e análises de mercado aprofundadas",
				Backstory:     "Sou um analista experiente com foco em identificar tendências e oportunidades de mercado.",
				Model:         "gpt-4",
				Temperature:   0.7,
				MaxTokens:     2048,
				ContextWindow: 4096,
			},
			ChiefMarketingStrategist: AgentConfig{
				Name:          "Estrategista Chefe de Marketing",
				Role:          "estrategista",
				Goal:          "Desenvolver estratégias de marketing eficazes",
				Backstory:     "Sou um estrategista experiente com histórico comprovado em campanhas de sucesso.",
				Model:         "gpt-4",
				Temperature:   0.7,
				MaxTokens:     2048,
				ContextWindow: 4096,
			},
			CreativeContentCreator: AgentConfig{
				Name:          "Criador de Conteúdo Criativo",
				Role:          "criador",
				Goal:          "Criar conteúdo criativo e persuasivo",
				Backstory:     "Sou um criador de conteúdo apaixonado por contar histórias memoráveis.",
				Model:         "gpt-4",
				Temperature:   0.8,
				MaxTokens:     2048,
				ContextWindow: 4096,
			},
		}
	}

	crew := &MarketingPostsCrew{
		memoryManager: memManager,
		ctx:           ctx,
	}

	// Cria o analista líder de mercado
	crew.leadMarketAnalyst = agents.NewCognitiveAgent(
		"lead-analyst",
		agentsConfig.LeadMarketAnalyst.Name,
		"Especialista em análise de mercado e pesquisa",
		5, // maxRounds
		agentsConfig.LeadMarketAnalyst.Model,
		agentsConfig.LeadMarketAnalyst.Role,
		agentsConfig.LeadMarketAnalyst.Goal,
		memManager,
	)
	crew.leadMarketAnalyst.Temperature = agentsConfig.LeadMarketAnalyst.Temperature
	crew.leadMarketAnalyst.MaxTokens = agentsConfig.LeadMarketAnalyst.MaxTokens
	crew.leadMarketAnalyst.ContextWindow = agentsConfig.LeadMarketAnalyst.ContextWindow
	crew.leadMarketAnalyst.SetBackstory(agentsConfig.LeadMarketAnalyst.Backstory)

	// Cria o estrategista chefe de marketing
	crew.chiefMarketingStrategist = agents.NewCognitiveAgent(
		"chief-strategist",
		agentsConfig.ChiefMarketingStrategist.Name,
		"Especialista em estratégias de marketing",
		5, // maxRounds
		agentsConfig.ChiefMarketingStrategist.Model,
		agentsConfig.ChiefMarketingStrategist.Role,
		agentsConfig.ChiefMarketingStrategist.Goal,
		memManager,
	)
	crew.chiefMarketingStrategist.Temperature = agentsConfig.ChiefMarketingStrategist.Temperature
	crew.chiefMarketingStrategist.MaxTokens = agentsConfig.ChiefMarketingStrategist.MaxTokens
	crew.chiefMarketingStrategist.ContextWindow = agentsConfig.ChiefMarketingStrategist.ContextWindow
	crew.chiefMarketingStrategist.SetBackstory(agentsConfig.ChiefMarketingStrategist.Backstory)

	// Cria o criador de conteúdo
	crew.creativeContentCreator = agents.NewCognitiveAgent(
		"content-creator",
		agentsConfig.CreativeContentCreator.Name,
		"Especialista em criação de conteúdo envolvente",
		5, // maxRounds
		agentsConfig.CreativeContentCreator.Model,
		agentsConfig.CreativeContentCreator.Role,
		agentsConfig.CreativeContentCreator.Goal,
		memManager,
	)
	crew.creativeContentCreator.Temperature = agentsConfig.CreativeContentCreator.Temperature
	crew.creativeContentCreator.MaxTokens = agentsConfig.CreativeContentCreator.MaxTokens
	crew.creativeContentCreator.ContextWindow = agentsConfig.CreativeContentCreator.ContextWindow
	crew.creativeContentCreator.SetBackstory(agentsConfig.CreativeContentCreator.Backstory)

	return crew
}

// ResearchTask executa a tarefa de pesquisa
func (c *MarketingPostsCrew) ResearchTask(topic string) error {
	research := map[string]interface{}{
		"topic":     topic,
		"timestamp": time.Now(),
	}

	// Memoriza a pesquisa
	err := c.leadMarketAnalyst.Memorize(c.ctx, research, 0.8, []string{"research", topic}, true)
	if err != nil {
		return fmt.Errorf("erro ao memorizar pesquisa: %v", err)
	}

	return nil
}

// ProjectUnderstandingTask executa a tarefa de compreensão do projeto
func (c *MarketingPostsCrew) ProjectUnderstandingTask(projectDetails map[string]interface{}) error {
	// Memoriza os detalhes do projeto
	err := c.chiefMarketingStrategist.Memorize(c.ctx, projectDetails, 0.9, []string{"project", "understanding"}, true)
	if err != nil {
		return fmt.Errorf("erro ao memorizar detalhes do projeto: %v", err)
	}

	return nil
}

// MarketingStrategyTask desenvolve a estratégia de marketing
func (c *MarketingPostsCrew) MarketingStrategyTask() (*MarketStrategy, error) {
	// Busca memórias relacionadas ao projeto
	memories, err := c.chiefMarketingStrategist.Remember(c.ctx, []string{"project", "research"})
	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar memórias: %v", err)
	}

	// Analisa as memórias para criar a estratégia
	var channels []string
	var tactics []string
	var kpis []string
	var projectName string
	var objective string

	for _, mem := range memories {
		if channels, ok := mem.Content["channels"].([]string); ok {
			channels = append(channels, channels...)
		}
		if objective, ok := mem.Content["objective"].(string); ok {
			objective = objective
		}
		if name, ok := mem.Content["name"].(string); ok {
			projectName = name
		}
	}

	// Se não encontrou dados nas memórias, usa valores padrão
	if len(channels) == 0 {
		channels = []string{"LinkedIn", "Twitter", "Email"}
	}
	if len(tactics) == 0 {
		tactics = []string{"Content Marketing", "Social Media", "Email Marketing"}
	}
	if len(kpis) == 0 {
		kpis = []string{"Engagement Rate", "Conversion Rate", "ROI"}
	}
	if projectName == "" {
		projectName = "Estratégia de Marketing Digital"
	}

	// Cria a estratégia baseada nas memórias
	strategy := &MarketStrategy{
		Name:     projectName,
		Tactics:  tactics,
		Channels: channels,
		KPIs:     kpis,
	}

	// Memoriza a estratégia
	strategyData := map[string]interface{}{
		"strategy":  strategy,
		"timestamp": time.Now(),
		"objective": objective,
	}
	err = c.chiefMarketingStrategist.Memorize(c.ctx, strategyData, 0.9, []string{"strategy", "marketing"}, true)
	if err != nil {
		return nil, fmt.Errorf("erro ao memorizar estratégia: %v", err)
	}

	return strategy, nil
}

// CampaignIdeaTask desenvolve uma ideia de campanha
func (c *MarketingPostsCrew) CampaignIdeaTask() (*CampaignIdea, error) {
	// Busca memórias relacionadas à estratégia
	memories, err := c.creativeContentCreator.Remember(c.ctx, []string{"strategy"})
	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar memórias: %v", err)
	}

	// Analisa as memórias para criar a ideia de campanha
	var audience string
	var channel string
	var description string
	var objective string

	for _, mem := range memories {
		if strategy, ok := mem.Content["strategy"].(*MarketStrategy); ok {
			if len(strategy.Channels) > 0 {
				channel = strategy.Channels[0] // Usa o primeiro canal como principal
			}
		}
		if target, ok := mem.Content["target"].(string); ok {
			audience = target
		}
		if obj, ok := mem.Content["objective"].(string); ok {
			objective = obj
			description = fmt.Sprintf("Série de posts interativos focados em %s", objective)
		}
	}

	// Se não encontrou dados nas memórias, usa valores padrão
	if audience == "" {
		audience = "Profissionais de Marketing Digital"
	}
	if channel == "" {
		channel = "LinkedIn"
	}
	if description == "" {
		description = "Série de posts interativos focados em educação e engajamento"
	}

	// Cria a ideia de campanha baseada nas memórias
	idea := &CampaignIdea{
		Name:        fmt.Sprintf("Campanha de %s", objective),
		Description: description,
		Audience:    audience,
		Channel:     channel,
	}

	// Memoriza a ideia
	ideaData := map[string]interface{}{
		"idea":      idea,
		"timestamp": time.Now(),
		"objective": objective,
	}
	err = c.creativeContentCreator.Memorize(c.ctx, ideaData, 0.8, []string{"campaign", "idea"}, true)
	if err != nil {
		return nil, fmt.Errorf("erro ao memorizar ideia: %v", err)
	}

	return idea, nil
}

// CopyCreationTask cria o texto publicitário
func (c *MarketingPostsCrew) CopyCreationTask() (*Copy, error) {
	// Busca memórias relacionadas à campanha e estratégia
	memories, err := c.creativeContentCreator.Remember(c.ctx, []string{"campaign", "strategy"})
	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar memórias: %v", err)
	}

	// Analisa as memórias para criar o texto
	var title string
	var body string
	var audience string
	var objective string
	var tactics []string

	for _, mem := range memories {
		if idea, ok := mem.Content["idea"].(*CampaignIdea); ok {
			audience = idea.Audience
		}
		if strategy, ok := mem.Content["strategy"].(*MarketStrategy); ok {
			tactics = strategy.Tactics
		}
		if obj, ok := mem.Content["objective"].(string); ok {
			objective = obj
		}
	}

	// Cria o título baseado no objetivo e público
	if audience != "" && objective != "" {
		title = fmt.Sprintf("Domine %s para %s", objective, audience)
	} else if objective != "" {
		title = fmt.Sprintf("Domine %s", objective)
	} else {
		title = "Domine o Marketing Digital"
	}

	// Cria o corpo do texto baseado nas táticas
	if len(tactics) > 0 {
		body = fmt.Sprintf("Descubra como usar %s e outras estratégias para revolucionar seus resultados em %s...",
			tactics[0], objective)
	} else {
		body = "Descubra as estratégias que estão revolucionando o mercado..."
	}

	// Cria o texto baseado nas memórias
	copy := &Copy{
		Title: title,
		Body:  body,
	}

	// Memoriza o texto
	copyData := map[string]interface{}{
		"copy":      copy,
		"timestamp": time.Now(),
		"audience":  audience,
		"objective": objective,
	}
	err = c.creativeContentCreator.Memorize(c.ctx, copyData, 0.7, []string{"copy", "content"}, true)
	if err != nil {
		return nil, fmt.Errorf("erro ao memorizar texto: %v", err)
	}

	return copy, nil
}

// ExecuteWorkflow executa o fluxo completo de trabalho
func (c *MarketingPostsCrew) ExecuteWorkflow(projectDetails map[string]interface{}) (*WorkflowResult, error) {
	// 1. Pesquisa
	if err := c.ResearchTask("marketing digital"); err != nil {
		return nil, fmt.Errorf("erro na pesquisa: %v", err)
	}

	// 2. Compreensão do projeto
	if err := c.ProjectUnderstandingTask(projectDetails); err != nil {
		return nil, fmt.Errorf("erro na compreensão do projeto: %v", err)
	}

	// 3. Estratégia de marketing
	strategy, err := c.MarketingStrategyTask()
	if err != nil {
		return nil, fmt.Errorf("erro na estratégia: %v", err)
	}

	// 4. Ideia de campanha
	idea, err := c.CampaignIdeaTask()
	if err != nil {
		return nil, fmt.Errorf("erro na ideia de campanha: %v", err)
	}

	// 5. Criação do texto
	copy, err := c.CopyCreationTask()
	if err != nil {
		return nil, fmt.Errorf("erro na criação do texto: %v", err)
	}

	// Consolida as memórias importantes
	if err := c.ConsolidateAllMemories(); err != nil {
		return nil, fmt.Errorf("erro ao consolidar memórias: %v", err)
	}

	return &WorkflowResult{
		Strategy: strategy,
		Campaign: idea,
		Copy:     copy,
	}, nil
}

// ConsolidateAllMemories consolida as memórias de todos os agentes
func (c *MarketingPostsCrew) ConsolidateAllMemories() error {
	agents := []*agents.CognitiveAgent{
		c.leadMarketAnalyst,
		c.chiefMarketingStrategist,
		c.creativeContentCreator,
	}

	for _, agent := range agents {
		if err := agent.ConsolidateMemories(c.ctx); err != nil {
			return fmt.Errorf("erro ao consolidar memórias do agente %s: %v", agent.GetRole(), err)
		}
	}

	return nil
}

// WorkflowResult contém os resultados do fluxo de trabalho
type WorkflowResult struct {
	Strategy *MarketStrategy `json:"strategy"`
	Campaign *CampaignIdea   `json:"campaign"`
	Copy     *Copy           `json:"copy"`
}

// String retorna uma representação em string do resultado
func (r *WorkflowResult) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}
