package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"

	"github.com/suissa/HiveMind/agents"
	"github.com/suissa/HiveMind/config"
	"github.com/suissa/HiveMind/orchestrator"
)

// Tipos de agents disponíveis
var agentTypes = []struct {
	Type        string
	Description string
}{
	{
		Type:        "analysis",
		Description: "Análise de requisitos e contexto",
	},
	{
		Type:        "research",
		Description: "Pesquisa e coleta de informações",
	},
	{
		Type:        "development",
		Description: "Desenvolvimento da solução",
	},
	{
		Type:        "validation",
		Description: "Validação e testes",
	},
	{
		Type:        "documentation",
		Description: "Documentação e relatórios",
	},
}

func main() {
	// Carrega as variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ Arquivo .env não encontrado, usando valores padrão")
	}

	// Configuração do RabbitMQ
	rabbitConfig := config.NewRabbitMQConfig()
	conn, err := config.ConnectRabbitMQ(rabbitConfig)
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Criando o contexto principal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Criando e iniciando o LLMRouter
	router, err := orchestrator.NewLLMRouter(conn)
	if err != nil {
		log.Fatalf("❌ Erro ao criar LLMRouter: %v", err)
	}
	defer router.Close()

	if err := router.Start(ctx); err != nil {
		log.Fatalf("❌ Erro ao iniciar LLMRouter: %v", err)
	}

	// Criando e iniciando os agents
	var wg sync.WaitGroup
	for i, agentType := range agentTypes {
		// Criando múltiplas instâncias de cada tipo de agent
		for j := 1; j <= 2; j++ { // 2 agents de cada tipo = 10 agents no total
			agentID := fmt.Sprintf("llm-agent-%d-%d", i+1, j)
			agent, err := agents.NewLLMAgent(agentID, agentType.Type, conn)
			if err != nil {
				log.Printf("❌ Erro ao criar agent %s: %v", agentID, err)
				continue
			}
			defer agent.Close()

			wg.Add(1)
			go func(a *agents.LLMAgent, typ string, desc string) {
				defer wg.Done()
				log.Printf("🤖 Iniciando %s (Tipo: %s - %s)", a.ID, typ, desc)
				if err := a.Start(ctx); err != nil {
					log.Printf("❌ Erro ao iniciar agent %s: %v", a.ID, err)
				}
			}(agent, agentType.Type, agentType.Description)
		}
	}

	// Enviando uma tarefa de exemplo
	task := orchestrator.TaskRequest{
		ID:          uuid.New().String(),
		Description: "Analisar o repositório RouteLLM (https://github.com/lm-sys/RouteLLM)",
		Parameters: map[string]interface{}{
			"repository": "https://github.com/lm-sys/RouteLLM",
			"priority":   "high",
			"context":    "Análise técnica e funcional do projeto",
		},
	}

	// Publicando a tarefa
	taskBytes, err := json.Marshal(task)
	if err != nil {
		log.Fatalf("❌ Erro ao serializar tarefa: %v", err)
	}

	err = router.channel.Publish(
		"",                // exchange
		router.inputQueue, // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskBytes,
		})
	if err != nil {
		log.Fatalf("❌ Erro ao publicar tarefa: %v", err)
	}

	log.Printf("🚀 Sistema iniciado com %d tipos de agents (total de %d agents)",
		len(agentTypes), len(agentTypes)*2)
	log.Printf("📤 Tarefa enviada: %s", task.Description)

	// Aguardando sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Encerrando graciosamente
	log.Println("👋 Encerrando o sistema...")
	cancel()
	wg.Wait()
}
