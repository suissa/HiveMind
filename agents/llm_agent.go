package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// LLMAgent representa um agent que processa tarefas do RouteLLM
type LLMAgent struct {
	ID          string
	Type        string
	conn        *amqp.Connection
	channel     *amqp.Channel
	taskQueue   string
	resultQueue string
}

// SubTask representa uma subtarefa a ser executada (mesma estrutura do orchestrator)
type SubTask struct {
	ID          string                 `json:"id"`
	ParentID    string                 `json:"parent_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      string                 `json:"status"`
}

// TaskResult representa o resultado do processamento de uma tarefa
type TaskResult struct {
	TaskID      string                 `json:"task_id"`
	ParentID    string                 `json:"parent_id"`
	AgentID     string                 `json:"agent_id"`
	Status      string                 `json:"status"`
	Result      map[string]interface{} `json:"result"`
	CompletedAt string                 `json:"completed_at"`
}

// NewLLMAgent cria um novo LLMAgent
func NewLLMAgent(id string, agentType string, conn *amqp.Connection) (*LLMAgent, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("erro ao criar canal: %v", err)
	}

	return &LLMAgent{
		ID:          id,
		Type:        agentType,
		conn:        conn,
		channel:     channel,
		taskQueue:   "llm_tasks",
		resultQueue: "llm_results",
	}, nil
}

// processTask simula o processamento de uma tarefa
func (a *LLMAgent) processTask(task SubTask) TaskResult {
	// Simula o tempo de processamento
	processingTime := time.Duration(2+time.Now().Unix()%3) * time.Second
	time.Sleep(processingTime)

	result := TaskResult{
		TaskID:      task.ID,
		ParentID:    task.ParentID,
		AgentID:     a.ID,
		Status:      "completed",
		CompletedAt: time.Now().Format(time.RFC3339),
		Result: map[string]interface{}{
			"processing_time": processingTime.String(),
			"analysis":        fmt.Sprintf("Análise da tarefa '%s' concluída com sucesso", task.Name),
			"details": map[string]interface{}{
				"agent_type": a.Type,
				"task_type":  task.Type,
				"parameters": task.Parameters,
			},
		},
	}

	return result
}

// Start inicia o processamento de tarefas
func (a *LLMAgent) Start(ctx context.Context) error {
	msgs, err := a.channel.Consume(
		a.taskQueue, // queue
		a.ID,        // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("erro ao consumir fila: %v", err)
	}

	log.Printf("🤖 Agent %s (%s) iniciado e aguardando tarefas...", a.ID, a.Type)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				var task SubTask
				if err := json.Unmarshal(msg.Body, &task); err != nil {
					log.Printf("❌ Agent %s: Erro ao deserializar tarefa: %v", a.ID, err)
					msg.Nack(false, true)
					continue
				}

				// Verifica se o agent pode processar este tipo de tarefa
				if task.Type != a.Type {
					msg.Nack(false, true) // Rejeita e recoloca na fila
					continue
				}

				log.Printf("🔄 Agent %s: Processando tarefa %s", a.ID, task.Name)

				// Processa a tarefa
				result := a.processTask(task)

				// Publica o resultado
				resultBytes, err := json.Marshal(result)
				if err != nil {
					log.Printf("❌ Agent %s: Erro ao serializar resultado: %v", a.ID, err)
					msg.Nack(false, true)
					continue
				}

				err = a.channel.Publish(
					"",            // exchange
					a.resultQueue, // routing key
					false,         // mandatory
					false,         // immediate
					amqp.Publishing{
						ContentType: "application/json",
						Body:        resultBytes,
					})
				if err != nil {
					log.Printf("❌ Agent %s: Erro ao publicar resultado: %v", a.ID, err)
					msg.Nack(false, true)
					continue
				}

				msg.Ack(false)
				log.Printf("✅ Agent %s: Tarefa %s concluída", a.ID, task.Name)
			}
		}
	}()

	return nil
}

// Close fecha a conexão do agent
func (a *LLMAgent) Close() error {
	if err := a.channel.Close(); err != nil {
		return fmt.Errorf("erro ao fechar canal: %v", err)
	}
	return nil
}
