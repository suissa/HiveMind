package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"HiveMind/agents"
	"HiveMind/agents/memory"
)

func main() {
	ctx := context.Background()

	// Obtém o diretório base
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Erro ao obter diretório atual: %v", err)
	}

	// Carrega configurações
	agentsConfig, err := agents.LoadAgentsConfig(filepath.Join(baseDir, "config", "agents.yaml"))
	if err != nil {
		log.Fatalf("Erro ao carregar configuração dos agentes: %v", err)
	}

	tasksConfig, err := agents.LoadTasksConfig(filepath.Join(baseDir, "config", "tasks.yaml"))
	if err != nil {
		log.Fatalf("Erro ao carregar configuração das tarefas: %v", err)
	}

	toolsConfig, err := agents.LoadToolsConfig(filepath.Join(baseDir, "config", "tools.yaml"))
	if err != nil {
		log.Fatalf("Erro ao carregar configuração das ferramentas: %v", err)
	}

	// Configuração do gerenciador de memória
	memConfig := &memory.MemoryConfig{
		RedisURL:            "redis://localhost:4567", // Usando a porta 4567 para o Redis
		MongoURL:            "mongodb://suissa:dc0b410b23dd26da2d423375437cceb4@195.35.19.148:27017/",
		MongoDB:             "HiveMind",
		Collection:          "memories",
		ShortTermTTL:        1 * time.Hour,
		ImportanceThreshold: 0.7,
		WeaviateURL:         "195.35.19.148:1111",
		WeaviateClass:       "Memory",
		WeaviateBatchSize:   100,
	}

	memManager, err := memory.NewHybridMemoryManager(ctx, memConfig)
	if err != nil {
		log.Fatalf("Erro ao criar gerenciador de memória: %v", err)
	}
	defer memManager.Close(ctx)

	// Cria uma equipe de marketing
	crew := agents.NewMarketingCrew(memManager)

	// Registra listener para todos os eventos
	crew.OnAnyEvent(func(event agents.Event) {
		log.Printf("\n🔔 === NOVO EVENTO [%s] ===\n%s\n========================", event.Type, event.ToJSON())
	})

	// Registra listeners específicos para cada tipo de evento
	crew.OnEvent(agents.EventAgentAction, func(event agents.Event) {
		if event.Data["action"] == "add_agent" {
			log.Printf("\n👤 AGENTE: Novo agente adicionado")
			log.Printf("Nome: %s", event.Data["agent_name"])
			log.Printf("Função: %s", event.Data["agent_role"])
			log.Printf("ID: %s", event.Data["agent_id"])
		}
	})

	crew.OnEvent(agents.EventTaskUpdate, func(event agents.Event) {
		action := event.Data["action"].(string)
		taskName := event.Data["task_name"].(string)
		assignedTo := event.Data["assigned_to"].(string)

		switch action {
		case "task_start":
			log.Printf("\n▶️ TAREFA: Iniciando")
			log.Printf("Nome: %s", taskName)
			log.Printf("Responsável: %s", assignedTo)
			if deadline, ok := event.Data["deadline"]; ok {
				log.Printf("Prazo: %s", deadline)
			}
		case "task_complete":
			log.Printf("\n✅ TAREFA: Concluída")
			log.Printf("Nome: %s", taskName)
			log.Printf("Responsável: %s", assignedTo)
			if duration, ok := event.Data["duration"]; ok {
				log.Printf("Duração: %s", duration)
			}
		}
	})

	crew.OnEvent(agents.EventWorkflowUpdate, func(event agents.Event) {
		action := event.Data["action"].(string)
		switch action {
		case "workflow_start":
			log.Printf("\n🚀 WORKFLOW: Iniciando")
			log.Printf("Projeto: %s", event.Data["project"])
			log.Printf("Objetivo: %s", event.Data["objective"])
			log.Printf("Timestamp: %s", event.Timestamp.Format(time.RFC3339))
		case "workflow_complete":
			log.Printf("\n🏁 WORKFLOW: Concluído")
			log.Printf("Duração: %s", event.Data["duration"])
			if results, ok := event.Data["results"].(map[string]interface{}); ok {
				log.Printf("\nResultados:")
				for k, v := range results {
					log.Printf("- %s: %v", k, v)
				}
			}
		}
	})

	crew.OnEvent(agents.EventProjectUpdate, func(event agents.Event) {
		if event.Data["action"] == "status_update" {
			log.Printf("\n📊 PROJETO: Atualização de Status")
			log.Printf("Progresso: %.2f%%", event.Data["progress"])
			log.Printf("Tarefas: %d/%d", event.Data["completed_tasks"], event.Data["total_tasks"])
			log.Printf("Tempo decorrido: %s", event.Data["elapsed_time"])
			log.Printf("Tempo restante: %s", event.Data["remaining_time"])
			if metrics, ok := event.Data["metrics"].(map[string]interface{}); ok {
				log.Printf("\nMétricas:")
				for k, v := range metrics {
					log.Printf("- %s: %v", k, v)
				}
			}
		}
	})

	crew.OnEvent(agents.EventMemoryOperation, func(event agents.Event) {
		action := event.Data["action"].(string)
		switch action {
		case "store":
			log.Printf("\n💾 MEMÓRIA: Armazenamento")
			log.Printf("Conteúdo: %s", event.Data["content"])
			if tags, ok := event.Data["tags"].([]string); ok {
				log.Printf("Tags: %v", tags)
			}
			if importance, ok := event.Data["importance"].(float64); ok {
				log.Printf("Importância: %.2f", importance)
			}
		case "retrieve":
			log.Printf("\n📖 MEMÓRIA: Recuperação")
			log.Printf("Conteúdo: %s", event.Data["content"])
			if memoryId, ok := event.Data["memory_id"].(string); ok {
				log.Printf("ID: %s", memoryId)
			}
		case "search":
			log.Printf("\n🔍 MEMÓRIA: Busca")
			log.Printf("Query: %s", event.Data["query"])
			if results, ok := event.Data["results"].([]interface{}); ok {
				log.Printf("Encontradas %d memórias similares", len(results))
				for i, result := range results {
					if memory, ok := result.(map[string]interface{}); ok {
						log.Printf("\nMemória #%d:", i+1)
						log.Printf("- Conteúdo: %s", memory["content"])
						log.Printf("- Similaridade: %.2f", memory["similarity"])
						if tags, ok := memory["tags"].([]string); ok {
							log.Printf("- Tags: %v", tags)
						}
					}
				}
			}
		}
	})

	// Configura os agentes
	for _, agentConfig := range agentsConfig.Agents {
		log.Printf("Configurando agente: %s (%s)", agentConfig.Name, agentConfig.Role)
		log.Printf("Objetivo: %s", agentConfig.Goal)
		log.Printf("Backstory: %s\n", agentConfig.Backstory)

		agent := agents.NewCognitiveAgent(
			agentConfig.ID,
			agentConfig.Name,
			agentConfig.Description,
			agentConfig.MaxRounds,
			agentConfig.Model,
			agentConfig.Role,
			agentConfig.Goal,
			memManager,
		)

		agent.SetBackstory(agentConfig.Backstory)
		crew.AddAgent(agent)
	}

	// Define detalhes do projeto
	projectDetails := &agents.MarketingProject{
		Name:      "Campanha de Lançamento de Produto",
		Objective: "Lançar novo produto no mercado com foco em sustentabilidade",
		TargetAudience: []string{
			"Consumidores conscientes",
			"Faixa etária 25-45 anos",
			"Classe A/B",
		},
		Budget:   100000.0,
		Duration: 90 * 24 * time.Hour,
		Channels: []string{"Social Media", "Content Marketing", "Influencer Marketing"},
		Constraints: []string{
			"Orçamento limitado",
			"Timeline agressivo",
			"Foco em sustentabilidade",
		},
	}

	// Mapeia tarefas do projeto
	for _, taskConfig := range tasksConfig.Tasks {
		projectDetails.AddTask(taskConfig)
	}

	// Registra ferramentas disponíveis por categoria
	log.Println("\nFerramentas disponíveis:")
	for category, tools := range toolsConfig.Tools {
		log.Printf("\nCategoria: %s", category)
		for _, tool := range tools {
			log.Printf("- %s: %s", tool.Name, tool.Description)
		}
	}

	// Executa o workflow
	results, err := crew.ExecuteWorkflow(projectDetails)
	if err != nil {
		log.Printf("Erro na execução do workflow: %v", err)
	} else {
		log.Printf("\nResultados do workflow:")
		log.Printf("Estratégia: %s", results.Strategy)
		log.Printf("Campanha: %s", results.Campaign)
		log.Printf("Copy: %s", results.Copy)
	}

	// Configura handler para sinais de término
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Monitora o projeto
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := crew.GetProjectStatus()
			log.Printf("\nStatus do projeto:")
			log.Printf("Progresso: %.2f%%", status.Progress)
			log.Printf("Tarefas completadas: %d/%d", status.CompletedTasks, status.TotalTasks)

		case sig := <-sigChan:
			log.Printf("Recebido sinal %v, encerrando...", sig)
			return
		}
	}
}
