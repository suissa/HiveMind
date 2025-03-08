package agents

import (
	"sync"
)

// BaseCrew fornece funcionalidade básica para equipes de agentes
type BaseCrew struct {
	agents        []Agent
	eventHandlers map[EventType][]EventHandler
	anyHandlers   []EventHandler
	mu            sync.RWMutex
}

// NewBaseCrew cria uma nova instância de BaseCrew
func NewBaseCrew() *BaseCrew {
	return &BaseCrew{
		agents:        make([]Agent, 0),
		eventHandlers: make(map[EventType][]EventHandler),
		anyHandlers:   make([]EventHandler, 0),
	}
}

// AddAgent adiciona um agente à equipe
func (c *BaseCrew) AddAgent(agent Agent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.agents = append(c.agents, agent)

	// Emite evento de adição de agente
	c.EmitEvent(Event{
		Type:   EventAgentAction,
		Source: "base_crew",
		Data: map[string]interface{}{
			"action":     "add_agent",
			"agent_id":   agent.GetID(),
			"agent_name": agent.GetName(),
			"agent_role": agent.GetRole(),
		},
	})
}

// OnEvent registra um handler para um tipo específico de evento
func (c *BaseCrew) OnEvent(eventType EventType, handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.eventHandlers[eventType] == nil {
		c.eventHandlers[eventType] = make([]EventHandler, 0)
	}
	c.eventHandlers[eventType] = append(c.eventHandlers[eventType], handler)
}

// OnAnyEvent registra um handler para todos os tipos de eventos
func (c *BaseCrew) OnAnyEvent(handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.anyHandlers = append(c.anyHandlers, handler)
}

// EmitEvent emite um evento para todos os handlers registrados
func (c *BaseCrew) EmitEvent(event Event) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Notifica handlers específicos do tipo de evento
	if handlers, ok := c.eventHandlers[event.Type]; ok {
		for _, handler := range handlers {
			handler(event)
		}
	}

	// Notifica handlers genéricos
	for _, handler := range c.anyHandlers {
		handler(event)
	}
}

// GetAgents retorna a lista de agentes na equipe
func (c *BaseCrew) GetAgents() []Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	agents := make([]Agent, len(c.agents))
	copy(agents, c.agents)
	return agents
}
