package main

import (
	"fmt"
	"log"
	"time"

	"HiveMind/agents"
	"HiveMind/agents/telemetry"

	"github.com/google/uuid"
)

func main() {
	// Inicializar telemetria
	err := telemetry.InitTelemetry("cognitive-agents")
	if err != nil {
		log.Fatalf("Erro ao inicializar telemetria: %v", err)
	}

	// Criar o Orchestrator
	orchestrator, err := agents.NewOrchestratorInfrastructureAgent()
	if err != nil {
		log.Fatalf("Erro ao criar orchestrator: %v", err)
	}
	defer orchestrator.Stop()

	taskManager := orchestrator.GetTaskManager()

	// Criar agentes cognitivos
	quizAgent := agents.NewQuizAgent(taskManager)
	challengeAgent := agents.NewChallengeAgent(taskManager)

	// Registrar agentes no orchestrator
	orchestrator.RegisterAgent(quizAgent)
	orchestrator.RegisterAgent(challengeAgent)

	// Adicionar algumas tarefas
	quizTasks := []struct {
		priority agents.TaskPriority
		data     interface{}
	}{
		{agents.PriorityHigh, map[string]interface{}{
			"subject": "Matemática",
			"topic":   "Álgebra Linear",
			"level":   "Intermediário",
		}},
		{agents.PriorityNormal, map[string]interface{}{
			"subject": "História",
			"topic":   "Segunda Guerra Mundial",
			"level":   "Básico",
		}},
		{agents.PriorityLow, map[string]interface{}{
			"subject": "Biologia",
			"topic":   "Genética",
			"level":   "Avançado",
		}},
	}

	challengeTasks := []struct {
		priority agents.TaskPriority
		data     interface{}
	}{
		{agents.PriorityHigh, map[string]interface{}{
			"subject":    "Programação",
			"topic":      "Algoritmos",
			"difficulty": "Médio",
		}},
		{agents.PriorityNormal, map[string]interface{}{
			"subject":    "Física",
			"topic":      "Mecânica",
			"difficulty": "Fácil",
		}},
	}

	// Adicionar tarefas de quiz
	for _, t := range quizTasks {
		task := &agents.Task{
			ID:       uuid.New().String(),
			Type:     "quiz",
			Priority: t.priority,
			Data:     t.data,
		}
		orchestrator.AddTask(task)
	}

	// Adicionar tarefas de desafio
	for _, t := range challengeTasks {
		task := &agents.Task{
			ID:       uuid.New().String(),
			Type:     "challenge",
			Priority: t.priority,
			Data:     t.data,
		}
		orchestrator.AddTask(task)
	}

	// Monitorar progresso
	for i := 0; i < 60; i++ {
		// Verificar saúde dos agentes
		quizHealth := taskManager.GetAgentHealth(quizAgent.Name)
		if quizHealth != nil {
			fmt.Printf("\nSaúde do Quiz Agent:\n")
			fmt.Printf("- Processando: %v\n", quizHealth.IsProcessing)
			fmt.Printf("- Tarefa Atual: %s\n", quizHealth.CurrentTaskID)
			fmt.Printf("- Tempo Médio: %.2fs\n", quizHealth.ProcessingTime)
			fmt.Printf("- Taxa de Sucesso: %.2f%%\n", quizHealth.SuccessRate*100)
		}

		challengeHealth := taskManager.GetAgentHealth(challengeAgent.Name)
		if challengeHealth != nil {
			fmt.Printf("\nSaúde do Challenge Agent:\n")
			fmt.Printf("- Processando: %v\n", challengeHealth.IsProcessing)
			fmt.Printf("- Tarefa Atual: %s\n", challengeHealth.CurrentTaskID)
			fmt.Printf("- Tempo Médio: %.2fs\n", challengeHealth.ProcessingTime)
			fmt.Printf("- Taxa de Sucesso: %.2f%%\n", challengeHealth.SuccessRate*100)
		}

		// Verificar escalabilidade
		orchestrator.ScaleSystem()

		time.Sleep(1 * time.Second)
	}
}
