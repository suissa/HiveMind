package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// LLMRouter é responsável por integrar com o RouteLLM
type LLMRouter struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	inputQueue  string
	taskQueue   string
	resultQueue string
}

// TaskRequest representa uma solicitação de tarefa
type TaskRequest struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// SubTask representa uma subtarefa gerada pela LLM
type SubTask struct {
	ID          string                 `json:"id"`
	ParentID    string                 `json:"parent_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      string                 `json:"status"`
}

// NewLLMRouter cria uma nova instância do LLMRouter
func NewLLMRouter(conn *amqp.Connection) (*LLMRouter, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("erro ao criar canal: %v", err)
	}

	// Declarando as filas
	inputQueue := "llm_input"
	taskQueue := "llm_tasks"
	resultQueue := "llm_results"

	// Fila de entrada
	_, err = channel.QueueDeclare(
		inputQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao declarar fila de entrada: %v", err)
	}

	// Fila de tarefas
	_, err = channel.QueueDeclare(
		taskQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao declarar fila de tarefas: %v", err)
	}

	// Fila de resultados
	_, err = channel.QueueDeclare(
		resultQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao declarar fila de resultados: %v", err)
	}

	return &LLMRouter{
		conn:        conn,
		channel:     channel,
		inputQueue:  inputQueue,
		taskQueue:   taskQueue,
		resultQueue: resultQueue,
	}, nil
}

// mockLLMBreakdown simula a quebra de tarefas pela LLM
func (r *LLMRouter) mockLLMBreakdown(task TaskRequest) []SubTask {
	// Aqui você integraria com o RouteLLM real
	// Por enquanto, vamos simular a quebra em subtarefas
	subtasks := []SubTask{
		{
			ID:          fmt.Sprintf("%s-1", task.ID),
			ParentID:    task.ID,
			Name:        "Análise de Requisitos",
			Description: "Analisar os requisitos e contexto da tarefa",
			Type:        "analysis",
			Parameters: map[string]interface{}{
				"priority": "high",
				"deadline": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			},
			Status: "pending",
		},
		{
			ID:          fmt.Sprintf("%s-2", task.ID),
			ParentID:    task.ID,
			Name:        "Pesquisa",
			Description: "Realizar pesquisa sobre o tema",
			Type:        "research",
			Parameters: map[string]interface{}{
				"priority": "high",
				"deadline": time.Now().Add(2 * time.Hour).Format(time.RFC3339),
			},
			Status: "pending",
		},
		{
			ID:          fmt.Sprintf("%s-3", task.ID),
			ParentID:    task.ID,
			Name:        "Desenvolvimento",
			Description: "Desenvolver a solução",
			Type:        "development",
			Parameters: map[string]interface{}{
				"priority": "high",
				"deadline": time.Now().Add(3 * time.Hour).Format(time.RFC3339),
			},
			Status: "pending",
		},
		{
			ID:          fmt.Sprintf("%s-4", task.ID),
			ParentID:    task.ID,
			Name:        "Validação",
			Description: "Validar a solução desenvolvida",
			Type:        "validation",
			Parameters: map[string]interface{}{
				"priority": "medium",
				"deadline": time.Now().Add(4 * time.Hour).Format(time.RFC3339),
			},
			Status: "pending",
		},
		{
			ID:          fmt.Sprintf("%s-5", task.ID),
			ParentID:    task.ID,
			Name:        "Documentação",
			Description: "Documentar a solução",
			Type:        "documentation",
			Parameters: map[string]interface{}{
				"priority": "medium",
				"deadline": time.Now().Add(5 * time.Hour).Format(time.RFC3339),
			},
			Status: "pending",
		},
	}

	return subtasks
}

// Start inicia o processamento de tarefas
func (r *LLMRouter) Start(ctx context.Context) error {
	msgs, err := r.channel.Consume(
		r.inputQueue, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("erro ao consumir fila: %v", err)
	}

	log.Printf("🚀 LLMRouter iniciado e aguardando tarefas na fila %s", r.inputQueue)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				var task TaskRequest
				if err := json.Unmarshal(msg.Body, &task); err != nil {
					log.Printf("❌ Erro ao deserializar tarefa: %v", err)
					msg.Nack(false, true)
					continue
				}

				log.Printf("📥 Recebida nova tarefa: %s", task.Description)

				// Quebra a tarefa em subtarefas usando a LLM
				subtasks := r.mockLLMBreakdown(task)
				log.Printf("🔄 Tarefa quebrada em %d subtarefas", len(subtasks))

				// Publica cada subtarefa na fila de tarefas
				for _, subtask := range subtasks {
					taskBytes, err := json.Marshal(subtask)
					if err != nil {
						log.Printf("❌ Erro ao serializar subtarefa: %v", err)
						continue
					}

					err = r.channel.Publish(
						"",          // exchange
						r.taskQueue, // routing key
						false,       // mandatory
						false,       // immediate
						amqp.Publishing{
							ContentType: "application/json",
							Body:        taskBytes,
						})
					if err != nil {
						log.Printf("❌ Erro ao publicar subtarefa: %v", err)
						continue
					}

					log.Printf("📤 Subtarefa publicada: %s", subtask.Name)
				}

				msg.Ack(false)
				log.Printf("✅ Tarefa processada com sucesso")
			}
		}
	}()

	return nil
}

// Close fecha a conexão
func (r *LLMRouter) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("erro ao fechar canal: %v", err)
	}
	return nil
}
