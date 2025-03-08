package agents

import (
	"encoding/json"
	"time"
)

// EventType representa o tipo de evento
type EventType string

const (
	EventAgentAction     EventType = "agent_action"
	EventTaskUpdate      EventType = "task_update"
	EventWorkflowUpdate  EventType = "workflow_update"
	EventProjectUpdate   EventType = "project_update"
	EventMemoryOperation EventType = "memory_operation"
)

// Event representa um evento no sistema
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// ToJSON converte o evento para uma string JSON formatada
func (e Event) ToJSON() string {
	bytes, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

// EventHandler é uma função que lida com eventos
type EventHandler func(Event)

// EventListener é uma função que processa eventos
type EventListener func(Event)

// EventEmitter gerencia a emissão e escuta de eventos
type EventEmitter struct {
	listeners map[EventType][]EventListener
}

// NewEventEmitter cria um novo emissor de eventos
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		listeners: make(map[EventType][]EventListener),
	}
}

// On registra um listener para um tipo específico de evento
func (e *EventEmitter) On(eventType EventType, listener EventListener) {
	e.listeners[eventType] = append(e.listeners[eventType], listener)
}

// OnAny registra um listener para todos os tipos de eventos
func (e *EventEmitter) OnAny(listener EventListener) {
	for _, eventType := range []EventType{
		EventAgentAction,
		EventTaskUpdate,
		EventMemoryOperation,
		EventWorkflowUpdate,
		EventProjectUpdate,
	} {
		e.On(eventType, listener)
	}
}

// Emit emite um evento para todos os listeners registrados
func (e *EventEmitter) Emit(event Event) {
	if listeners, ok := e.listeners[event.Type]; ok {
		for _, listener := range listeners {
			go listener(event)
		}
	}
}
