package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"HiveMind/agents"
)

func main() {
	// Cria configuração de treinamento
	config := agents.TrainingConfig{
		MaxRounds:       5,
		TrainingTimeout: 2 * time.Second,
		ValidationRatio: 0.2,
		MinAccuracy:     0.8,
		BatchSize:       32,
		LearningRate:    0.001,
		UseHistorical:   true,
	}

	// Cria o gerenciador de treinamento
	trainer := agents.NewAgentTrainer(config)

	// Cria alguns agentes cognitivos para teste
	agent1 := agents.NewCognitiveAgent(
		"agent1",
		"Assistente de Pesquisa",
		"Agente especializado em pesquisa e análise de dados",
		3,
		"gpt-4",
		"researcher",
		"Realizar pesquisas aprofundadas e análises de dados",
	)
	agent1.SetBackstory("Sou um assistente de pesquisa com experiência em análise de dados e machine learning")

	agent2 := agents.NewCognitiveAgent(
		"agent2",
		"Engenheiro de Software",
		"Agente especializado em desenvolvimento de software",
		5,
		"gpt-4",
		"engineer",
		"Desenvolver e otimizar código",
	)
	agent2.SetBackstory("Sou um engenheiro de software com foco em arquitetura e boas práticas")

	agent3 := agents.NewCognitiveAgent(
		"agent3",
		"Gerente de Projeto",
		"Agente especializado em gerenciamento de projetos",
		4,
		"gpt-4",
		"manager",
		"Coordenar equipes e garantir entrega de projetos",
	)
	agent3.SetBackstory("Sou um gerente de projetos com experiência em metodologias ágeis")

	// Adiciona templates de prompts para cada agente
	agent1.AddPromptTemplate("research", "Realize uma pesquisa sobre {topic} considerando {aspects}")
	agent2.AddPromptTemplate("code_review", "Analise o código {code} e sugira melhorias")
	agent3.AddPromptTemplate("task_delegation", "Delegue a tarefa {task} para o membro mais adequado da equipe")

	// Adiciona conhecimento base para os agentes
	agent1.AddToKnowledgeBase("research_methods", []string{"qualitative", "quantitative", "mixed"})
	agent2.AddToKnowledgeBase("programming_languages", []string{"Go", "Python", "JavaScript"})
	agent3.AddToKnowledgeBase("methodologies", []string{"Scrum", "Kanban", "XP"})

	// Adiciona os agentes ao trainer
	trainer.AddAgent(agent1)
	trainer.AddAgent(agent2)
	trainer.AddAgent(agent3)

	// Cria contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Executa o treinamento
	fmt.Println("Iniciando treinamento dos agentes...")
	if err := trainer.Train(ctx); err != nil {
		log.Fatalf("Erro durante treinamento: %v", err)
	}

	// Exibe métricas de cada agente
	fmt.Println("\nMétricas de treinamento:")
	metrics := trainer.GetAllMetrics()
	for agent, metric := range metrics {
		var cogAgent *agents.CognitiveAgent
		switch a := agent.(type) {
		case *agents.CognitiveAgent:
			cogAgent = a
		}

		if cogAgent != nil {
			fmt.Printf("\nAgente: %s (%s)\n", cogAgent.Name, cogAgent.GetRole())
			fmt.Printf("Objetivo: %s\n", cogAgent.GetGoal())
			fmt.Printf("Rounds executados: %d/%d\n", metric.RoundsExecuted, cogAgent.GetMaxRounds())
			fmt.Printf("Tempo de treinamento: %v\n", metric.EndTime.Sub(metric.StartTime))
			fmt.Printf("Acurácia: %.2f\n", metric.Accuracy)
			fmt.Printf("Loss: %.2f\n", metric.Loss)

			// Exibe estatísticas de performance
			stats := cogAgent.GetPerformanceStats()
			fmt.Println("\nEstatísticas de performance:")
			for key, value := range stats {
				fmt.Printf("- %s: %.2f\n", key, value)
			}

			if len(metric.Errors) > 0 {
				fmt.Printf("\nErros: %v\n", metric.Errors)
			} else {
				fmt.Println("\nSem erros")
			}
		}
	}

	// Tenta salvar o estado dos agentes
	fmt.Println("\nSalvando estado dos agentes...")
	for agent := range metrics {
		var cogAgent *agents.CognitiveAgent
		switch a := agent.(type) {
		case *agents.CognitiveAgent:
			cogAgent = a
		}

		if cogAgent != nil {
			path := fmt.Sprintf("agent_%s_state.json", cogAgent.ID)
			if err := cogAgent.SaveState(path); err != nil {
				fmt.Printf("Erro ao salvar estado do agente %s: %v\n", cogAgent.Name, err)
			} else {
				fmt.Printf("Estado do agente %s salvo em %s\n", cogAgent.Name, path)
			}
		}
	}
}
